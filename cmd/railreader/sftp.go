package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	"github.com/headblockhead/sftp"
	"github.com/headblockhead/sftp/localfs"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

type SFTPCommand struct {
	ListenAddress    string `env:"SFTP_LISTEN_ADDRESS" help:"Address to listen for SFTP on." default:"127.0.0.1:822"`
	HashedPassword   string `env:"SFTP_HASHED_PASSWORD" help:"A bcrypt hashed password to authenticate incoming connections." required:""`
	WorkingDirectory string `env:"SFTP_WORKING_DIRECTORY" help:"An existing directory to write incoming files into." type:"existingdir" required:""`

	Logging struct {
		Level  string `enum:"debug,info,warn,error" env:"LOG_LEVEL" default:"warn"`
		Format string `enum:"json,console" env:"LOG_FORMAT" default:"console"`
	} `embed:"" prefix:"log."`

	log  *slog.Logger `kong:"-"`
	root *os.Root     `kong:"-"`
}

func (c *SFTPCommand) sshPasswordCallback(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	if conn.User() != "darwin" {
		return nil, errors.New("invalid username")
	}
	// bcrypt.CompareHashAndPassword is a constant time comparison
	if err := bcrypt.CompareHashAndPassword([]byte(c.HashedPassword), pass); err != nil {
		return nil, fmt.Errorf("password rejected for %s", conn.User())
	}
	return nil, nil
}

func (c *SFTPCommand) sshAuthLogCallback(conn ssh.ConnMetadata, method string, err error) {
	sessionLog := c.log.With(slog.String("sessionID", hex.EncodeToString(conn.SessionID())), slog.String("username", conn.User()), slog.String("remoteAddress", conn.RemoteAddr().String()))
	sessionLog.Debug("authentication attempt", slog.String("method", method))
	if err != nil {
		if err == ssh.ErrNoAuth {
			return
		}
		sessionLog.Warn("failed authentication attempt", slog.Any("error", err))
		return
	}
	if method == "password" {
		sessionLog.Info("successful authentication attempt")
	}
}

func (c *SFTPCommand) Run() error {
	c.log = getLogger(c.Logging.Level, c.Logging.Format == "json")

	// Open the working directory.
	root, err := os.OpenRoot(c.WorkingDirectory)
	if err != nil {
		return fmt.Errorf("failed to open root directory: %w", err)
	}
	defer root.Close()
	c.root = root

	// Setup directory structure for each 'user' (service).
	if err := c.root.Mkdir("darwin", 0755); err != nil {
		return fmt.Errorf("failed to create darwin directory: %w", err)
	}
	if err := c.root.Mkdir("darwin/EHSnapshot", 0755); err != nil {
		return fmt.Errorf("failed to create darwin/EHSnapshot directory: %w", err)
	}
	if err := c.root.Mkdir("darwin/PPTimetable", 0755); err != nil {
		return fmt.Errorf("failed to create darwin/PPTimetable directory: %w", err)
	}

	// Create a new SSH server and generate a new set of random host keys.
	config := &ssh.ServerConfig{
		PasswordCallback: c.sshPasswordCallback,
		AuthLogCallback:  c.sshAuthLogCallback,
		MaxAuthTries:     1,
	}
	privateKeyBits, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	privateKeyDER := x509.MarshalPKCS1PrivateKey(privateKeyBits)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyDER,
	})
	privateKey, err := ssh.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		return err
	}
	config.AddHostKey(privateKey)
	c.log.Debug("generated new SSH host key", slog.String("fingerprint", ssh.FingerprintSHA256(privateKey.PublicKey())))

	// Start listening for incoming connections.
	var listenerGroup sync.WaitGroup
	listener, err := net.Listen("tcp", c.ListenAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", c.ListenAddress, err)
	}
	c.log.Info("listening on", slog.String("address", c.ListenAddress))

	// Handle what to do on exit signals.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go onSignal(c.log, func() {
		cancel()
		listener.Close()
		c.log.Debug("no longer listening on", slog.String("address", listener.Addr().String()))
	})

	var handlerGroup sync.WaitGroup

	listenerGroup.Go(func() {
		for {
			// Block until there is a new netConnection to the server or the listener is closed.
			netConnection, err := listener.Accept()
			if err != nil {
				// If the listeners have not been closed, log the error.
				if ctx.Err() == nil {
					c.log.Error("error accepting an incoming connection", slog.Any("error", err))
				}
				break
			}
			// Handle the new connection in a new goroutine to be able to accept more connections as fast as possible.
			go c.handleConnection(&handlerGroup, netConnection, config)
		}
	})

	listenerGroup.Wait() // finish listening for new connections
	c.log.Info("waiting for existing sftp sessions to end")
	handlerGroup.Wait() // finish handling existing authenticated connections
	c.log.Info("sftp server stopped")

	return nil
}

