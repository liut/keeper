package reaper

import (
	"log"
	"time"
)

var defaultInterval = time.Minute * 5

type CleanFunc func() error

// Run invokes a reap function as a goroutine.
func Run(interval time.Duration, cf CleanFunc) (chan<- struct{}, <-chan struct{}) {
	if interval <= 0 {
		interval = defaultInterval
	}

	quit, done := make(chan struct{}), make(chan struct{})
	go reap(interval, cf, quit, done)
	return quit, done
}

// Quit terminates the reap goroutine.
func Quit(quit chan<- struct{}, done <-chan struct{}) {
	quit <- struct{}{}
	<-done
}

// reap with special action at set intervals.
func reap(interval time.Duration, cf CleanFunc, quit <-chan struct{}, done chan<- struct{}) {
	log.Printf("starting reaper by inerval %s ...", interval)
	ticker := time.NewTicker(interval)

	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-quit:
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
