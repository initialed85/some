package ctrlc

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/initialed85/some/somecontext"
	"github.com/initialed85/some/somesync"
)

var noopCancelOrDone = func() {}

var noopWait = func(...time.Duration) {}

func WaitForCtrlC(ctx context.Context, cancel context.CancelFunc, done somesync.DoneFunc, wait somesync.WaitFunc) {
	if ctx == nil {
		ctx = context.Background()
	}

	if cancel == nil {
		cancel = noopCancelOrDone
	}

	if done == nil {
		done = noopCancelOrDone
	}

	if wait == nil {
		wait = noopWait
	}

	defer somecontext.Cleanup(cancel, done, wait)

	// 128 in case Ctrl + C gets spammed
	c := make(chan os.Signal, 128)

	signal.Notify(c, syscall.SIGINT)

	// block until the context is cancelled or until we get the Ctrl + C signal
	select {
	case <-ctx.Done():
		return
	case <-c:
		return
	}
}
