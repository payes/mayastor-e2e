package mayastorclient

import (
	"context"
	"errors"
	"fmt"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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

const ctxTimeout = 5 * time.Second

// retryBackoff iterator over backoffTimes, calls function and if the error is deadline_exceeded sleeps for the
// iteration time. effectively a do-while loop
// limitation: the set of back off times is fixed.
func retryBackoff(f func() (err error)) {
	timeout := ctxTimeout
	for i := 0; i < 6; i++ {
		if !errors.Is(f(), context.DeadlineExceeded) {
			return
		}
		logf.Log.Info("retrying gRPC call", "after", timeout)
		time.Sleep(timeout)
		timeout *= 2
	}
}
