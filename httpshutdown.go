package httpshutdown

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

// httpServer is used to make one particular unit test (where the
// server errors when doing ListenAndServe) simpler. In practice you
// would pass in an *http.Server.
type httpServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

// ListenAndServe will start a server and shut it down within a
// specified timeout if it recieves any message on the "shutdown"
// channel.
func ListenAndServe(srv httpServer, timeout time.Duration, shutdown chan os.Signal) error {
	serverErr := make(chan error)
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			serverErr <- err
		}
	}()
	select {
	case <-shutdown:
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("requests did not complete before shutdown finished: %v", err)
		}
	case err := <-serverErr:
		return fmt.Errorf("server exited unexpectedly: %v", err)
	}
	return nil
}
