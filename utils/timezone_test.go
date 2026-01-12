package utils

import (
	"testing"
	"time"
)

func TestTimezoneHelper_TodayRange(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
	}{
		{"UTC", "UTC"},
		{"Asia/Shanghai", "Asia/Shanghai"},
		{"America/New_York", "America/New_York"},
		{"Europe/London", "Europe/London"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, err := time.LoadLocation(tt.timezone)
			if err != nil {
				t.Fatalf("Failed to load timezone %s: %v", tt.timezone, err)
			}

			helper := NewTimezoneHelper(location)
			start, end := helper.TodayRange()

			// Verify start is at beginning of day
			if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
				t.Errorf("Start time should be at 00:00:00, got %02d:%02d:%02d",
					start.Hour(), start.Minute(), start.Second())
			}

			// Verify end is exactly 24 hours after start
			expectedEnd := start.Add(24 * time.Hour)
			if !end.Equal(expectedEnd) {
				t.Errorf("End time should be exactly 24 hours after start")
			}

			// Verify end is at beginning of next day
			if end.Hour() != 0 || end.Minute() != 0 || end.Second() != 0 {
				t.Errorf("End time should be at 00:00:00 of next day, got %02d:%02d:%02d",
					end.Hour(), end.Minute(), end.Second())
			}

			// Verify timezone
			if start.Location().String() != location.String() {
				t.Errorf("Start time location mismatch: got %s, want %s",
					start.Location(), location)
			}
		})
	}
}

func TestTimezoneHelper_DateRange(t *testing.T) {
	location, _ := time.LoadLocation("Asia/Shanghai")
	helper := NewTimezoneHelper(location)

	// Test date range: 2024-01-01 to 2024-01-31
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, location)
	endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, location)

	start, end := helper.DateRange(startDate, endDate)

	// Verify start is at beginning of start date
	if start.Day() != 1 || start.Hour() != 0 || start.Minute() != 0 {
		t.Errorf("Start should be at 2024-01-01 00:00:00, got %s", start)
	}

	// Verify end is at end of end date
	if end.Day() != 31 || end.Hour() != 23 || end.Minute() != 59 {
		t.Errorf("End should be at 2024-01-31 23:59:59, got %s", end)
	}

	// Verify timezone
	if start.Location().String() != location.String() {
		t.Errorf("Start time location mismatch")
	}
}

func TestTimezoneHelper_IsToday(t *testing.T) {
	location, _ := time.LoadLocation("UTC")
	helper := NewTimezoneHelper(location)

	now := time.Now().In(location)
	today := helper.TodayInLocation()

	// Test current time is today
	if !helper.IsToday(now) {
		t.Error("Current time should be today")
	}

	// Test today's start is today
	if !helper.IsToday(today) {
		t.Error("Today's start should be today")
	}

	// Test yesterday is not today
	yesterday := today.Add(-24 * time.Hour)
	if helper.IsToday(yesterday) {
		t.Error("Yesterday should not be today")
	}

	// Test tomorrow is not today
	tomorrow := today.Add(24 * time.Hour)
	if helper.IsToday(tomorrow) {
		t.Error("Tomorrow should not be today")
	}

	// Test zero time is not today
	var zeroTime time.Time
	if helper.IsToday(zeroTime) {
		t.Error("Zero time should not be today")
	}
}

func TestTimezoneHelper_StartOfDay(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
	}{
		{"UTC", "UTC"},
		{"Asia/Shanghai", "Asia/Shanghai"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, _ := time.LoadLocation(tt.timezone)
			helper := NewTimezoneHelper(location)

			// Test with a time in the middle of the day
			testTime := time.Date(2024, 1, 15, 14, 30, 45, 123456789, location)
			startOfDay := helper.StartOfDay(testTime)

			// Verify it's at the beginning of the day
			if startOfDay.Hour() != 0 || startOfDay.Minute() != 0 || startOfDay.Second() != 0 {
				t.Errorf("Start of day should be at 00:00:00, got %02d:%02d:%02d",
					startOfDay.Hour(), startOfDay.Minute(), startOfDay.Second())
			}

			// Verify it's the same day
			if startOfDay.Day() != testTime.Day() {
				t.Errorf("Start of day should be same day as input")
			}

			// Verify timezone
			if startOfDay.Location().String() != location.String() {
				t.Errorf("Timezone mismatch")
			}
		})
	}
}

func TestTimezoneHelper_EndOfDay(t *testing.T) {
	location, _ := time.LoadLocation("UTC")
	helper := NewTimezoneHelper(location)

	// Test with a time in the middle of the day
	testTime := time.Date(2024, 1, 15, 14, 30, 45, 0, location)
	endOfDay := helper.EndOfDay(testTime)

	// Verify it's at the end of the day
	if endOfDay.Hour() != 23 || endOfDay.Minute() != 59 || endOfDay.Second() != 59 {
		t.Errorf("End of day should be at 23:59:59, got %02d:%02d:%02d",
			endOfDay.Hour(), endOfDay.Minute(), endOfDay.Second())
	}

	// Verify it's the same day
	if endOfDay.Day() != testTime.Day() {
		t.Errorf("End of day should be same day as input")
	}
}

func TestTimezoneHelper_BoundaryTests(t *testing.T) {
	location, _ := time.LoadLocation("Asia/Shanghai")
	helper := NewTimezoneHelper(location)

	// Test midnight boundary
	midnight := time.Date(2024, 1, 15, 0, 0, 0, 0, location)
	startOfDay := helper.StartOfDay(midnight)

	if !startOfDay.Equal(midnight) {
		t.Errorf("Midnight should equal start of day")
	}

	// Test end of day boundary
	endOfDay := helper.EndOfDay(midnight)
	nextMidnight := time.Date(2024, 1, 15, 23, 59, 59, 999999999, location)

	if endOfDay.Before(nextMidnight) {
		t.Errorf("End of day should be at or very close to 23:59:59.999999999")
	}
}

func TestTimezoneHelper_ParseInLocation(t *testing.T) {
	location, _ := time.LoadLocation("America/New_York")
	helper := NewTimezoneHelper(location)

	// Test parsing
	parsed, err := helper.ParseInLocation("2006-01-02", "2024-01-15")
	if err != nil {
		t.Fatalf("Failed to parse date: %v", err)
	}

	// Verify timezone
	if parsed.Location().String() != location.String() {
		t.Errorf("Parsed time should be in configured timezone")
	}

	// Verify values
	if parsed.Year() != 2024 || parsed.Month() != 1 || parsed.Day() != 15 {
		t.Errorf("Parsed date values incorrect: got %s", parsed)
	}
}

func TestTimezoneHelper_FormatInLocation(t *testing.T) {
	location, _ := time.LoadLocation("Europe/Paris")
	helper := NewTimezoneHelper(location)

	testTime := time.Date(2024, 1, 15, 14, 30, 0, 0, location)
	formatted := helper.FormatInLocation(testTime, "2006-01-02 15:04:05")

	expected := "2024-01-15 14:30:00"
	if formatted != expected {
		t.Errorf("Formatted time incorrect: got %s, want %s", formatted, expected)
	}
}

// Benchmark tests
func BenchmarkTimezoneHelper_TodayRange(b *testing.B) {
	location, _ := time.LoadLocation("UTC")
	helper := NewTimezoneHelper(location)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = helper.TodayRange()
	}
}

func BenchmarkTimezoneHelper_IsToday(b *testing.B) {
	location, _ := time.LoadLocation("UTC")
	helper := NewTimezoneHelper(location)
	now := time.Now().In(location)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = helper.IsToday(now)
	}
}
