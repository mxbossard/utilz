package utilz

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/flock"
)

func FileLock(fl *flock.Flock, timeout time.Duration) error {
	lockCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	locked, err := fl.TryLockContext(lockCtx, time.Millisecond)
	if err != nil {
		return err
	}
	if !locked {
		err = fmt.Errorf("unable to acquire lock on file: %s", fl.Path())
		if err != nil {
			return err
		}
	}
	return nil
}

func FileLockOrPanic(fl *flock.Flock, timeout time.Duration) {
	err := FileLock(fl, timeout)
	if err != nil {
		panic(err)
	}
}

func FileUnlock(fl *flock.Flock) error {
	err := fl.Unlock()
	return err
}

func FileUnlockOrPanic(fl *flock.Flock) {
	err := FileUnlock(fl)
	if err != nil {
		panic(err)
	}
}
