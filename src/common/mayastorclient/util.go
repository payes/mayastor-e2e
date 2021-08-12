package mayastorclient

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func niceError(err error) error {
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			// stop huge print out of error on deadline exceeded
			return fmt.Errorf("context deadline exceeded")
		}
	}
	return err
}

var backOffTimes = []time.Duration{
	5 * time.Second,
	10 * time.Second,
	20 * time.Second,
	40 * time.Second,
	80 * time.Second,
	160 * time.Second,
	240 * time.Second,
}

// retryBackoff iterator over backoffTimes, calls function and if the error is deadline_exceeded sleeps for the
// iteration time. effectively a do-while loop
// limitation: the set of back off times is fixed.
func retryBackoff(f func() (err error)) {
	for i := 0; i < len(backOffTimes); i++ {
		if !errors.Is(f(), context.DeadlineExceeded) {
			return
		}
		time.Sleep(backOffTimes[i])
	}
}
