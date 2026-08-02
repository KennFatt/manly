package analytics

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

const batchIDSize = 16

// NewBatch creates one timestamped, deduplicated batch of concept loads.
func NewBatch(entryPoint EntryPoint, conceptIDs []string) ([]ConceptLoad, error) {
	if !validEntryPoint(entryPoint) {
		return nil, fmt.Errorf("analytics: unsupported entry point %q", entryPoint)
	}
	if len(conceptIDs) == 0 {
		return nil, nil
	}

	ids := make([]string, 0, len(conceptIDs))
	seen := make(map[string]struct{}, len(conceptIDs))
	for _, conceptID := range conceptIDs {
		if strings.TrimSpace(conceptID) == "" {
			return nil, errors.New("analytics: concept ID must not be empty")
		}
		if _, exists := seen[conceptID]; exists {
			continue
		}
		seen[conceptID] = struct{}{}
		ids = append(ids, conceptID)
	}

	batchID, err := newBatchID()
	if err != nil {
		return nil, fmt.Errorf("analytics: create batch ID: %w", err)
	}
	occurredAt := time.Now().UTC()
	loads := make([]ConceptLoad, 0, len(ids))
	for _, conceptID := range ids {
		loads = append(loads, ConceptLoad{
			OccurredAt: occurredAt,
			BatchID:    batchID,
			ConceptID:  conceptID,
			EntryPoint: entryPoint,
		})
	}
	return loads, nil
}

func newBatchID() (string, error) {
	value := make([]byte, batchIDSize)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func validEntryPoint(entryPoint EntryPoint) bool {
	return entryPoint == EntryPointShow || entryPoint == EntryPointContext
}
