// Package httpshutdown provides a utility function to start and
// gracefully shut down a server.
package httpshutdown

import (
	"context"
	"net/http"
	"os"
	"time"
)

// ListenAndServe will start a server and shut it down within a
// specified timeout if it receives any message on the "shutdown"
// channel.
func ListenAndServe(srv *http.Server, timeout time.Duration, shutdown chan os.Signal) error {
	shutdownErr := make(chan error)
	go func() {
		<-shutdown
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		shutdownErr <- srv.Shutdown(ctx)
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return <-shutdownErr
}
