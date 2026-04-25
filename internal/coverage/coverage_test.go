package coverage

import "testing"

func TestTrackerEmpty(t *testing.T) {
	var tracker Tracker

	report := tracker.Report()
	if report.TotalURLs != 0 {
		t.Fatalf("TotalURLs = %d, want 0", report.TotalURLs)
	}

	if report.Coverage != 0 {
		t.Fatalf("Coverage = %v, want 0", report.Coverage)
	}

	if report.AvgRecordsPerURL != 0 {
		t.Fatalf("AvgRecordsPerURL = %v, want 0", report.AvgRecordsPerURL)
	}

	if report.HeadingCoverage != 1 {
		t.Fatalf("HeadingCoverage = %v, want 1 (no expected headings)", report.HeadingCoverage)
	}
}

func TestTrackerCountsURLsAndRecords(t *testing.T) {
	var tracker Tracker

	tracker.Add(3, 4, 3)
	tracker.Add(0, 0, 0)
	tracker.Add(5, 2, 2)
	tracker.Add(2, 4, 3)

	report := tracker.Report()
	if report.TotalURLs != 4 {
		t.Fatalf("TotalURLs = %d, want 4", report.TotalURLs)
	}

	if report.URLsWithRecords != 3 {
		t.Fatalf("URLsWithRecords = %d, want 3", report.URLsWithRecords)
	}

	if report.TotalRecords != 10 {
		t.Fatalf("TotalRecords = %d, want 10", report.TotalRecords)
	}

	if report.Coverage != 0.75 {
		t.Fatalf("Coverage = %v, want 0.75", report.Coverage)
	}

	if report.AvgRecordsPerURL != 2.5 {
		t.Fatalf("AvgRecordsPerURL = %v, want 2.5", report.AvgRecordsPerURL)
	}

	if report.ExpectedHeadings != 10 {
		t.Fatalf("ExpectedHeadings = %d, want 10", report.ExpectedHeadings)
	}

	if report.ExtractedHeadings != 8 {
		t.Fatalf("ExtractedHeadings = %d, want 8", report.ExtractedHeadings)
	}

	if report.HeadingCoverage != 0.8 {
		t.Fatalf("HeadingCoverage = %v, want 0.8", report.HeadingCoverage)
	}
}

func TestReportFormat(t *testing.T) {
	report := Report{
		TotalURLs:         4,
		URLsWithRecords:   3,
		TotalRecords:      10,
		Coverage:          0.75,
		AvgRecordsPerURL:  2.5,
		ExpectedHeadings:  10,
		ExtractedHeadings: 8,
		HeadingCoverage:   0.8,
	}

	got := report.Format()
	want := "coverage: 3/4 URLs returned records (75.00%); 10 records total; " +
		"average 2.50 records/URL; headings 8/10 (80.00%)"

	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}
