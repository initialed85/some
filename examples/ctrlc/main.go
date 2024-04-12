package main

import (
	"context"
	"log"
	"time"

	"github.com/initialed85/some/ctrlc"
	"github.com/initialed85/some/somecontext"
)

func work(ctx context.Context, i int, tasks chan struct{}) {
	ctx, _, done, _ := somecontext.WithCancelAndDoneAndWait(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Printf("%v: cleaning up...", i)
			time.Sleep(time.Second * 1)
			done()
			log.Printf("%v: done...", i)
			return
		case <-tasks:
			log.Printf("%v: doing work...", i)
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func main() {
	ctx, cancel, done, wait := somecontext.WithCancelAndDoneAndWait(context.Background())

	// we just replace our cleanup with this
	defer ctrlc.WaitForCtrlC(ctx, cancel, done, wait)

	tasks := make(chan struct{}, 1000000)

	for i := 0; i < 1000000; i++ {
		tasks <- struct{}{}
	}

	for i := 0; i < 4; i++ {
		go work(ctx, i, tasks)
	}

	// we'll sit inside the defer, waiting for Ctrl + C at this point
}
