package reaper

import (
	"log"
	"time"
)

const defaultInterval = time.Minute * 5

// LoopFunc ...
type LoopFunc func() error

// Run invokes a reap function as a goroutine.
func Run(interval time.Duration, cf LoopFunc) (chan<- struct{}, <-chan struct{}) {
	if interval <= 0 {
		interval = defaultInterval
	}

	quit, done := make(chan struct{}), make(chan struct{})
	go reap(interval, cf, quit, done)
	return quit, done
}

// Quit terminates the reap goroutine.
func Quit(quit chan<- struct{}, done <-chan struct{}) {
	log.Print("quiting")
	quit <- struct{}{}
	<-done
}

// reap with special action at set intervals.
func reap(interval time.Duration, cf LoopFunc, quit <-chan struct{}, done chan<- struct{}) {
	log.Printf("starting loop by interval %s ...", interval)
	ticker := time.NewTicker(interval)

	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-quit:
			log.Print("got quit sign")
			// Handle the quit signal.
			done <- struct{}{}
			return
		case <-ticker.C:
			// Execute function of clean.
			if err := cf(); err != nil {
				log.Printf("reaper: ERROR: %v", err)
			}
		}
	}
}
