package main

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"iter"
	"log/slog"
	"net"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/pkg/sftp/v2"
	sshfx "github.com/pkg/sftp/v2/encoding/ssh/filexfer"
	"github.com/pkg/sftp/v2/encoding/ssh/filexfer/openssh"
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

	var listenerGroup sync.WaitGroup
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
	netLog.Debug("received new net connection")

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
			requestLog.Debug("received request")
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
	root.Mkdir("sftp", 0755)
	srv := &sftp.Server{
		Handler: NewOSRootSFTPHandler(log, root),
	}
	log.Debug("serving sftp session")
	defer srv.GracefulStop()
	return srv.Serve(channel)
}

// adaptation of puellanivis (Cassondra Foesch's) sftp localfs to use os.Root for filesystem containment, with some variables and methods renamed to be consistent with the rest of railreader.
type osRootSFTPHandler struct {
	sftp.UnimplementedServerHandler
	log              *slog.Logger
	root             *os.Root
	workingDirectory string
	fileHandlerCount atomic.Uint64
}

func NewOSRootSFTPHandler(log *slog.Logger, root *os.Root) *osRootSFTPHandler {
	return &osRootSFTPHandler{
		log:              log,
		root:             root,
		workingDirectory: "/sftp",
	}
}

func (h *osRootSFTPHandler) toLocalPath(p string) (string, error) {
	if h.workingDirectory != "" && !path.IsAbs(p) {
		p = path.Join(h.workingDirectory, p)
	} else {
		p = path.Clean(p)
	}
	if p == "" {
		h.log.Warn("failed to convert to local path, provided path is empty", slog.String("path", p))
		return "", sshfx.StatusNoSuchFile
	}
	p = strings.TrimPrefix(p, "/")
	return p, nil
}

func (h *osRootSFTPHandler) Mkdir(_ context.Context, req *sshfx.MkdirPacket) error {
	h.log.Debug("mkdir", slog.String("path", req.Path))
	lpath, err := h.toLocalPath(req.Path)
	if err != nil {
		return err
	}
	var perms sshfx.FileMode = 0755
	if req.Attrs.HasPermissions() {
		perms = req.Attrs.GetPermissions().Perm()
	}
	return h.root.Mkdir(lpath, fs.FileMode(perms))
}

func (h *osRootSFTPHandler) Remove(_ context.Context, req *sshfx.RemovePacket) error {
	h.log.Debug("remove", slog.String("path", req.Path))
	lpath, err := h.toLocalPath(req.Path)
	if err != nil {
		return err
	}
	fi, err := h.root.Stat(lpath)
	if err != nil {
		h.log.Warn("error stating file to remove", slog.String("path", lpath), slog.Any("error", err))
		return err
	}
	if fi.IsDir() {
		h.log.Warn("attempted to remove a directory with Remove")
		return &fs.PathError{
			Op:   "remove",
			Path: lpath,
			Err:  fmt.Errorf("is a directory"),
		}
	}
	return h.root.Remove(lpath)
}

func (h *osRootSFTPHandler) Rename(_ context.Context, req *sshfx.RenamePacket) error {
	h.log.Debug("rename", slog.String("from", req.OldPath), slog.String("to", req.NewPath))
	from, err := h.toLocalPath(req.OldPath)
	if err != nil {
		return err
	}
	to, err := h.toLocalPath(req.NewPath)
	if err != nil {
		return err
	}
	if _, err := h.root.Stat(to); !errors.Is(err, fs.ErrNotExist) {
		if err == nil {
			h.log.Warn("attempted to rename to a path that already exists", slog.String("path", to))
			return fs.ErrExist
		}
		h.log.Warn("error stating target path of rename", slog.String("path", to), slog.Any("error", err))
		return err
	}
	return h.root.Rename(from, to)
}

func (h *osRootSFTPHandler) POSIXRename(_ context.Context, req *openssh.POSIXRenameExtendedPacket) error {
	h.log.Debug("posix-rename", slog.String("from", req.OldPath), slog.String("to", req.NewPath))
	from, err := h.toLocalPath(req.OldPath)
	if err != nil {
		return err
	}
	to, err := h.toLocalPath(req.NewPath)
	if err != nil {
		return err
	}
	return h.root.Rename(from, to)
}

