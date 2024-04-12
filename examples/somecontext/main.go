package main

import (
	"context"
	"log"
	"time"

	"github.com/initialed85/some/somecontext"
)

func work(ctx context.Context, i int, tasks chan struct{}) {
	// this will find the created TimeoutWaitGroup in the parent context, so the
	// done function will apply to it
	ctx, _, done, _ := somecontext.WithCancelAndDoneAndWait(ctx)

	for {
		select {
		case <-ctx.Done():
			// time consuming cleanup
			log.Printf("%v: cleaning up...", i)
			time.Sleep(time.Second * 1)
			done()
			log.Printf("%v: done...", i)
			return
		case <-tasks:
			// some work
			log.Printf("%v: doing work...", i)
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func main() {
	// this will seed the returned context with a TimeoutWaitGroup, which is what has
	// provided us with the done and wait functions; additionally, it's called
	// .Add(1) on the TimeoutWaitGroup, meaning if we use this wait function, it will
	// block until the done function is called
	ctx, cancel, done, wait := somecontext.WithCancelAndDoneAndWait(context.Background())

	// this is just sugar for:
	//   defer cancel()
	//   defer done()
	//   defer wait(time.Second * 60)
	defer somecontext.Cleanup(cancel, done, wait)

	tasks := make(chan struct{}, 1000000)

	for i := 0; i < 1000000; i++ {
		tasks <- struct{}{}
	}

	for i := 0; i < 4; i++ {
		go work(ctx, i, tasks)
	}

	// wait impatiently for some work to happen
	time.Sleep(time.Second * 2)

	// get sick of waiting and cancel the workers
	cancel()

	// now that cancel has been triggered, our deferred cleanup should gracefully shut us down
}
