//go:build darwin || linux

package analytics

import (
	"strings"
	"testing"
	"time"
)

func TestFileLockTimesOutWhenContended(t *testing.T) {
	path := lockPath(t.TempDir() + "/analytics.csv")
	first, err := acquireFileLock(path, true)
	if err != nil {
		t.Fatal(err)
	}
	defer first.close()

	started := time.Now()
	_, err = acquireFileLock(path, true)
	elapsed := time.Since(started)
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("contended lock error = %v", err)
	}
	if elapsed < fileLockTimeout || elapsed > 2*time.Second {
		t.Fatalf("lock timeout took %s, want between %s and 2s", elapsed, fileLockTimeout)
	}
}
