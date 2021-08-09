package mayastorclient

import (
	"context"
	"errors"
	"fmt"
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
