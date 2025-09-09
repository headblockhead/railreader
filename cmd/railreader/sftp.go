package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

type SFTPCommand struct {
	Addresses          []string `env:"SFTP_ADDRESSES" help:"A list of addresses to listen to." default:"127.0.0.1:822"`
	HashedPassword     string   `env:"SFTP_HASHED_PASSWORD" help:"A bcrypt hashed password to authenticate incoming SFTP connections." required:""`
	PrivateHostKeyFile []byte   `env:"SFTP_PRIVATE_HOST_KEY_FILE" help:"File containing the SFTP server's SSH private key. Must be one of these algorithms: ssh-rsa, ssh-dss, ecdsa-sha2-nistp256, ecdsa-sha2-nistp384, or ecdsa-sha2-nistp521." type:"filecontent" required:""`
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
		MaxAuthTries: 1,
	}
	config.AddHostKey(privateKey)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	go onSignal(log, func() {
		cancel()
		for _, listener := range listeners {
			log.Debug("closing listener", slog.String("address", listener.Addr().String()))
			listener.Close()
		}
	})

	var handlerGroup sync.WaitGroup

	for _, listener := range listeners {
		listenerGroup.Go(func() {
			for {
				log.Debug("waiting for new connection", slog.String("address", listener.Addr().String()))
				// Block until there is a new connection to the server or the listener is closed.
				connection, err := listener.Accept()
				if err != nil {
					// If the context has been cancelled, don't show an error message, the error is intentional.
					if ctx.Err() == nil {
						log.Error("error accepting an incoming connection", slog.Any("error", err))
					}
					break
				}
				go c.handleConnection(&handlerGroup, log, connection, config)
			}
		})
	}

	listenerGroup.Wait()
	handlerGroup.Wait()

	return nil
}

