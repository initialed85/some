package ctrlc

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/initialed85/some/somecontext"
	"github.com/stretchr/testify/require"
)

func TestWaitForCtrlC(t *testing.T) {
	ctx, cancel, done, wait := somecontext.WithCancelAndDoneAndWait(context.Background())
	defer cancel()
	defer done()

	waiting := true

	go func() {
		WaitForCtrlC(ctx, cancel, done, wait)
		waiting = false
	}()

	time.Sleep(time.Millisecond * 100)
	require.True(t, waiting)

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	time.Sleep(time.Millisecond * 100)
	require.False(t, waiting)

	<-ctx.Done()
}
