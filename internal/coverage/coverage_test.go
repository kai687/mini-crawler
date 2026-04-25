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
}

func TestTrackerCountsURLsAndRecords(t *testing.T) {
	var tracker Tracker

	tracker.Add(3)
	tracker.Add(0)
	tracker.Add(5)
	tracker.Add(2)

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
}

func TestReportFormat(t *testing.T) {
	report := Report{
		TotalURLs:        4,
		URLsWithRecords:  3,
		TotalRecords:     10,
		Coverage:         0.75,
		AvgRecordsPerURL: 2.5,
	}

	got := report.Format()
	want := "coverage: 3/4 URLs returned records (75.00%); average 2.50 records/URL"

	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}