func (c *SFTPCommand) handleConnection(handlerGroup *sync.WaitGroup, log *slog.Logger, connection net.Conn, config *ssh.ServerConfig) {
	netLog := log.With(slog.GroupAttrs("net", slog.String("localAddress", connection.LocalAddr().String()), slog.String("remoteAddress", connection.RemoteAddr().String())))
	netLog.Debug("recieved new net connection")

	config.AuthLogCallback = func(conn ssh.ConnMetadata, method string, err error) {
		sshConnectionGroup := slog.GroupAttrs("ssh", slog.String("username", conn.User()))
		attemptGroup := slog.GroupAttrs("attempt", slog.String("method", method))
		attemptLog := netLog.With(sshConnectionGroup, attemptGroup)
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

	connection.SetDeadline(time.Now().Add(2 * time.Minute))
	sshConnection, channels, reqs, err := ssh.NewServerConn(connection, config)
	if err != nil {
		netLog.Error("error performing SSH handshake", slog.Any("error", err))
		return
	}
	connection.SetDeadline(time.Time{})
	go ssh.DiscardRequests(reqs)

	connectionLog := netLog.With(slog.GroupAttrs("ssh", slog.String("username", sshConnection.User())))
	connectionLog.Debug("completed SSH handshake")

	handlerGroup.Go(func() {
		c.handleSSHChannelRequests(connectionLog, channels)
	})
}

func (c *SFTPCommand) handleSSHChannelRequests(log *slog.Logger, channels <-chan ssh.NewChannel) {
	for channelRequest := range channels {
		channelLog := log.With(slog.GroupAttrs("channel", slog.String("type", channelRequest.ChannelType())))
		channelLog.Debug("handling channel creation request")
		// Handle only "session" channels
		if channelRequest.ChannelType() != "session" {
			channelRequest.Reject(ssh.UnknownChannelType, "unknown channel type")
			channelLog.Warn("rejected request to create channel of unhandled type (type != 'session')")
			continue
		}
		channel, requests, err := channelRequest.Accept()
		if err != nil {
			channelLog.Error("error accepting channel creation request", slog.Any("error", err))
		}
		channelLog.Debug("accepted channel creation request")

		for request := range requests {
			ok := false
			requestLog := channelLog.With(slog.GroupAttrs("request", slog.String("type", request.Type), slog.Bool("wantReply", request.WantReply)))
			requestLog.Debug("recieved request")
			if request.Type == "subsystem" {
				if len(request.Payload) >= 4 && bytes.Equal(request.Payload[4:], []byte("sftp")) {
					requestLog.Debug("request OK")
					ok = true
					if err := c.handleSFTPChannel(channelLog, channel); err != nil {
						channelLog.Error("error handling a channel", slog.Any("error", err))
					}
				} else {
					requestLog.Warn("rejected non-SFTP subsystem request")
				}
			} else {
				requestLog.Warn("rejected request of unhandled type (type != 'subsystem')")
			}
			if request.WantReply {
				request.Reply(ok, nil)
			}
		}
	}
	log.Debug("handled all channel requests")
}

func (c *SFTPCommand) handleSFTPChannel(log *slog.Logger, channel ssh.Channel) error {
	log.Debug("starting sftp server session")
	root, err := os.OpenRoot(c.DarwinDirectory)
	if err != nil {
		return fmt.Errorf("failed to open root directory: %w", err)
	}
	defer root.Close()
	handlers := newSFTPHandlers(log, root)
	server := sftp.NewRequestServer(channel, handlers)
	log.Debug("serving sftp session")
	defer server.Close()
	return server.Serve()
}

func newSFTPHandlers(log *slog.Logger, root *os.Root) sftp.Handlers {
	doer := sftpFileDoer{
		log:  log,
		mtx:  &sync.Mutex{},
		root: root,
	}
	return sftp.Handlers{
		FileGet:  doer,
		FilePut:  doer,
		FileCmd:  doer,
		FileList: doer,
	}
}

type sftpFileDoer struct {
	log  *slog.Logger
	mtx  *sync.Mutex
	root *os.Root
}

func addReqToLog(log *slog.Logger, req *sftp.Request) *slog.Logger {
	return log.With(slog.GroupAttrs("file", slog.String("method", req.Method), slog.String("filepath", req.Filepath)))
}

func (d sftpFileDoer) Fileread(req *sftp.Request) (io.ReaderAt, error) {
	log := addReqToLog(d.log, req)
	d.mtx.Lock()
	defer d.mtx.Unlock()
	log.Debug("file read request")
	file, err := d.root.Open(strings.TrimPrefix(req.Filepath, "/"))
	if err != nil {
		log.Error("error opening file for read", slog.Any("error", err))
		return nil, err
	}
	return file, nil
}
func (d sftpFileDoer) Filewrite(req *sftp.Request) (io.WriterAt, error) {
	log := addReqToLog(d.log, req)
	d.mtx.Lock()
	defer d.mtx.Unlock()
	log.Debug("file write request")
	file, err := d.root.OpenFile(strings.TrimPrefix(req.Filepath, "/"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		d.log.Error("error opening file for write", slog.Any("error", err))
		return nil, err
	}
	return file, nil
}
func (d sftpFileDoer) Filecmd(req *sftp.Request) error {
	log := addReqToLog(d.log, req)
	log.Debug("file command request")
	return nil
}

type listerat []os.FileInfo

// Modeled after strings.Reader's ReadAt() implementation
func (f listerat) ListAt(ls []os.FileInfo, offset int64) (int, error) {
	var n int
	if offset >= int64(len(f)) {
		return 0, io.EOF
	}
	n = copy(ls, f[offset:])
	if n < len(ls) {
		return n, io.EOF
	}
	return n, nil
}

func readdir(f fs.FS, pathname string) ([]os.FileInfo, error) {
	info, err := fs.Lstat(f, pathname)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []os.FileInfo{info}, nil
	}

	dir, err := fs.ReadDir(f, pathname)
	if err != nil {
		return nil, err
	}

	var files []os.FileInfo

	for _, file := range dir {
		info, err := file.Info()
		if err != nil {
			return nil, err
		}
		files = append(files, info)
	}

	return files, nil
}

func (d sftpFileDoer) Filelist(req *sftp.Request) (sftp.ListerAt, error) {
	log := addReqToLog(d.log, req)
	log.Debug("file list request")
	d.mtx.Lock()
	defer d.mtx.Unlock()
	switch req.Method {
	case "List":
		files, err := readdir(d.root.FS(), strings.TrimPrefix(req.Filepath, "/"))
		if err != nil {
			return nil, err
		}
		return listerat(files), nil
	case "Stat":
		file, err := fs.Lstat(d.root.FS(), strings.TrimPrefix(req.Filepath, "/"))
		if err != nil {
			return nil, err
		}
		return listerat{file}, nil
	}
	log.Warn("unsupported file list method")
	return nil, errors.New("unsupported")
}
