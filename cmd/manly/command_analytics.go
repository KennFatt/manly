package main

import (
	"context"
	"os"
	"time"

	"github.com/KennFatt/manly/internal/analytics"
)

// AnalyticsCommand reports local concept retrieval usage.
type AnalyticsCommand struct {
	Limit  int    `default:"10" help:"Maximum top concepts and recent batches."`
	Since  string `help:"Only include events within a duration such as 24h or 7d."`
	Format string `default:"${format}" help:"Output format."`
}

func (command *AnalyticsCommand) Run(app *appContext) error {
	format, err := parseFormat(command.Format)
	if err != nil {
		return err
	}
	since, err := analytics.ParseSince(command.Since, time.Now())
	if err != nil {
		return err
	}
	if app.analyticsReader == nil {
		return renderAnalytics(os.Stdout, analytics.Report{Enabled: false}, format)
	}
	report, err := app.analyticsReader.Report(context.Background(), analytics.ReportOptions{
		Since: since,
		Limit: command.Limit,
	})
	if err != nil {
		return err
	}
	return renderAnalytics(os.Stdout, report, format)
}
