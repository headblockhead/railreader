package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

type SFTPCommand struct {
	EnableIPv4         bool   `env:"SFTP_ENABLE_IPV4" help:"Whether to listen on IPv4." default:"true"`
	EnableIPv6         bool   `env:"SFTP_ENABLE_IPV6" help:"Whether to listen on IPv6." default:"false"`
	Address            string `env:"SFTP_ADDRESS" help:"Address (v4 or v6) to listen on for file transfers from the Rail Data Marketplace." default:"127.0.0.1:2022"`
	HashedPasswordFile []byte `env:"SFTP_HASHED_PASSWORD_FILE" help:"File containing a bcrypt hashed password to authenticate incoming SFTP connections" type:"filecontent" required:""`
	PrivateKeyFile     []byte `env:"SFTP_PRIVATE_KEY_FILE" help:"File containing the SFTP server's private key" type:"filecontent" required:""`
	Logging            struct {
		Level  string `enum:"debug,info,warn,error" env:"LOG_LEVEL" default:"warn"`
		Format string `enum:"json,console" env:"LOG_FORMAT" default:"json"`
	} `embed:"" prefix:"log."`
}

func (c *SFTPCommand) Run() error {
	log := getLogger(c.Logging.Level, c.Logging.Format == "json")

	listenerContext, listenerCancel := context.WithCancel(context.Background())
	go cancelOnSignal(listenerCancel, log)

	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			log.Info("new login attempt", slog.String("username", conn.User()), slog.String("local-addr-network", conn.LocalAddr().Network()), slog.String("local-addr", conn.LocalAddr().String()), slog.String("remote-addr-network", conn.RemoteAddr().Network()), slog.String("remote-addr", conn.RemoteAddr().String()))
			err := bcrypt.CompareHashAndPassword(c.HashedPasswordFile, pass) // constant time comparison
			if err == nil {
				log.Info("successful login attempt", slog.String("username", conn.User()))
				return nil, nil
			}
			log.Info("unsuccessful login attempt", slog.String("username", conn.User()), slog.String("error", err.Error()))
			return nil, fmt.Errorf("password rejected for %q", conn.User())
		},
	}
	private, err := ssh.ParsePrivateKey(c.PrivateKeyFile)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}
	config.AddHostKey(private)
	go listen()
	return nil
}

func listen() {
	listener, err := net.Listen("tcp", c.Address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", c.Address, err)
	}
	log.Info("Listening on", slog.String("address", c.Address))

	for {
		// Block until we get a new connection
		nConn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept incoming connection: %w", err)
		}
		_, chans, reqs, err := ssh.NewServerConn(nConn, config)
		if err != nil {
			return fmt.Errorf("failed to handshake: %w", err)
		}
		go ssh.DiscardRequests(reqs)

		for newChannel := range chans {
			log.Info("New channel", slog.String("type", newChannel.ChannelType()))
			if newChannel.ChannelType() != "session" {
				newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
				log.Warn("Unknown channel type", slog.String("type", newChannel.ChannelType()))
				continue
			}
			channel, requests, err := newChannel.Accept()
			if err != nil {
				return fmt.Errorf("could not accept channel: %w", err)
			}
			log.Info("Channel accepted")

			// Sessions have out-of-band requests such as "shell",
			// "pty-req" and "env".  Here we handle only the
			// "subsystem" request.
			go func(in <-chan *ssh.Request) {
				for req := range in {
					log.Info("Request", slog.String("type", req.Type), slog.Bool("want-reply", req.WantReply))
					ok := false
					switch req.Type {
					case "subsystem":
						log.Info("Subsystem", slog.String("name", string(req.Payload[4:])))
						if string(req.Payload[4:]) == "sftp" {
							ok = true
						}
					}
					log.Info("Request processed", slog.String("type", req.Type), slog.Bool("accepted", ok))
					req.Reply(ok, nil)
				}
			}(requests)

			serverOptions := []sftp.ServerOption{}

			server, err := sftp.NewServer(
				channel,
				serverOptions...,
			)
			if err != nil {
				log.Error("Failed to start sftp server:", slog.String("error", err.Error()))
				continue
			}
			if err := server.Serve(); err != nil {
				if err != io.EOF {
					log.Error("sftp server completed with error:", slog.String("error", err.Error()))
				} else {
					log.Info("sftp server completed session.")
				}
			}
			server.Close()
		}
	}

}
