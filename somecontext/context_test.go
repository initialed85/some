package somecontext

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWithCancelAndWaitGroup(t *testing.T) {
	ctx, cancel, done, wait := WithCancelAndDoneAndWait(context.Background())
	defer cancel()
	waiting := true
	go func() {
		<-ctx.Done()
		wait()
		waiting = false
	}()

	ctx1, cancel1, done1, wait1 := WithCancelAndDoneAndWait(ctx)
	defer cancel1()
	waiting1 := true
	go func() {
		<-ctx1.Done()
		wait1()
		waiting1 = false
	}()

	ctx2, cancel2, done2, wait2 := WithCancelAndDoneAndWait(ctx)
	defer cancel2()
	waiting2 := true
	go func() {
		<-ctx2.Done()
		wait2()
		waiting2 = false
	}()

	require.True(t, waiting)
	require.True(t, waiting1)
	require.True(t, waiting2)

	cancel2()
	require.True(t, waiting)
	require.True(t, waiting1)
	require.True(t, waiting2)

	done2()
	time.Sleep(time.Millisecond * 100)
	require.True(t, waiting)
	require.True(t, waiting1)
	require.True(t, waiting2)

	cancel1()
	require.True(t, waiting)
	require.True(t, waiting1)
	require.True(t, waiting2)

	done1()
	time.Sleep(time.Millisecond * 100)
	require.True(t, waiting)
	require.True(t, waiting1)
	require.True(t, waiting2)

	cancel()
	require.True(t, waiting)
	require.True(t, waiting1)
	require.True(t, waiting2)

	done()
	time.Sleep(time.Millisecond * 100)
	require.False(t, waiting)
	require.False(t, waiting1)
	require.False(t, waiting2)
}
