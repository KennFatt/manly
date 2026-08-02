package analytics

import (
	"context"
	"fmt"
	"strings"
)

// Settings configures the local analytics service.
type Settings struct {
	Enabled   bool
	Provider  Provider
	Directory string
}

type provider interface {
	Recorder
	Reader
}

// Service isolates analytics persistence from retrieval commands.
type Service struct {
	settings  Settings
	backend   provider
	initError error
}

var _ Recorder = (*Service)(nil)
var _ Reader = (*Service)(nil)

// New creates a lazy analytics service. Provider initialization is deferred
// until recording or reporting so unrelated commands do not create storage.
func New(settings Settings) *Service {
	settings.Provider = Provider(strings.ToLower(strings.TrimSpace(string(settings.Provider))))
	return &Service{settings: settings}
}

// RecordConceptLoads records events when analytics is enabled.
func (service *Service) RecordConceptLoads(ctx context.Context, loads []ConceptLoad) error {
	if !service.settings.Enabled || len(loads) == 0 {
		return nil
	}
	backend, err := service.getBackend()
	if err != nil {
		return err
	}
	return backend.RecordConceptLoads(ctx, loads)
}

// Report reads analytics when enabled. Initialization and storage errors are
// returned because reporting is the primary purpose of this command.
func (service *Service) Report(ctx context.Context, options ReportOptions) (Report, error) {
	report := emptyReport(options)
	report.Enabled = service.settings.Enabled
	report.Provider = service.settings.Provider
	if !service.settings.Enabled {
		return report, nil
	}

	backend, err := service.getBackend()
	if err != nil {
		return report, err
	}
	result, err := backend.Report(ctx, options)
	if err != nil {
		return report, err
	}
	result.Enabled = true
	result.Provider = service.settings.Provider
	result.Since = normalizedSince(options.Since)
	return result, nil
}

// Close releases provider resources when they were initialized.
func (service *Service) Close() error {
	closer, ok := service.backend.(interface{ Close() error })
	if !ok {
		return nil
	}
	return closer.Close()
}

func (service *Service) getBackend() (provider, error) {
	if service.backend != nil {
		return service.backend, nil
	}
	if service.initError != nil {
		return nil, service.initError
	}

	var backend provider
	var err error
	switch service.settings.Provider {
	case ProviderSQLite:
		backend, err = newSQLiteProvider(service.settings.Directory)
	case ProviderCSV:
		backend, err = newCSVProvider(service.settings.Directory)
	default:
		err = fmt.Errorf("analytics: unsupported provider %q", service.settings.Provider)
	}
	if err != nil {
		service.initError = err
		return nil, err
	}
	service.backend = backend
	return backend, nil
}
