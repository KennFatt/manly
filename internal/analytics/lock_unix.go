//go:build darwin || linux

package analytics

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

const (
	fileLockTimeout = 250 * time.Millisecond
	fileLockRetry   = 10 * time.Millisecond
)

func acquireFileLock(path string, exclusive bool) (*fileLock, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	how := syscall.LOCK_SH
	if exclusive {
		how = syscall.LOCK_EX
	}
	deadline := time.Now().Add(fileLockTimeout)
	for {
		err := syscall.Flock(int(file.Fd()), how|syscall.LOCK_NB)
		if err == nil {
			return &fileLock{file: file}, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			_ = file.Close()
			return nil, fmt.Errorf("lock analytics file: %w", err)
		}
		if !time.Now().Before(deadline) {
			_ = file.Close()
			return nil, fmt.Errorf("lock analytics file: timed out after %s", fileLockTimeout)
		}
		time.Sleep(fileLockRetry)
	}
}

func (lock *fileLock) close() error {
	unlockErr := syscall.Flock(int(lock.file.Fd()), syscall.LOCK_UN)
	closeErr := lock.file.Close()
	if unlockErr != nil {
		return fmt.Errorf("unlock analytics file: %w", unlockErr)
	}
	return closeErr
}
