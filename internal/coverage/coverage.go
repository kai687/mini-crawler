package coverage

import "fmt"

// Report summarizes URL coverage and record density across one crawl.
type Report struct {
	TotalURLs         int     `json:"totalURLs"`
	URLsWithRecords   int     `json:"urlsWithRecords"`
	TotalRecords      int     `json:"totalRecords"`
	Coverage          float64 `json:"coverage"`
	AvgRecordsPerURL  float64 `json:"avgRecordsPerURL"`
	ExpectedHeadings  int     `json:"expectedHeadings"`
	ExtractedHeadings int     `json:"extractedHeadings"`
	HeadingCoverage   float64 `json:"headingCoverage"`
}

// Tracker accumulates per-URL record counts during a crawl.
type Tracker struct {
	total             int
	withRecords       int
	totalRecords      int
	expectedHeadings  int
	extractedHeadings int
}

// Add records that one URL produced recordCount records, with expected/extracted heading counts.
func (t *Tracker) Add(recordCount, expectedHeadings, extractedHeadings int) {
	t.total++
	t.totalRecords += recordCount
	t.expectedHeadings += expectedHeadings
	t.extractedHeadings += extractedHeadings

	if recordCount > 0 {
		t.withRecords++
	}
}

// Report returns the current coverage summary.
func (t *Tracker) Report() Report {
	report := Report{
		TotalURLs:         t.total,
		URLsWithRecords:   t.withRecords,
		TotalRecords:      t.totalRecords,
		ExpectedHeadings:  t.expectedHeadings,
		ExtractedHeadings: t.extractedHeadings,
	}

	if t.total > 0 {
		report.Coverage = float64(t.withRecords) / float64(t.total)
		report.AvgRecordsPerURL = float64(t.totalRecords) / float64(t.total)
	}

	if t.expectedHeadings > 0 {
		report.HeadingCoverage = float64(t.extractedHeadings) / float64(t.expectedHeadings)
	} else {
		report.HeadingCoverage = 1
	}

	return report
}

// Format returns a one-line human-readable summary of the report.
func (r Report) Format() string {
	return fmt.Sprintf(
		"coverage: %d/%d URLs returned records (%.2f%%); %d records total; "+
			"average %.2f records/URL; headings %d/%d (%.2f%%)",
		r.URLsWithRecords,
		r.TotalURLs,
		r.Coverage*100,
		r.TotalRecords,
		r.AvgRecordsPerURL,
		r.ExtractedHeadings,
		r.ExpectedHeadings,
		r.HeadingCoverage*100,
	)
}
