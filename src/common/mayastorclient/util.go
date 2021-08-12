package mayastorclient

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var backOffTimes = []time.Duration{
	5 * time.Second,
	10 * time.Second,
	20 * time.Second,
	40 * time.Second,
	80 * time.Second,
	160 * time.Second,
	240 * time.Second,
}

func niceError(err error) error {
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			// stop huge print out of error on deadline exceeded
			return fmt.Errorf("context deadline exceeded")
		}
	}
	return err
}
