package analytics

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

const csvFilename = "analytics.csv"

var csvHeader = []string{"occurred_at", "batch_id", "concept_id", "entry_point"}

type csvProvider struct {
	path string
}

var _ provider = (*csvProvider)(nil)

func newCSVProvider(directory string) (*csvProvider, error) {
	path, err := analyticsPath(directory, csvFilename)
	if err != nil {
		return nil, err
	}
	return &csvProvider{path: path}, nil
}

func (provider *csvProvider) RecordConceptLoads(ctx context.Context, loads []ConceptLoad) error {
	if len(loads) == 0 {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	for _, load := range loads {
		if err := validateLoad(load); err != nil {
			return err
		}
	}
	return withFileLock(provider.path, true, func() error {
		file, err := os.OpenFile(provider.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return fmt.Errorf("open csv analytics file: %w", err)
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			return fmt.Errorf("stat csv analytics file: %w", err)
		}
		var buffer bytes.Buffer
		writer := csv.NewWriter(&buffer)
		if info.Size() == 0 {
			if err := writer.Write(csvHeader); err != nil {
				return fmt.Errorf("encode csv analytics header: %w", err)
			}
		}
		for _, load := range loads {
			if err := writer.Write([]string{
				load.OccurredAt.UTC().Format(time.RFC3339Nano),
				load.BatchID,
				load.ConceptID,
				string(load.EntryPoint),
			}); err != nil {
				return fmt.Errorf("encode csv analytics event: %w", err)
			}
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			return fmt.Errorf("flush csv analytics records: %w", err)
		}
		data := buffer.Bytes()
		written, err := file.Write(data)
		if err != nil {
			return fmt.Errorf("append csv analytics records: %w", err)
		}
		if written != len(data) {
			return io.ErrShortWrite
		}
		if err := file.Sync(); err != nil {
			return fmt.Errorf("sync csv analytics records: %w", err)
		}
		return nil
	})
}

func (provider *csvProvider) Report(ctx context.Context, options ReportOptions) (Report, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	report := emptyReport(options)
	err := withFileLock(provider.path, false, func() error {
		events, err := provider.readEvents(ctx)
		if err != nil {
			return err
		}
		report, err = aggregateEvents(events, options)
		return err
	})
	return report, err
}

func (provider *csvProvider) readEvents(ctx context.Context) ([]ConceptLoad, error) {
	file, err := os.Open(provider.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open csv analytics file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if errors.Is(err, io.EOF) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read csv analytics header: %w", err)
	}
	if !sameStrings(header, csvHeader) {
		return nil, fmt.Errorf("csv analytics header = %v, want %v", header, csvHeader)
	}

	events := make([]ConceptLoad, 0)
	for recordNumber := 2; ; recordNumber++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read csv analytics record %d: %w", recordNumber, err)
		}
		if len(record) != len(csvHeader) {
			return nil, fmt.Errorf("csv analytics record %d has %d fields, want %d", recordNumber, len(record), len(csvHeader))
		}
		occurredAt, err := time.Parse(time.RFC3339Nano, record[0])
		if err != nil {
			return nil, fmt.Errorf("parse csv analytics timestamp on record %d: %w", recordNumber, err)
		}
		load := ConceptLoad{
			OccurredAt: occurredAt,
			BatchID:    record[1],
			ConceptID:  record[2],
			EntryPoint: EntryPoint(record[3]),
		}
		if err := validateLoad(load); err != nil {
			return nil, fmt.Errorf("csv analytics record %d: %w", recordNumber, err)
		}
		events = append(events, load)
	}
	return events, nil
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
