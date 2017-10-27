package httpshutdown_test

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lag13/httpshutdown"
)

func ExampleListenAndServe() {
	http.Handle("/hello-world", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world!\n"))
	}))
	http.Handle("/sleep", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		w.Write([]byte("done sleeping!\n"))
	}))
	srv := &http.Server{Addr: "localhost:8080", Handler: http.DefaultServeMux}
	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, syscall.SIGTERM)
	if err := httpshutdown.ListenAndServe(srv, 5*time.Second, shutdown); err != nil {
		log.Fatalf("server did not shutdown properly: %v", err)
	}
}