func (h *osRootSFTPHandler) Rmdir(_ context.Context, req *sshfx.RmdirPacket) error {
	h.log.Debug("rmdir", slog.String("path", req.Path))
	lpath, err := h.toLocalPath(req.Path)
	if err != nil {
		return err
	}
	fi, err := h.root.Stat(lpath)
	if err != nil {
		h.log.Warn("error stating directory to remove", slog.String("path", lpath), slog.Any("error", err))
		return err
	}
	if !fi.IsDir() {
		h.log.Warn("attempted to remove a non-directory with Rmdir")
		return &fs.PathError{
			Op:   "rmdir",
			Path: lpath,
			Err:  fmt.Errorf("not a directory"),
		}
	}
	return h.root.Remove(lpath)
}

func (h *osRootSFTPHandler) SetStat(_ context.Context, req *sshfx.SetStatPacket) error {
	h.log.Debug("setstat", slog.String("path", req.Path), slog.Any("attrs", req.Attrs))
	lpath, err := h.toLocalPath(req.Path)
	if err != nil {
		return err
	}
	if req.Attrs.HasSize() {
		// Truncate is not supported by os.Root, log and ignore.
		h.log.Warn("size set with setstat, but isn't supported yet")
		/*sz := req.Attrs.GetSize()*/
		/*   if err := h.Root.Truncate(lpath, int64(sz)); err != nil {*/
		/*return err*/
		/*}*/
	}
	if req.Attrs.HasUIDGID() {
		h.log.Debug("changing ownership", slog.String("path", lpath))
		uid, gid := req.Attrs.GetUIDGID()
		if err := h.root.Chown(lpath, int(uid), int(gid)); err != nil {
			return err
		}
	}
	if req.Attrs.HasPermissions() {
		h.log.Debug("changing permissions", slog.String("path", lpath))
		perms := req.Attrs.GetPermissions()
		if err := h.root.Chmod(lpath, fs.FileMode(perms.Perm())); err != nil {
			return err
		}
	}
	if req.Attrs.HasACModTime() {
		h.log.Debug("changing access and modification times", slog.String("path", lpath))
		atime, mtime := req.Attrs.GetACModTime()
		if err := h.root.Chtimes(lpath, time.Unix(int64(atime), 0), time.Unix(int64(mtime), 0)); err != nil {
			return err
		}
	}
	return nil
}

func (h *osRootSFTPHandler) Symlink(_ context.Context, req *sshfx.SymlinkPacket) error {
	h.log.Debug("symlink", slog.String("target", req.TargetPath), slog.String("link", req.LinkPath))
	target, err := h.toLocalPath(req.TargetPath)
	if err != nil {
		return err
	}
	link, err := h.toLocalPath(req.LinkPath)
	if err != nil {
		return err
	}
	return h.root.Symlink(target, link)
}

func fileInfoToAttrs(fi fs.FileInfo) *sshfx.Attributes {
	attrs := new(sshfx.Attributes)
	attrs.SetSize(uint64(fi.Size()))
	attrs.SetPermissions(sshfx.FromGoFileMode(fi.Mode()))

	mtime := uint32(fi.ModTime().Unix())
	attrs.SetACModTime(mtime, mtime)

	if statt, ok := fi.Sys().(*syscall.Stat_t); ok {
		attrs.SetUIDGID(statt.Uid, statt.Gid)
		attrs.SetACModTime(uint32(statt.Atim.Sec), uint32(statt.Mtim.Sec))
	}

	return attrs
}

func (h *osRootSFTPHandler) LStat(_ context.Context, req *sshfx.LStatPacket) (*sshfx.Attributes, error) {
	h.log.Debug("lstat", slog.String("path", req.Path))
	lpath, err := h.toLocalPath(req.Path)
	if err != nil {
		return nil, err
	}
	fi, err := h.root.Lstat(lpath)
	if err != nil {
		h.log.Error("error stating file", slog.String("path", lpath), slog.Any("error", err))
		return nil, err
	}
	return fileInfoToAttrs(fi), nil
}

