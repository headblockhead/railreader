package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

type SFTPCommand struct {
	Addresses          []string `env:"SFTP_ADDRESSES" help:"A list of addresses to listen to." default:"127.0.0.1:822"`
	HashedPassword     string   `env:"SFTP_HASHED_PASSWORD" help:"A bcrypt hashed password to authenticate incoming SFTP connections." required:""`
	PrivateHostKeyFile []byte   `env:"SFTP_PRIVATE_HOST_KEY_FILE" help:"File containing the SFTP server's SSH private key." type:"filecontent" required:""`
	DarwinDirectory    string   `env:"SFTP_DARWIN_DIRECTORY" help:"Directory to store Darwin's SFTP data in. The ingest command must have access to this directory." default:"./darwin" type:"existingdir" required:""`
	Logging            struct {
		Level  string `enum:"debug,info,warn,error" env:"LOG_LEVEL" default:"warn"`
		Format string `enum:"json,console" env:"LOG_FORMAT" default:"json"`
	} `embed:"" prefix:"log."`
}

func (c *SFTPCommand) Run() error {
	log := getLogger(c.Logging.Level, c.Logging.Format == "json")

	privateKey, err := ssh.ParsePrivateKey(c.PrivateHostKeyFile)
	if err != nil {
		return fmt.Errorf("failed to parse private host key: %w", err)
	}

	var listenerGroup sync.WaitGroup
	var listeners []net.Listener
	for _, address := range c.Addresses {
		listener, err := net.Listen("tcp", address)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %w", address, err)
		}
		log.Info("listening on", slog.String("address", address))
		listeners = append(listeners, listener)
	}

	var handlerGroup sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	go onSignal(log, func() {
		cancel()
		for _, listener := range listeners {
			listener.Close()
		}
	})

	for _, listener := range listeners {
		listenerGroup.Go(func() {
			for {
				// Block until there is a new connection to the server or the listener is closed.
				connection, err := listener.Accept()
				if err != nil {
					// If the context has been cancelled, don't show an error message, the error is intentional.
					if ctx.Err() == nil {
						log.Error("error accepting an incoming connection", slog.Any("error", err))
					}
					break
				}
				connectionGroup := slog.GroupAttrs("connection", slog.String("localAddress", connection.LocalAddr().String()), slog.String("remoteAddress", connection.RemoteAddr().String()))
				connectionLog := log.With(connectionGroup)
				go c.handleConnection(&handlerGroup, connectionLog, connection, privateKey)
			}
		})
	}

	listenerGroup.Wait()
	handlerGroup.Wait()

	return nil
}

func (c *SFTPCommand) handleConnection(handlerGroup *sync.WaitGroup, log *slog.Logger, connection net.Conn, privateKey ssh.Signer) {
	log.Debug("recieved new connection")

	bytesOfHashedPassword := []byte(c.HashedPassword)
	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if conn.User() != "darwin" {
				return nil, errors.New("incorrect username")
			}
			// bcrypt.CompareHashAndPassword is a constant time comparison
			if err := bcrypt.CompareHashAndPassword(bytesOfHashedPassword, pass); err != nil {
				return nil, fmt.Errorf("password rejected for %s", conn.User())
			}
			return nil, nil
		},
		AuthLogCallback: func(conn ssh.ConnMetadata, method string, err error) {
			sshConnectionGroup := slog.GroupAttrs("ssh", slog.String("username", conn.User()))
			attemptGroup := slog.GroupAttrs("attempt", slog.String("method", method))
			attemptLog := log.With(sshConnectionGroup, attemptGroup)
			if err != nil {
				if err == ssh.ErrNoAuth {
					attemptLog.Debug("authentication attempt started")
					return
				}
				attemptLog.Info("unsuccessful authentication attempt", slog.Any("error", err))
				return
			}
			if method == "password" {
				attemptLog.Info("successful authentication attempt")
			}
		},
		MaxAuthTries: 1,
	}
	config.AddHostKey(privateKey)

	sshConnection, channels, reqs, err := ssh.NewServerConn(connection, config)
	if err != nil {
		log.Error("error performing SSH handshake", slog.Any("error", err))
		return
	}
	go ssh.DiscardRequests(reqs)
	sshConnectionGroup := slog.GroupAttrs("ssh", slog.String("username", sshConnection.User()))
	sshLog := log.With(sshConnectionGroup)
	sshLog.Debug("completed SSH handshake")

	handlerGroup.Go(func() {
		c.handleSSHConnection(sshLog, channels)
		connection.Close()
	})
}

func (c *SFTPCommand) handleSSHConnection(log *slog.Logger, channels <-chan ssh.NewChannel) {
	for channel := range channels {
		channelGroup := slog.GroupAttrs("channel", slog.String("type", channel.ChannelType()))
		channelLog := log.With(channelGroup)
		channelLog.Debug("handling channel creation request")
		// Handle only "session" channels
		if channel.ChannelType() != "session" {
			channel.Reject(ssh.UnknownChannelType, "unknown channel type")
			log.Warn("rejected request to create channel of unhandled type (type != 'session')")
			continue
		}
		channel, requests, err := channel.Accept()
		if err != nil {
			log.Error("error accepting channel creation request", slog.Any("error", err))
		}
		log.Debug("accepted channel creation request")

		go func() {
			for request := range requests {
				requestGroup := slog.GroupAttrs("request", slog.String("type", request.Type), slog.Bool("wantReply", request.WantReply))
				requestLog := channelLog.With(requestGroup)
				requestLog.Debug("recieved request")
				if request.Type != "subsystem" {
					requestLog.Warn("rejected request of unhandled type (type != 'subsystem')")
					request.Reply(false, nil)
					continue
				}
				if len(request.Payload) > 4 {
					if string(request.Payload[4:]) != "sftp" {
						requestLog.Warn("rejected non-SFTP subsystem request")
						request.Reply(false, nil)
						continue
					}
				} else {
					requestLog.Warn("rejected invalid length subsystem request")
					request.Reply(false, nil)
					continue
				}
				requestLog.Debug("request OK")
				request.Reply(true, nil)
			}
		}()

		server, err := sftp.NewServer(channel, sftp.WithServerWorkingDirectory(c.DarwinDirectory), sftp.WithDebug(os.Stdout))
		if err != nil {
			channelLog.Error("error while initialising sftp session", slog.Any("error", err))
			continue
		}
		if err := server.Serve(); err != nil {
			if err != io.EOF {
				channelLog.Error("error while running sftp session", slog.Any("error", err))
			} else {
				channelLog.Info("sftp server session completed successfully")
			}
		}
		if err := server.Close(); err != nil {
			channelLog.Error("error while closing sftp session", slog.Any("error", err))
		}
	}
}
