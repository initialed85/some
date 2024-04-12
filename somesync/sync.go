package somesync

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

var empty = struct{}{}

type DoneFunc func()

type WaitFunc func(...time.Duration)

type WaitGroup interface {
	Add(delta int)
	Done()
	Wait(timeouts ...time.Duration)
}

type TimeoutWaitGroup struct {
	constructed atomic.Bool
	mu          sync.Mutex
	counter     int
	waiterById  map[uuid.UUID]chan struct{}
}

// constructIfRequired means we can provide the same construction interface
// as sync.WaitGroup i.e. wg := sync.WaitGroup{} or wg := new(sync.WaitGroup)
func (wg *TimeoutWaitGroup) constructIfRequired() {
	// this helps us to quickly drop out of this function after construction
	if wg.constructed.Load() {
		return
	}

	// this avoids races for concurrent calls when we're not yet constructed
	if !wg.mu.TryLock() {
		return
	}
	defer wg.mu.Unlock()

	wg.constructed.Store(true)

	wg.waiterById = make(map[uuid.UUID]chan struct{})
}

// Add increments the wait counter by the given amount
func (wg *TimeoutWaitGroup) Add(delta int) {
	wg.constructIfRequired()

	wg.mu.Lock()
	defer wg.mu.Unlock()

	wg.counter++
}

// Done decrements the wait counter
func (wg *TimeoutWaitGroup) Done() {
	wg.constructIfRequired()

	wg.mu.Lock()
	defer wg.mu.Unlock()

	if wg.counter == 0 {
		panic("timeoutwaitgroup: negative TimeoutWaitGroup counter")
	}

	wg.counter--

	if wg.counter > 0 {
		return
	}

	for _, waiter := range wg.waiterById {
		waiter <- empty
	}
}

// Wait waits for the wait counter to be 0 and may be given an optional timeout; the variadic arguments
// pattern is used to provide a calling interface (not a literal interface) that matches sync.WaitGroup
// but also permits being given a timeout
func (wg *TimeoutWaitGroup) Wait(timeouts ...time.Duration) {
	wg.constructIfRequired()

	if len(timeouts) > 1 {
		panic("somegoutils.sync: TimeoutWaitGroup.Wait must have timeouts not specified or specific exactly once")
	}

	wg.mu.Lock()

	if wg.counter == 0 {
		wg.mu.Unlock()
		return
	}

	id := uuid.Must(uuid.NewRandom())

	// a buffered channel ensures that the final call to .Done() won't block as it loads all
	// the waiters
	waiter := make(chan struct{}, 1)

	wg.waiterById[id] = waiter

	wg.mu.Unlock()

	defer func() {
		wg.mu.Lock()
		delete(wg.waiterById, id)
		wg.mu.Unlock()
	}()

	if len(timeouts) == 1 {
		select {
		case <-time.After(timeouts[0]):
			return
		case <-waiter:
			return
		}
	}

	<-waiter
}
