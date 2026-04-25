package coverage

import "fmt"

// Report summarizes URL coverage and record density across one crawl.
type Report struct {
	TotalURLs        int     `json:"totalURLs"`
	URLsWithRecords  int     `json:"urlsWithRecords"`
	TotalRecords     int     `json:"totalRecords"`
	Coverage         float64 `json:"coverage"`
	AvgRecordsPerURL float64 `json:"avgRecordsPerURL"`
}

// Tracker accumulates per-URL record counts during a crawl.
type Tracker struct {
	total        int
	withRecords  int
	totalRecords int
}

// Add records that one URL produced recordCount records.
func (t *Tracker) Add(recordCount int) {
	t.total++
	t.totalRecords += recordCount

	if recordCount > 0 {
		t.withRecords++
	}
}

// Report returns the current coverage summary.
func (t *Tracker) Report() Report {
	report := Report{
		TotalURLs:       t.total,
		URLsWithRecords: t.withRecords,
		TotalRecords:    t.totalRecords,
	}

	if t.total > 0 {
		report.Coverage = float64(t.withRecords) / float64(t.total)
		report.AvgRecordsPerURL = float64(t.totalRecords) / float64(t.total)
	}

	return report
}

// Format returns a one-line human-readable summary of the report.
func (r Report) Format() string {
	return fmt.Sprintf(
		"coverage: %d/%d URLs returned records (%.2f%%); average %.2f records/URL",
		r.URLsWithRecords,
		r.TotalURLs,
		r.Coverage*100,
		r.AvgRecordsPerURL,
	)
}
