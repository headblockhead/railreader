package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

type SFTPCommand struct {
	Address        string `env:"SFTP_ADDRESS" help:"SFTP server address to listen on for file transfers" default:"127.0.0.1:8022"`
	HashedPassword string `env:"SFTP_HASHED_PASSWORD" help:"bcrypt hashed password for incoming SFTP connections" required:""`
	PrivateKey     []byte `env:"SFTP_PRIVATE_KEY_FILE" help:"Path to the private key file for the SFTP server" type:"filecontent" required:""`
	Logging        struct {
		Level string `enum:"debug,info,warn,error" env:"LOGGING_LEVEL" default:"warn"`
		Type  string `enum:"json,console" env:"LOGGING_TYPE" default:"json"`
	} `embed:"" prefix:"logging."`
}

func (c *SFTPCommand) Run() error {
	log := getLogger(c.Logging.Level, c.Logging.Type == "json")

	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			log.Info("Login attempt", slog.String("username", conn.User()))
			err := bcrypt.CompareHashAndPassword([]byte(c.HashedPassword), pass)
			if err == nil {
				log.Info("Login successful", slog.String("username", conn.User()))
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", conn.User())
		},
	}
	private, err := ssh.ParsePrivateKey(c.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}
	config.AddHostKey(private)

	listener, err := net.Listen("tcp", c.Address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", c.Address, err)
	}
	log.Info("Listening on", slog.String("address", c.Address))

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
		log.Info("sftp server closed.")
	}
	return nil
}
