package grpc

import (
	"context"
	"log/slog"
	"net"

	"github.com/headblockhead/railreader/egesters/grpc/proto"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

// Egester implements the interface railreader.Egester.
type Egester struct {
	proto.UnimplementedRailReaderServer

	ctx    context.Context
	cancel context.CancelFunc
	log    *slog.Logger
	dbpool *pgxpool.Pool
	server *grpc.Server
}

func NewEgester(ctx context.Context, log *slog.Logger, dbpool *pgxpool.Pool) (*Egester, error) {
	cctx, cancel := context.WithCancel(ctx)
	server := grpc.NewServer()
	proto.RegisterRailReaderServer(server, &Egester{})
	return &Egester{
		ctx:    cctx,
		cancel: cancel,
		log:    log,
		dbpool: dbpool,
		server: server,
	}, nil
}

func (e *Egester) Close() error {
	e.cancel()
	e.server.GracefulStop()
	return nil
}

func (e *Egester) Serve(l net.Listener) error {
	return e.server.Serve(l)
}