func (h *osRootSFTPHandler) Stat(_ context.Context, req *sshfx.StatPacket) (*sshfx.Attributes, error) {
	h.log.Debug("stat", slog.String("path", req.Path))
	lpath, err := h.toLocalPath(req.Path)
	if err != nil {
		return nil, err
	}
	fi, err := h.root.Stat(lpath)
	if err != nil {
		h.log.Error("error stating file", slog.String("path", lpath), slog.Any("error", err))
		return nil, err
	}
	return fileInfoToAttrs(fi), nil
}

func (h *osRootSFTPHandler) ReadLink(_ context.Context, req *sshfx.ReadLinkPacket) (string, error) {
	h.log.Debug("readlink", slog.String("path", req.Path))
	lpath, err := h.toLocalPath(req.Path)
	if err != nil {
		return "", err
	}
	return h.root.Readlink(lpath)
}

func (h *osRootSFTPHandler) RealPath(_ context.Context, req *sshfx.RealPathPacket) (string, error) {
	h.log.Debug("realpath", slog.String("path", req.Path))
	lpath, err := h.toLocalPath(req.Path)
	if err != nil {
		return "", err
	}
	return path.Join("/", filepath.ToSlash(lpath)), nil
}

// File wraps an [os.File] to provide the additional operations necessary to implement [sftp.FileHandler].
type File struct {
	*os.File

	root os.Root

	filename string
	handle   string
	idLookup sftp.NameLookup

	mu      sync.Mutex
	lastErr error
	lastEnt *sshfx.NameEntry
	entries []fs.FileInfo
}

// Handle returns the SFTP handle associated with the file.
func (f *File) Handle() string {
	return f.handle
}

// Stat overrides the [os.File.Stat] receiver method
// by converting the [fs.FileInfo] into a [sshfx.Attributes].
func (f *File) Stat() (*sshfx.Attributes, error) {
	fi, err := f.File.Stat()
	if err != nil {
		return nil, err
	}

	return fileInfoToAttrs(fi), nil
}

// rangedir returns an iterator over the directory entries of the directory.
// It will only ever yield either a [fs.FileInfo] or an error, never both.
// No error will be yielded until all available FileInfos have been yielded.
// Only one error will be yielded per invocation.
//
// We do not expose an iterator, because none has been standardized yet,
// and we do not want to accidentally implement an API inconsistent with future standards.
// However, for internal usage, we can separate the paginated Readdir code from the conversion to SFTP entries.
//
// Callers must guarantee synchronization by either holding the file lock, or holding an exclusive reference.
func (f *File) rangedir(grow func(int)) iter.Seq2[fs.FileInfo, error] {
	return func(yield func(fs.FileInfo, error) bool) {
		for {
			grow(len(f.entries))

			for i, entry := range f.entries {
				if !yield(entry, nil) {
					// This is a break condition.
					// We need to remove all entries that have been consumed,
					// and that includes the one we are currently on.
					f.entries = slices.Delete(f.entries, 0, i+1)
					return
				}
			}

			// We have consumed all of the saved entries, so we remove everything.
			f.entries = slices.Delete(f.entries, 0, len(f.entries))

			if f.lastErr != nil {
				yield(nil, f.lastErr)
				f.lastErr = nil
				return
			}

			// We cannot guarantee we only get entries, or an error, never both.
			// So we need to just save these, and loop.
			f.entries, f.lastErr = f.Readdir(128)
		}
	}
}

