package somesync

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeoutWaitGroup(t *testing.T) {
	t.Run("Reference", func(t *testing.T) {
		wg := TimeoutWaitGroup{}

		// a zero wait group should fall through on wait
		wg.Wait()

		// a zero wait group should panic on negative decrement
		require.Panics(t, func() {
			wg.Done()
		})

		//
		// standard behaviour as per sync.WaitGroup
		//

		wg.Add(1)

		waiting := true
		go func() {
			wg.Wait()
			waiting = false
		}()
		time.Sleep(time.Millisecond * 100)
		require.True(t, waiting)

		wg.Done()
		time.Sleep(time.Millisecond * 100)
		require.False(t, waiting)

		//
		// timeout behaviour
		//

		wg.Add(1)

		waiting = true
		go func() {
			wg.Wait(time.Millisecond * 200)
			waiting = false
		}()
		time.Sleep(time.Millisecond * 100)
		require.True(t, waiting)

		time.Sleep(time.Millisecond * 222)
		require.False(t, waiting)

		wg.Done()
		require.False(t, waiting)
	})

	t.Run("ReferenceConcurrency", func(t *testing.T) {
		// no dogfood here!
		step := new(sync.WaitGroup)
		task := new(sync.WaitGroup)

		wg := TimeoutWaitGroup{}

		wg.Add(1)

		waiting := true
		go func() {
			wg.Wait()
			waiting = false
		}()
		time.Sleep(time.Millisecond * 100)
		require.True(t, waiting)

		step.Add(1)
		for i := 0; i < 1000; i++ {
			task.Add(1)
			go func() {
				defer task.Done()
				step.Wait()

				// a zero wait group should fall through on wait
				wg.Add(1)
				wg.Done()
			}()
		}
		step.Done()
		task.Wait()

		require.True(t, waiting)

		wg.Done()
		time.Sleep(time.Millisecond * 100)

		require.False(t, waiting)
	})

	t.Run("ReferencePerformanceMainThread", func(t *testing.T) {
		wg := TimeoutWaitGroup{}

		for i := 0; i < 1000000; i++ {
			wg.Add(1)
		}

		for i := 0; i < 1000000; i++ {
			wg.Done()
		}

		wg.Wait()
	})

	t.Run("ReferencePerformanceGoroutines", func(t *testing.T) {
		wg := TimeoutWaitGroup{}

		for i := 0; i < 1000000; i++ {
			go func() {
				wg.Add(1)
			}()
		}

		for i := 0; i < 1000000; i++ {
			go func() {
				wg.Done()
			}()
		}

		time.Sleep(time.Second * 1)

		wg.Wait()
	})

	t.Run("PointerConcurrency", func(t *testing.T) {
		// no dogfood here!
		step := new(sync.WaitGroup)
		task := new(sync.WaitGroup)

		wg := new(TimeoutWaitGroup)

		wg.Add(1)

		waiting := true
		go func() {
			wg.Wait()
			waiting = false
		}()
		time.Sleep(time.Millisecond * 100)
		require.True(t, waiting)

		step.Add(1)
		for i := 0; i < 1000; i++ {
			task.Add(1)
			go func() {
				defer task.Done()
				step.Wait()

				// a zero wait group should fall through on wait
				wg.Add(1)
				wg.Done()
			}()
		}
		step.Done()
		task.Wait()

		require.True(t, waiting)

		wg.Done()
		runtime.Gosched()
		time.Sleep(time.Millisecond * 100)

		require.False(t, waiting)
	})

	t.Run("PointerPerformanceMainThread", func(t *testing.T) {
		wg := new(TimeoutWaitGroup)

		for i := 0; i < 1000000; i++ {
			wg.Add(1)
		}

		for i := 0; i < 1000000; i++ {
			wg.Done()
		}

		wg.Wait()
	})

	t.Run("PointerPerformanceGoroutines", func(t *testing.T) {
		wg := new(TimeoutWaitGroup)

		for i := 0; i < 1000000; i++ {
			go func() {
				wg.Add(1)
			}()
		}

		for i := 0; i < 1000000; i++ {
			go func() {
				wg.Done()
			}()
		}

		time.Sleep(time.Second * 1)

		wg.Wait()
	})
}
