package healthcheck

import (
	"net"
	"net/http"
)

// listener allows for testing of ServiceHealthServer and ProxierHealthServer.
type Listener interface {
	// Listen is very much like net.Listen, except the first arg (network) is
	// fixed to be "tcp".
	Listen(addr string) (net.Listener, error)
}

// httpServerFactory allows for testing of ServiceHealthServer and ProxierHealthServer.
type httpServerFactory interface {
	// New creates an instance of a type satisfying HTTPServer.  This is
	// designed to include http.Server.
	New(addr string, handler http.Handler) httpServer
}

// httpServer allows for testing of ServiceHealthServer and ProxierHealthServer.
// It is designed so that http.Server satisfies this interface,
type httpServer interface {
	Serve(listener net.Listener) error
}

// Implement listener in terms of net.Listen.
type stdNetListener struct{}

func (stdNetListener) Listen(addr string) (net.Listener, error) {
	return net.Listen("tcp", addr)
}

var _ Listener = stdNetListener{}

// Implement httpServerFactory in terms of http.Server.
type stdHTTPServerFactory struct{}

func (stdHTTPServerFactory) New(addr string, handler http.Handler) httpServer {
	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}

var _ httpServerFactory = stdHTTPServerFactory{}

