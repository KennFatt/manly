package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	_ "modernc.org/sqlite"
)

const (
	sqliteFilename   = "analytics.db"
	busyTimeoutMS    = 5000
	sqliteTimeFormat = "2006-01-02T15:04:05.000000000Z"
)

var sqliteSchema = []string{
	`CREATE TABLE IF NOT EXISTS concept_load_events (
		id INTEGER PRIMARY KEY,
		occurred_at TEXT NOT NULL,
		batch_id TEXT NOT NULL,
		concept_id TEXT NOT NULL,
		entry_point TEXT NOT NULL CHECK (entry_point IN ('show', 'context'))
	)`,
	`CREATE INDEX IF NOT EXISTS concept_load_events_time_idx ON concept_load_events(occurred_at)`,
	`CREATE INDEX IF NOT EXISTS concept_load_events_batch_idx ON concept_load_events(batch_id)`,
	`CREATE INDEX IF NOT EXISTS concept_load_events_concept_idx ON concept_load_events(concept_id)`,
}

type sqliteProvider struct {
	db   *sql.DB
	path string
}

var _ provider = (*sqliteProvider)(nil)

func newSQLiteProvider(directory string) (*sqliteProvider, error) {
	path, err := analyticsPath(directory, sqliteFilename)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("analytics: open sqlite database: %w", err)
	}
	db.SetMaxOpenConns(1)
	provider := &sqliteProvider{db: db, path: path}
	if err := withFileLock(path, true, func() error {
		ctx := context.Background()
		if _, err := db.ExecContext(ctx, fmt.Sprintf("PRAGMA busy_timeout = %d", busyTimeoutMS)); err != nil {
			return fmt.Errorf("set sqlite busy timeout: %w", err)
		}
		for _, statement := range sqliteSchema {
			if _, err := db.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("initialize sqlite schema: %w", err)
			}
		}
		return nil
	}); err != nil {
		_ = db.Close()
		return nil, err
	}
	return provider, nil
}

func (provider *sqliteProvider) Close() error {
	return provider.db.Close()
}

func (provider *sqliteProvider) RecordConceptLoads(ctx context.Context, loads []ConceptLoad) error {
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
		transaction, err := provider.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin sqlite analytics transaction: %w", err)
		}
		defer transaction.Rollback()

		statement, err := transaction.PrepareContext(ctx, `
			INSERT INTO concept_load_events (occurred_at, batch_id, concept_id, entry_point)
			VALUES (?, ?, ?, ?)
		`)
		if err != nil {
			return fmt.Errorf("prepare sqlite analytics insert: %w", err)
		}
		defer statement.Close()
		for _, load := range loads {
			if _, err := statement.ExecContext(
				ctx,
				load.OccurredAt.UTC().Format(sqliteTimeFormat),
				load.BatchID,
				load.ConceptID,
				load.EntryPoint,
			); err != nil {
				return fmt.Errorf("insert sqlite analytics event: %w", err)
			}
		}
		if err := transaction.Commit(); err != nil {
			return fmt.Errorf("commit sqlite analytics transaction: %w", err)
		}
		return nil
	})
}

