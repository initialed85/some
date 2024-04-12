# some

Some Go utils / helpers for me to use, maybe you as well.

## `somesync`

Have you ever wanted a `sync.WaitGroup` with a timeout? Well want no more.

You can use `somesync.TimeoutWaitGroup` in almost the exact same way you use `sync.WaitGroup`; refactors should be pretty easy, because while not literally identical in terms of interface, the interface is textually identical (if you're not using the timeout), making a pretty easy search-and-replace migrate path.

Ref.: [examples/somesync/main.go](examples/somesync/main.go)

```go
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
```

## `somecontext`

Have you ever wanted a helper like `context.WithCancel` that provides some kind of "cancel requested" / "cancel completed" semantic? You're right, that's a bit obscure; but I have, so here it is:

Ref.: [examples/somecontext/main.go](examples/somecontext/main.go)

```go
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
```

## `ctrlc`

May as well button it all up with a Ctrl + C handler:

Ref.: [examples/ctrlc/main.go](examples/ctrlc/main.go)

```go
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
```
