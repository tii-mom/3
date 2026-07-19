package requestmetrics

import "sync/atomic"

var activeRequests atomic.Int64

func Begin() func() {
	activeRequests.Add(1)
	var finished atomic.Bool
	return func() {
		if finished.CompareAndSwap(false, true) {
			activeRequests.Add(-1)
		}
	}
}

func Active() int64 {
	return activeRequests.Load()
}