func (c *SFTPCommand) handleConnection(handlerGroup *sync.WaitGroup, netConnection net.Conn, config *ssh.ServerConfig) {
	c.log.Info("recieved new connection", slog.String("remoteAddress", netConnection.RemoteAddr().String()))

	netConnection.SetDeadline(time.Now().Add(2 * time.Minute))
	sshConnection, channels, reqs, err := ssh.NewServerConn(netConnection, config)
	if err != nil {
		c.log.Warn("error performing SSH handshake", slog.Any("error", err))
		return
	}
	sessionLog := c.log.With(slog.String("sessionID", hex.EncodeToString(sshConnection.SessionID())), slog.String("username", sshConnection.User()), slog.String("remoteAddress", sshConnection.RemoteAddr().String()))
	sessionLog.Debug("SSH handshake complete")
	netConnection.SetDeadline(time.Time{})
	go ssh.DiscardRequests(reqs)

	// The user has authenticated successfully, so will now block the program shutdown until their session is complete.
	handlerGroup.Go(func() {
		c.handleSSHChannelRequests(sessionLog, sshConnection, channels)
	})
}

func (c *SFTPCommand) handleSSHChannelRequests(sessionLog *slog.Logger, sshConnection ssh.Conn, channels <-chan ssh.NewChannel) {
	for channelRequest := range channels {
		// Handle only "session" channels
		if channelRequest.ChannelType() != "session" {
			channelRequest.Reject(ssh.UnknownChannelType, "unknown channel type")
			sessionLog.Warn("rejected request to create channel of unhandled type (type != 'session')")
			continue
		}
		channel, requests, err := channelRequest.Accept()
		if err != nil {
			sessionLog.Warn("error accepting channel creation request", slog.Any("error", err))
		}

		for request := range requests {
			ok := false
			if request.Type == "subsystem" {
				if len(request.Payload) >= 4 && bytes.Equal(request.Payload[4:], []byte("sftp")) {
					ok = true
					if err := c.handleSFTPChannel(sessionLog, sshConnection, channel); err != nil {
						sessionLog.Warn("error handling sftp channel", slog.Any("error", err))
						continue
					}
				} else {
					sessionLog.Warn("rejected request to access subsystem of unhandled type (first 4 bytes of payload != 'sftp')")
				}
			} else {
				sessionLog.Warn("rejected request of unhandled type (type != 'subsystem')")
			}
			if request.WantReply {
				request.Reply(ok, nil)
			}
		}
	}
}

func (c *SFTPCommand) handleSFTPChannel(sessionLog *slog.Logger, sshConnection ssh.Conn, channel ssh.Channel) error {
	username := sshConnection.User()
	srv := &sftp.Server{
		Handler: &localfs.ServerHandler{
			Root:    c.root,
			WorkDir: "/" + username,
		},
	}
	sessionLog.Info("serving sftp session")
	defer srv.GracefulStop()
	err := srv.Serve(channel)
	if err != nil {
		return fmt.Errorf("sftp server completed with error: %w", err)
	}
	sessionLog.Info("sftp session ended")
	return nil
}