func (provider *sqliteProvider) Report(ctx context.Context, options ReportOptions) (Report, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	report := emptyReport(options)
	err := withFileLock(provider.path, false, func() error {
		where, arguments := sqliteFilter(options.Since)
		var conceptLoads, retrievalBatches int64
		totalQuery := `SELECT COUNT(*), COUNT(DISTINCT batch_id) FROM concept_load_events` + where
		if err := provider.db.QueryRowContext(ctx, totalQuery, arguments...).Scan(&conceptLoads, &retrievalBatches); err != nil {
			return fmt.Errorf("read sqlite analytics totals: %w", err)
		}
		report.ConceptLoads = int(conceptLoads)
		report.RetrievalBatches = int(retrievalBatches)
		if retrievalBatches > 0 {
			report.AverageConceptsPerBatch = float64(conceptLoads) / float64(retrievalBatches)
		}

		entryQuery := `SELECT entry_point, COUNT(*) FROM concept_load_events` + where + ` GROUP BY entry_point ORDER BY entry_point`
		entryRows, err := provider.db.QueryContext(ctx, entryQuery, arguments...)
		if err != nil {
			return fmt.Errorf("read sqlite analytics entry points: %w", err)
		}
		for entryRows.Next() {
			var entryPoint string
			var count int64
			if err := entryRows.Scan(&entryPoint, &count); err != nil {
				_ = entryRows.Close()
				return fmt.Errorf("scan sqlite analytics entry point: %w", err)
			}
			report.EntryPoints[EntryPoint(entryPoint)] = int(count)
		}
		if err := entryRows.Err(); err != nil {
			_ = entryRows.Close()
			return fmt.Errorf("iterate sqlite analytics entry points: %w", err)
		}
		if err := entryRows.Close(); err != nil {
			return fmt.Errorf("close sqlite analytics entry points: %w", err)
		}

		topQuery := `SELECT concept_id, COUNT(*) AS load_count FROM concept_load_events` + where + ` GROUP BY concept_id ORDER BY load_count DESC, concept_id ASC`
		topArguments := append([]any(nil), arguments...)
		if options.Limit > 0 {
			topQuery += ` LIMIT ?`
			topArguments = append(topArguments, options.Limit)
		}
		topRows, err := provider.db.QueryContext(ctx, topQuery, topArguments...)
		if err != nil {
			return fmt.Errorf("read sqlite analytics top concepts: %w", err)
		}
		for topRows.Next() {
			var conceptID string
			var count int64
			if err := topRows.Scan(&conceptID, &count); err != nil {
				_ = topRows.Close()
				return fmt.Errorf("scan sqlite analytics top concept: %w", err)
			}
			report.TopConcepts = append(report.TopConcepts, ConceptUsage{ConceptID: conceptID, LoadCount: int(count)})
		}
		if err := topRows.Err(); err != nil {
			_ = topRows.Close()
			return fmt.Errorf("iterate sqlite analytics top concepts: %w", err)
		}
		if err := topRows.Close(); err != nil {
			return fmt.Errorf("close sqlite analytics top concepts: %w", err)
		}

		batches, err := provider.queryBatches(ctx, options.Since)
		if err != nil {
			return err
		}
		report.RecentBatches = batches
		if options.Limit > 0 && len(report.RecentBatches) > options.Limit {
			report.RecentBatches = report.RecentBatches[:options.Limit]
		}
		return nil
	})
	return report, err
}

func (provider *sqliteProvider) queryBatches(ctx context.Context, since *time.Time) ([]RetrievalBatch, error) {
	where, arguments := sqliteFilter(since)
	query := `SELECT batch_id, occurred_at, concept_id, entry_point FROM concept_load_events` + where + ` ORDER BY occurred_at DESC, batch_id ASC, id ASC`
	rows, err := provider.db.QueryContext(ctx, query, arguments...)
	if err != nil {
		return nil, fmt.Errorf("read sqlite analytics batches: %w", err)
	}
	defer rows.Close()

	batches := make(map[string]*batchReport)
	for rows.Next() {
		var batchID, occurredAtValue, conceptID, entryPointValue string
		if err := rows.Scan(&batchID, &occurredAtValue, &conceptID, &entryPointValue); err != nil {
			return nil, fmt.Errorf("scan sqlite analytics batch: %w", err)
		}
		occurredAt, err := time.Parse(time.RFC3339Nano, occurredAtValue)
		if err != nil {
			return nil, fmt.Errorf("parse sqlite analytics timestamp %q: %w", occurredAtValue, err)
		}
		load := ConceptLoad{
			OccurredAt: occurredAt,
			BatchID:    batchID,
			ConceptID:  conceptID,
			EntryPoint: EntryPoint(entryPointValue),
		}
		if err := validateLoad(load); err != nil {
			return nil, err
		}
		batch, exists := batches[batchID]
		if !exists {
			batch = &batchReport{
				batchID:    batchID,
				occurredAt: occurredAt.UTC(),
				entryPoint: load.EntryPoint,
				conceptIDs: make(map[string]struct{}),
			}
			batches[batchID] = batch
		}
		if batch.entryPoint != load.EntryPoint {
			return nil, fmt.Errorf("analytics: batch %q contains multiple entry points", batchID)
		}
		if occurredAt.After(batch.occurredAt) {
			batch.occurredAt = occurredAt.UTC()
		}
		batch.conceptIDs[conceptID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sqlite analytics batches: %w", err)
	}

	result := make([]RetrievalBatch, 0, len(batches))
	for _, batch := range batches {
		conceptIDs := make([]string, 0, len(batch.conceptIDs))
		for conceptID := range batch.conceptIDs {
			conceptIDs = append(conceptIDs, conceptID)
		}
		sort.Strings(conceptIDs)
		result = append(result, RetrievalBatch{
			BatchID:    batch.batchID,
			OccurredAt: batch.occurredAt,
			EntryPoint: batch.entryPoint,
			ConceptIDs: conceptIDs,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if !result[i].OccurredAt.Equal(result[j].OccurredAt) {
			return result[i].OccurredAt.After(result[j].OccurredAt)
		}
		return result[i].BatchID < result[j].BatchID
	})
	return result, nil
}

func sqliteFilter(since *time.Time) (string, []any) {
	if since == nil {
		return "", nil
	}
	return " WHERE occurred_at >= ?", []any{since.UTC().Format(sqliteTimeFormat)}
}
