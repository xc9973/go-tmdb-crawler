package utils

import (
	"time"
)

// TimezoneHelper provides timezone-aware time operations
type TimezoneHelper struct {
	location *time.Location
}

// NewTimezoneHelper creates a new timezone helper with the specified location
func NewTimezoneHelper(location *time.Location) *TimezoneHelper {
	return &TimezoneHelper{
		location: location,
	}
}

// TodayInLocation returns today's date in the configured timezone
// Returns the start of the day (00:00:00) in the configured timezone
func (h *TimezoneHelper) TodayInLocation() time.Time {
	now := time.Now().In(h.location)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, h.location)
}

// StartOfDay returns the start of the given day (00:00:00) in the configured timezone
func (h *TimezoneHelper) StartOfDay(t time.Time) time.Time {
	localTime := t.In(h.location)
	return time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, h.location)
}

// EndOfDay returns the end of the given day (23:59:59.999999999) in the configured timezone
func (h *TimezoneHelper) EndOfDay(t time.Time) time.Time {
	localTime := t.In(h.location)
	return time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 23, 59, 59, 999999999, h.location)
}

// TodayRange returns the start and end of today in the configured timezone
// The range is [start, end) - start is inclusive, end is exclusive
func (h *TimezoneHelper) TodayRange() (time.Time, time.Time) {
	start := h.TodayInLocation()
	end := start.Add(24 * time.Hour)
	return start, end
}

// DateRange returns the start and end of a date range in the configured timezone
// The range is [startDate, endDate] - both inclusive
func (h *TimezoneHelper) DateRange(startDate, endDate time.Time) (time.Time, time.Time) {
	start := h.StartOfDay(startDate)
	end := h.EndOfDay(endDate)
	return start, end
}

// IsToday checks if the given time is today in the configured timezone
func (h *TimezoneHelper) IsToday(t time.Time) bool {
	if t.IsZero() {
		return false
	}
	today := h.TodayInLocation()
	localTime := t.In(h.location)
	return localTime.Year() == today.Year() &&
		localTime.Month() == today.Month() &&
		localTime.Day() == today.Day()
}

// InLocation converts a time to the configured timezone
func (h *TimezoneHelper) InLocation(t time.Time) time.Time {
	return t.In(h.location)
}

// NowInLocation returns current time in the configured timezone
func (h *TimezoneHelper) NowInLocation() time.Time {
	return time.Now().In(h.location)
}

// GetLocation returns the configured location
func (h *TimezoneHelper) GetLocation() *time.Location {
	return h.location
}

// ParseInLocation parses a time string in the configured timezone
// The layout should match the format of the time string
func (h *TimezoneHelper) ParseInLocation(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, h.location)
}

// FormatInLocation formats a time in the configured timezone
func (h *TimezoneHelper) FormatInLocation(t time.Time, layout string) string {
	return t.In(h.location).Format(layout)
}