// ReadDir overrides the [os.File.ReadDir] receiver method
// by converting the slice of [fs.DirEntry] into into a slice of [sshfx.NameEntry].
func (f *File) ReadDir(maxDataLen uint32) (entries []*sshfx.NameEntry, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.lastEnt != nil {
		// Last ReadDir left an entry for us to include in this call.
		entries = append(entries, f.lastEnt)
		f.lastEnt = nil
	}

	grow := func(more int) {
		entries = slices.Grow(entries, more)
	}

	var size int
	for fi, err := range f.rangedir(grow) {
		if err != nil {
			if len(entries) != 0 {
				return entries, nil
			}

			return nil, err
		}

		attrs := fileInfoToAttrs(fi)

		entry := &sshfx.NameEntry{
			Filename: fi.Name(),
			Longname: sftp.FormatLongname(fi, f.idLookup),
			Attrs:    *attrs,
		}

		size += entry.MarshalSize()

		if size > int(maxDataLen) {
			// This would exceed the packet data length,
			// so save this one for the next call,
			// and return.
			f.lastEnt = entry
			break
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// SetStat implements [sftp.SetStatFileHandler].
func (f *File) SetStat(attrs *sshfx.Attributes) (err error) {
	if len(attrs.Extended) > 0 {
		err = &sshfx.StatusPacket{
			StatusCode:   sshfx.StatusOpUnsupported,
			ErrorMessage: "unsupported fsetstat: extended atributes",
		}
	}

	if attrs.HasSize() {
		sz := attrs.GetSize()
		err = cmp.Or(f.Truncate(int64(sz)), err)
	}

	if attrs.HasPermissions() {
		perm := attrs.GetPermissions()
		err = cmp.Or(f.Chmod(fs.FileMode(perm.Perm())), err)
	}

	if attrs.HasACModTime() {
		return &sshfx.StatusPacket{
			StatusCode:   sshfx.StatusOpUnsupported,
			ErrorMessage: "unsupported fsetstat: atime/mtime",
		}
		/*atime, mtime := attrs.GetACModTime()*/
		/*err = cmp.Or(os.Chtimes(f.filename, time.Unix(int64(atime), 0), time.Unix(int64(mtime), 0)), err)*/
	}

	if attrs.HasUIDGID() {
		uid, gid := attrs.GetUIDGID()
		err = cmp.Or(f.Chown(int(uid), int(gid)), err)
	}

	return err
}

func (h *osRootSFTPHandler) openfile(path string, flag int, mod fs.FileMode) (*File, error) {
	f, err := h.root.OpenFile(path, flag, mod)
	if err != nil {
		h.log.Error("error opening file", slog.String("path", path), slog.Int("flag", flag), slog.Any("error", err))
		return nil, err
	}

	return &File{
		filename: path,
		handle:   fmt.Sprint(h.fileHandlerCount.Add(1)),
		File:     f,
	}, nil
}

func (h *osRootSFTPHandler) Open(_ context.Context, req *sshfx.OpenPacket) (sftp.FileHandler, error) {
	h.log.Debug("open", slog.String("filename", req.Filename), slog.Any("attrs", req.Attrs))
	lpath, err := h.toLocalPath(req.Filename)
	if err != nil {
		return nil, err
	}

	var osFlags int
	switch {
	case req.PFlags&sshfx.FlagRead != 0:
		if req.PFlags&sshfx.FlagWrite != 0 {
			osFlags |= os.O_RDWR
		} else {
			osFlags |= os.O_RDONLY
		}
	case req.PFlags&sshfx.FlagWrite != 0:
		osFlags |= os.O_WRONLY
	default:
		return nil, fs.ErrInvalid
	}

	// Don't use O_APPEND flag as it conflicts with WriteAt.
	// The sshfx.FlagAppend is a no-op here as the client sends the offsets anyways.
	if req.PFlags&sshfx.FlagCreate != 0 {
		osFlags |= os.O_CREATE
	}
	if req.PFlags&sshfx.FlagTruncate != 0 {
		osFlags |= os.O_TRUNC
	}
	if req.PFlags&sshfx.FlagExclusive != 0 {
		osFlags |= os.O_EXCL
	}

	// Like OpenSSH, we only handle permissions here, and only when the file is being created.
	// Otherwise, the permissions are ignored.
	var perms sshfx.FileMode = 0666
	if req.Attrs.HasPermissions() {
		perms = req.Attrs.GetPermissions().Perm()
	}

	return h.openfile(lpath, osFlags, fs.FileMode(perms))
}

func (h *osRootSFTPHandler) OpenDir(_ context.Context, req *sshfx.OpenDirPacket) (sftp.DirHandler, error) {
	h.log.Debug("opendir", slog.String("path", req.Path))
	lpath, err := h.toLocalPath(req.Path)
	if err != nil {
		return nil, err
	}

	return h.openfile(lpath, os.O_RDONLY, 0)
}
