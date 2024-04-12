package main

import (
	"log"
	"runtime"
	"time"

	"github.com/initialed85/some/somesync"
)

func main() {
	// we can create the pointer flavour like as below, but we could also
	// create the reference flavour with somesync.TimeoutWaitGroup{}
	wg := new(somesync.TimeoutWaitGroup)

	log.Printf("waiting...")
	wg.Wait()
	log.Printf("fell through instantly.")

	wg.Add(1)

	go func() {
		log.Printf("calling done in 2 seconds...")
		time.Sleep(time.Second * 2)
		wg.Done()
		log.Printf("done called.")
	}()
	runtime.Gosched()

	log.Printf("waiting...")
	wg.Wait(time.Second * 1)
	log.Printf("timed out.")

	// this could block forever, but it'll also block for a second because
	// the goroutine in the background was told to complete the wait group
	// after 2 seconds
	log.Printf("waiting...")
	wg.Wait()
	log.Printf("done.")
}
