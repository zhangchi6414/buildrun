package utils

import (
	"fmt"
	"time"
)

type TimeoutError struct {
	after   time.Duration
	message string
}

// Error implements the Go error interface.
func (t *TimeoutError) Error() string {
	if len(t.message) > 0 {
		return fmt.Sprintf("%s timed out after %v", t.message, t.after)
	}
	return fmt.Sprintf("function timed out after %v", t.after)
}

func TimeoutAfter(t time.Duration, errorMsg string, f func(*time.Timer) error) error {
	c := make(chan error, 1)
	timer := time.NewTimer(t)
	go func() {
		err := f(timer)
		if !IsTimeoutError(err) {
			c <- err
		}
	}()
	select {
	case err := <-c:
		timer.Stop()
		return err
	case <-timer.C:
		return &TimeoutError{after: t, message: errorMsg}
	}
}
func IsTimeoutError(e error) bool {
	_, ok := e.(*TimeoutError)
	return ok
}
