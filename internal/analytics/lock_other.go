//go:build !darwin && !linux

package analytics

import "errors"

func acquireFileLock(_ string, _ bool) (*fileLock, error) {
	return nil, errors.New("analytics file locking is unsupported on this platform")
}

func (lock *fileLock) close() error {
	return lock.file.Close()
}
