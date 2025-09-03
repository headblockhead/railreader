package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

type SFTPCommand struct {
	Address            string `env:"SFTP_ADDRESS" help:"Address (v4 or v6) to listen on for file transfers from the Rail Data Marketplace." default:"127.0.0.1:2022"`
	HashedPasswordFile []byte `env:"SFTP_HASHED_PASSWORD_FILE" help:"File containing a bcrypt hashed password to authenticate incoming SFTP connections" type:"filecontent" required:""`
	PrivateHostKeyFile []byte `env:"SFTP_PRIVATE_HOST_KEY_FILE" help:"File containing the SFTP server's private key" type:"filecontent" required:""`
	Logging            struct {
		Level  string `enum:"debug,info,warn,error" env:"LOG_LEVEL" default:"warn"`
		Format string `enum:"json,console" env:"LOG_FORMAT" default:"json"`
	} `embed:"" prefix:"log."`
}

func (c *SFTPCommand) Run() error {
	log := getLogger(c.Logging.Level, c.Logging.Format == "json")

	//listenerContext, listenerCancel := context.WithCancel(context.Background())
	//go cancelOnSignal(listenerCancel, log)

	config := &ssh.ServerConfig{
		MaxAuthTries: 1,
	}
	privateKey, err := ssh.ParsePrivateKey(c.PrivateHostKeyFile)
	if err != nil {
		return fmt.Errorf("failed to parse private host key: %w", err)
	}
	config.AddHostKey(privateKey)

	listener, err := net.Listen("tcp", c.Address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", c.Address, err)
	}
	log.Info("listening on", slog.String("address", c.Address))
	for {
		// Block until new connection to the server
		connection, err := listener.Accept()
		if err != nil {
			log.Error("error accepting an incoming connection", slog.Any("error", err))
			continue
		}
		connectionGroup := slog.GroupAttrs("connection", slog.String("localAddress", connection.LocalAddr().String()), slog.String("remoteAddress", connection.RemoteAddr().String()))
		connectionLog := log.With(connectionGroup)
		connectionLog.Debug("recieved new connection")

		config.PasswordCallback = func(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if conn.User() != "darwin" {
				return nil, errors.New("incorrect username")
			}
			// bcrypt.CompareHashAndPassword is a constant time comparison
			if err := bcrypt.CompareHashAndPassword(c.HashedPasswordFile, pass); err != nil {
				return nil, fmt.Errorf("password rejected for %s", conn.User())
			}
			return nil, nil
		}
		config.AuthLogCallback = func(conn ssh.ConnMetadata, method string, err error) {
			sshConnectionGroup := slog.GroupAttrs("ssh", slog.String("username", conn.User()), slog.Any("sessionID", conn.SessionID()))
			attemptGroup := slog.GroupAttrs("attempt", slog.String("method", method))
			attemptLog := connectionLog.With(sshConnectionGroup, attemptGroup)
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
		}

		sshConnection, channels, reqs, err := ssh.NewServerConn(connection, config)
		if err != nil {
			log.Error("error performing SSH handshake", slog.Any("error", err))
			continue
		}
		go ssh.DiscardRequests(reqs)
		sshConnectionGroup := slog.GroupAttrs("ssh", slog.String("username", sshConnection.User()), slog.Any("sessionID", sshConnection.SessionID()))
		sshLog := connectionLog.With(sshConnectionGroup)
		sshLog.Debug("completed SSH handshake")

		for channel := range channels {
			channelGroup := slog.GroupAttrs("channel", slog.String("type", channel.ChannelType()))
			channelLog := connectionLog.With(channelGroup)
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
					requestLog.Debug("recieved request", slog.String("type", request.Type), slog.Bool("wantReply", request.WantReply))
					if request.Type != "subsystem" {
						requestLog.Warn("rejected request of unhandled type (type != 'subsystem')")
						continue
					}
					// TODO: check SFTP
					requestLog.Debug("request OK")
					request.Reply(true, nil)
				}
			}()

			server, err := sftp.NewServer(channel)
			if err != nil {
				channelLog.Error("error while initialising STFP session", slog.Any("error", err))
				continue
			}
			if err := server.Serve(); err != nil {
				if err != io.EOF {
					channelLog.Error("error while running SFTP session", slog.Any("error", err))
				} else {
					channelLog.Info("SFTP server session completed successfully")
				}
			}
			server.Close()
		}
	}
}
