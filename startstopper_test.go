package startstopper

import (
	"math/rand"
	"testing"
	"time"
)

// TestStartStopper_sync ascertains that in a simple synchronous setting,
// StartStopper exhibits a causally consistent state. It says it's stopped
// after being told to stop, and says it's started after being told to start.
func TestStartStopper_sync(t *testing.T) {

	// Phase 1: New
	ss := StartStopper{}
	select {
	default:
		// OK
	case <-ss.Stopped():
		t.Fatalf("NewStartStopper returned a stopped StartStopper:")
	}
	if ss.IsStopped() {
		t.Fatalf("Stopped() returned true before Stop called.")
	}

	// Phase 2: Stopped
	ss.Stop()
	if !ss.IsStopped() {
		t.Fatalf("Stopped() returned false after Stop called.")
	}
	select {
	default:
		t.Fatalf("Stopped() blocked after Stop() called.")
	case <-ss.Stopped():
		// OK
	}

	// Phase 3: Restarted
	ss.Start()
	if ss.IsStopped() {
		t.Fatalf("Stopped() returned true after Start called.")
	}
	select {
	default:
		// OK
	case <-ss.Stopped():
		t.Fatalf("Stopped() did not block after Start() called.")
	}

	// Phase 4: Stopped again
	ss.Stop()
	if !ss.IsStopped() {
		t.Fatalf("Stopped() returned false after Stop called.")
	}
	select {
	default:
		t.Fatalf("Stopped() blocked after Stop() called.")
	case <-ss.Stopped():
		// OK
	}
}

// TestStartStopper_async tries to trip the race detector by stressing a
// StartStopper with arbitrary calls. It then checks the results look
// statistically reasonable within a very wide margin of error.
func TestStartStopper_async(t *testing.T) {

	ss := StartStopper{}

	// done is just used to bring down the goroutines below after the test.
	done := make(chan struct{})

	// Let's set 1000 goroutines randomly shouting at a StartStopper.
	for i := 0; i < 1000; i++ {
		go func() {
			for {
				time.Sleep(time.Duration(rand.Intn(10)) * time.Microsecond)
				ss.Start()
				time.Sleep(time.Duration(rand.Intn(10)) * time.Microsecond)
				ss.Stop()
				select {
				default:
				case <-done:
					return
				}
			}
		}()
	}

	starteds, stoppeds := 0, 0
	for i := 0; i < 1000; i++ {
		time.Sleep(time.Duration(rand.Intn(10)) * time.Microsecond)
		select {
		case <-ss.Stopped():
			stoppeds++
		default:
			starteds++
		}
	}

	close(done)

	t.Logf("StartStopper was started %d times and stopped %d times.",
		starteds, stoppeds)

	if starteds == 0 {
		t.Fatalf("In 1000 random samples, StartStopper was never started.")
	}
	if stoppeds == 0 {
		t.Fatalf("In 1000 random samples, StartStopper was never stopped.")
	}

}

func BenchmarkStartStopper(b *testing.B) {

	ss := StartStopper{}
	for i := 0; i < b.N; i++ {
		ss.Stop()
		ss.Start()
	}

}
