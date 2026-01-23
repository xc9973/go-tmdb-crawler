package correction

import (
	"time"
)

// StaleShowInfo represents information about a detected stale show
type StaleShowInfo struct {
	ShowID            uint      `json:"show_id"`
	TmdbID            int       `json:"tmdb_id"`
	ShowName          string    `json:"show_name"`
	NormalInterval    int       `json:"normal_interval"`    // Expected update interval in days
	DaysOverdue       int       `json:"days_overdue"`       // How many days past threshold
	LatestEpisodeDate time.Time `json:"latest_episode_date"` // Latest episode air date
	Priority          int       `json:"priority"`           // Higher for more overdue shows
}

// Detector analyzes shows for staleness
type Detector struct {
	// Could add configuration here later
}

// NewDetector creates a new detector instance
func NewDetector() *Detector {
	return &Detector{}
}

// DetectStale analyzes a single show to determine if it's stale
// Returns nil if show is not stale
func (d *Detector) DetectStale(
	showID uint,
	tmdbID int,
	showName string,
	episodeDates []time.Time,
	customThreshold *int,
) *StaleShowInfo {
	// Need at least 3 episodes to analyze
	if len(episodeDates) < 3 {
		return nil
	}

	// Get last 10 episodes for pattern analysis
	n := 10
	if len(episodeDates) < n {
		n = len(episodeDates)
	}
	lastN := episodeDates[len(episodeDates)-n:]

	// Calculate intervals
	intervals := d.calculateIntervals(lastN)
	if len(intervals) == 0 {
		return nil
	}

	// Analyze pattern
	pattern := CalculateUpdatePattern(intervals)

	// Use custom threshold if set, otherwise use calculated
	threshold := pattern.Threshold
	if customThreshold != nil {
		threshold = *customThreshold
	}

	// Get latest episode date
	latestDate := lastN[len(lastN)-1]
	daysSinceLatest := int(time.Since(latestDate).Hours() / 24)

	// Check if stale
	if daysSinceLatest <= threshold {
		return nil // Not stale
	}

	// Calculate priority based on how overdue
	daysOverdue := daysSinceLatest - threshold
	priority := daysOverdue
	if priority > 100 {
		priority = 100 // Cap at 100
	}

	return &StaleShowInfo{
		ShowID:            showID,
		TmdbID:            tmdbID,
		ShowName:          showName,
		NormalInterval:    pattern.Mode,
		DaysOverdue:       daysOverdue,
		LatestEpisodeDate: latestDate,
		Priority:          priority,
	}
}

// calculateIntervals converts sorted dates to day intervals
func (d *Detector) calculateIntervals(dates []time.Time) []int {
	if len(dates) < 2 {
		return nil
	}

	intervals := make([]int, 0, len(dates)-1)
	for i := 1; i < len(dates); i++ {
		days := int(dates[i].Sub(dates[i-1]).Hours() / 24)
		if days > 0 { // Only positive intervals
			intervals = append(intervals, days)
		}
	}

	return intervals
}
