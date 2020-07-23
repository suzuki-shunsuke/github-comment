package signal

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var once sync.Once //nolint:gochecknoglobals

// Handle calls the function "callback" when the sinal is sent.
// This is useful to support canceling by signal.
// Usage:
//   c, cancel := context.WithCancel(ctx)
//   go signal.Handle(os.Stderr, cancel)
//   ...
func Handle(stderr io.Writer, callback func()) {
	once.Do(func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(
			signalChan, syscall.SIGHUP, syscall.SIGINT,
			syscall.SIGTERM, syscall.SIGQUIT)
		sig := <-signalChan
		fmt.Fprintf(stderr, "send signal %d\n", sig)
		callback()
	})
}
