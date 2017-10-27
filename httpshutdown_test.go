package httpshutdown_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lag13/httpshutdown"
)

// TestRealServerIsStartedAndShutsDown makes sure that the server
// starts up, responds to requests, and shuts down properly. This test
// starts up a real server on a static port. I did some research into
// how to dynamically allocate ports (seems that you use address
// "127.0.0.1:0") but do not think it is possible to dynamically get a
// port in this situation since net.Listen("tcp", addr) (which returns
// the port information) is called inside of
// http.Server.ListenAndServe() and it cannot be inspected. So we are
// stuck hard coding a port which feels bad but in practice who really
// cares, it works just fine.
func TestRealServerIsStartedAndShutsDown(t *testing.T) {
	const testEndpoint = "127.0.0.1:8087"
	var handlerRespBody = []byte("hello world!")
	http.Handle("/hello-world", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(handlerRespBody)
	}))
	shutdown := make(chan os.Signal)
	go func() {
		// give the server some time to start up
		time.Sleep(100 * time.Millisecond)
		resp, err := http.Get("http://" + testEndpoint + "/hello-world")
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
	}()
	srv := &http.Server{Addr: testEndpoint, Handler: http.DefaultServeMux}
	if err := httpshutdown.ListenAndServe(srv, 0, shutdown); err != nil {
		panic(fmt.Sprintf("server shutdown unexpectedly: %v", err))
	}
}

// TestContextTimeoutError tests that we get the expected error
// message when the server exceeds the timeout when shutting down.
func TestContextTimeoutError(t *testing.T) {
	const testEndpoint = "127.0.0.1:8088"
	http.Handle("/sleep", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
	}))
	srv := &http.Server{Addr: testEndpoint, Handler: http.DefaultServeMux}
	shutdown := make(chan os.Signal)
	go func() {
		// give the server some time to start up
		time.Sleep(100 * time.Millisecond)
		go func() {
			resp, err := http.Get("http://" + testEndpoint + "/sleep")
			if err != nil {
				panic(fmt.Sprintf("got error when sending request: %v", err))
			}
			defer resp.Body.Close()
		}()
		// give some time for the request to send then shut
		// down the server
		time.Sleep(100 * time.Millisecond)
		shutdown <- os.Interrupt
	}()
	err := httpshutdown.ListenAndServe(srv, 300*time.Millisecond, shutdown)
	if err == nil {
		t.Errorf("expected to get a non-nil shutdown error")
	} else if got, want := err.Error(), "context deadline exceeded"; got != want {
		t.Errorf("got error %q, wanted %q", got, want)
	}
}

// TestListenAndServeError tests that the expected error is returned
// when the server exits unexpectedly from ListenAndServe().
func TestListenAndServeError(t *testing.T) {
	srv := &http.Server{Addr: "invalid-addr:1", Handler: http.DefaultServeMux}
	err := httpshutdown.ListenAndServe(srv, 0, make(chan os.Signal))
	if got, want := err.Error(), "invalid-addr"; !strings.Contains(got, want) {
		t.Errorf("error returned was %q, but should contain the text %q", got, want)
	}
}
