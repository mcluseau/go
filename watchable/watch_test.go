package watchable

import (
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	wable := New[int]()

	wable.Set(1)

	expired := false
	go func() {
		time.Sleep(2 * time.Millisecond)
		expired = true
		wable.Close()
	}()

	t.Logf("%+v", wable)

	w := wable.NewWatch()

	hadV := false
	loopCount := 0
	for {
		loopCount++

		v, ok, timedOut := w.NextWithTimeout(4 * time.Millisecond)
		if timedOut {
			t.Error("timed out")
		}
		if !ok {
			break
		}

		hadV = true
		if v != 1 {
			t.Error("v should be 1")
		}
	}

	if loopCount != 2 {
		t.Errorf("expired only 2 loops, got %d", loopCount)
	}

	if !hadV {
		t.Error("got no value")
	} else if !expired {
		t.Error("exited before expiry")
	}
}
