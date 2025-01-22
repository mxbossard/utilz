package errorz

import (
	"errors"
	"fmt"
	"time"
)

type timeout struct {
	Duration time.Duration
	Message  string
}

func (e timeout) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("timeout after %s", e.Duration)
	}
	return fmt.Sprintf("timeout after %s while: %s", e.Duration, e.Message)
}

func Timeout(d time.Duration, msg string) timeout {
	return timeout{d, msg}
}

func Timeoutf(d time.Duration, format string, a any) timeout {
	return timeout{d, fmt.Sprintf(format, a)}
}

func IsTimeout(e error) bool {
	return errors.As(e, &timeout{})
}
