// Most of these tests start up a real *http.Server and make sure
// things work as expected. I could have used a mock but since this
// code really is just a convenience wrapper for starting and stopping
// a server, I wanted to make sure it actually works in practice.
package httpshutdown

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestRealServerIsStartedAndShutsDown passes in a real *http.Server
// and makes sure that it starts up and shuts down properly. It's
// basically my little sanity check that the happy path works in the
// real world.
func TestRealServerIsStartedAndShutsDown(t *testing.T) {
	const testEndpoint = "127.0.0.1:8087"
	var handlerRespBody = []byte("hello world!")
	http.Handle("/some-route", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(handlerRespBody)
	}))
	srv := &http.Server{Addr: testEndpoint, Handler: http.DefaultServeMux}
	shutdown := make(chan os.Signal)
	serverDidShutdown := make(chan struct{})
	go func() {
		if err := ListenAndServe(srv, 0, shutdown); err != nil {
			panic(fmt.Sprintf("server shutdown unexpectedly: %v", err))
		}
		serverDidShutdown <- struct{}{}
	}()
	// give the server some time to start up
	time.Sleep(100 * time.Millisecond)
	resp, err := http.Get("http://" + testEndpoint + "/some-route")
	if err != nil {
		panic(fmt.Sprintf("got error when sending request: %v", err))
	}
	defer resp.Body.Close()
	sut, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("got error when reading response: %v", err))
	}
	if got, want := string(sut), string(handlerRespBody); got != want {
		t.Errorf("got response body %q, wanted %q", got, want)
	}
	shutdown <- os.Interrupt
	// ensures that the server properly shuts down.
	<-serverDidShutdown
}

// TestContextTimeoutError tests that we get the expected error
// message when the server exceeds the timeout when shutting down.
func TestContextTimeoutError(t *testing.T) {
	const testEndpoint = "127.0.0.1:8088"
	http.Handle("/some-other-route", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
	}))
	srv := &http.Server{Addr: testEndpoint, Handler: http.DefaultServeMux}
	shutdown := make(chan os.Signal)
	shutdownErr := make(chan error)
	go func() {
		shutdownErr <- ListenAndServe(srv, 300*time.Millisecond, shutdown)
	}()
	// give the server some time to start up
	time.Sleep(100 * time.Millisecond)
	go func() {
		resp, err := http.Get("http://" + testEndpoint + "/some-other-route")
		if err != nil {
			panic(fmt.Sprintf("got error when sending request: %v", err))
		}
		defer resp.Body.Close()
	}()
	// give some time for the request to send
	time.Sleep(100 * time.Millisecond)
	shutdown <- os.Interrupt
	// wait for the server to shutdown
	err := <-shutdownErr
	if err == nil {
		t.Errorf("expected to get a non-nil shutdown error")
	} else {
		if got, want := err.Error(), "context deadline exceeded"; got != want {
			t.Errorf("got error %q, wanted %q", got, want)
		}
	}
}

type mockHttpServer struct {
	listenAndServeErr error
}

func (m *mockHttpServer) ListenAndServe() error {
	return m.listenAndServeErr
}

func (m *mockHttpServer) Shutdown(ctx context.Context) error {
	return nil
}

// TestListenAndServeError tests that the expected error is returned
// when the server exits unexpectedly from ListenAndServe().
func TestListenAndServeError(t *testing.T) {
	mockServer := &mockHttpServer{
		listenAndServeErr: errors.New("hello there"),
	}
	shutdown := make(chan os.Signal)
	err := ListenAndServe(mockServer, 0, shutdown)
	if got, want := err.Error(), "hello there"; got != want {
		t.Errorf("error returned had text %q, wanted %q", got, want)
	}
}
