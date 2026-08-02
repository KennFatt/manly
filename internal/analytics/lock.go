package analytics

import (
	"os"
	"path/filepath"
)

const lockFileName = ".analytics.lock"

type fileLock struct {
	file *os.File
}

func lockPath(dataPath string) string {
	return filepath.Join(filepath.Dir(dataPath), lockFileName)
}

func withFileLock(dataPath string, exclusive bool, operation func() error) error {
	lock, err := acquireFileLock(lockPath(dataPath), exclusive)
	if err != nil {
		return err
	}
	operationErr := operation()
	closeErr := lock.close()
	if operationErr != nil {
		return operationErr
	}
	return closeErr
}
