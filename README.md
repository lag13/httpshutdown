# httpshutdown
Provides a small wrapper function around golang's `ListenAndServe` and
`Shutdown` functions to start up a server which can be gracefully shut
down. Created because the code required to start and then shutdown a
server upon receiving some sort of signal was *just* enough that I
didn't feel like repeating it across multiple golang http server
repositories.

# Example
The following example starts up a server and when the go program
receives a SIGTERM signal the server will be gracefully stopped within
5 seconds:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lag13/httpshutdown"
)

func main() {
	const exposedPort = 8080
	http.Handle("/health-check", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("healthy!\n"))
	}))
	http.Handle("/sleep", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		w.Write([]byte("done sleeping!\n"))
	}))
	log.Printf("listening on port %d", exposedPort)
	srv := &http.Server{Addr: fmt.Sprintf("localhost:%d", exposedPort), Handler: http.DefaultServeMux}
	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, syscall.SIGTERM)
	if err := httpshutdown.ListenAndServe(srv, 5*time.Second, shutdown); err != nil {
		log.Fatalf("server did not shutdown properly: %v", err)
	}
}
```
