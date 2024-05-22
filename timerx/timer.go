package timerx

import "time"

func StopTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
			// drained timer channel
		default:
			// timer channel was empty
		}
	}
}
