package railreader

import (
	"net"
)

type Egester interface {
	Serve(net.Listener) error
	Close() error
}
