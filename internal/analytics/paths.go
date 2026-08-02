package analytics

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func analyticsPath(directory, filename string) (string, error) {
	if strings.TrimSpace(directory) == "" {
		return "", errors.New("analytics: storage directory is empty")
	}
	absoluteDirectory, err := filepath.Abs(directory)
	if err != nil {
		return "", fmt.Errorf("analytics: resolve storage directory: %w", err)
	}
	if err := os.MkdirAll(absoluteDirectory, 0o700); err != nil {
		return "", fmt.Errorf("analytics: create storage directory: %w", err)
	}
	return filepath.Join(absoluteDirectory, filename), nil
}
