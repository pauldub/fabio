package exit

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var wg sync.WaitGroup
var quit = make(chan bool)

// Listen will listen to OS signals (currently SIGINT, SIGKILL, SIGTERM)
// and will trigger the callback when signal are received from OS
func Listen(fn func(os.Signal)) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		// we use buffered to mitigate losing the signal
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt, os.Kill, syscall.SIGTERM)

		var sig os.Signal
		select {
		case sig = <-sigchan:
		case <-quit:
		}
		if fn != nil {
			fn(sig)
		}
	}()
}

// Fatal is a replacement for log.Fatal which will trigger
// the closure of all registered exit handlers and waits
// for their completion and then call os.Exit(1).
func Fatal(v ...interface{}) {
	log.Print(v...)
	close(quit)
	wg.Wait()
	os.Exit(1)
}

// Fatalf is a replacement for log.Fatalf and behaves
// like Fatal.
func Fatalf(format string, v ...interface{}) {
	log.Printf(format, v...)
	close(quit)
	wg.Wait()
	os.Exit(1)
}

// Wait waits for all registered exit handlers to complete.
func Wait() {
	wg.Wait()
}
