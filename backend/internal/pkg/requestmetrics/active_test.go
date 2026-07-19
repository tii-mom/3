package requestmetrics

import (
	"sync"
	"testing"
)

func TestBeginTracksConcurrentRequestsAndFinishesOnce(t *testing.T) {
	baseline := Active()
	const count = 32
	finishes := make([]func(), count)

	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			finishes[index] = Begin()
		}(i)
	}
	wg.Wait()

	if got := Active(); got != baseline+count {
		t.Fatalf("Active() = %d, want %d", got, baseline+count)
	}
	for _, finish := range finishes {
		finish()
		finish()
	}
	if got := Active(); got != baseline {
		t.Fatalf("Active() after finish = %d, want %d", got, baseline)
	}
}
