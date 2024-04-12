package somecontext

import (
	"context"
	"time"

	"github.com/initialed85/some/somesync"
)

const defaultTimeout = time.Second * 60

type wgKeyType struct{}

var (
	wgKey = wgKeyType{}
)

// WithCancelAndWaitGroup returns a context with the optionally provided WaitGroup injected, or else injects a new one- it doesn't
// interact with the WaitGroup in any way, for maximum user flexibility
func WithCancelAndWaitGroup(ctx context.Context, wgs ...somesync.WaitGroup) (context.Context, context.CancelFunc, somesync.WaitGroup) {
	lenWgs := len(wgs)

	if lenWgs > 1 {
		panic("somegoutils.context: WithCancelAndDone must have wgs not specified or specific exactly once")
	}

	var wg somesync.WaitGroup

	// if we're given a WaitGroup, use that and put it in the context
	if lenWgs == 1 {
		wg = wgs[0]
		ctx = context.WithValue(ctx, wgKey, wg)
	}

	// otherwise look to see if one already exists in the context
	if lenWgs == 0 {
		rawWg := ctx.Value(wgKey)
		wg, _ = rawWg.(somesync.WaitGroup)
	}

	// failing all, create a new TimeoutWaitGroup and put it in the context
	if wg == nil {
		wg = &somesync.TimeoutWaitGroup{}
		ctx = context.WithValue(ctx, wgKey, wg)
	}

	ctx, cancel := context.WithCancel(ctx)

	return ctx, cancel, wg
}

// WithCancelAndDoneAndWait is a convenience function that returns a context with the optionally provided WaitGroup injected, or else injects a new
// one- it doesn't return a WaitGroup, instead returning a DoneFunc and a WaitFunc (coupled with calling .Add(1) on the WaitGroup for you)
func WithCancelAndDoneAndWait(ctx context.Context, wgs ...somesync.WaitGroup) (context.Context, context.CancelFunc, somesync.DoneFunc, somesync.WaitFunc) {
	ctx, cancel, wg := WithCancelAndWaitGroup(ctx, wgs...)

	wg.Add(1)

	return ctx, cancel, wg.Done, wg.Wait
}

func Cleanup(cancel context.CancelFunc, done somesync.DoneFunc, wait somesync.WaitFunc, timeouts ...time.Duration) {
	if len(timeouts) > 1 {
		panic("somegoutils.context: Cleanup must have timeouts not specified or specific exactly once")
	}

	timeout := defaultTimeout
	if len(timeouts) == 1 {
		timeout = timeouts[0]
	}

	cancel()
	done()
	wait(timeout)
}
