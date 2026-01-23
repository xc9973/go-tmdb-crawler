package correction

// UpdateInterval represents the calculated update pattern
type UpdateInterval struct {
	Mode         int  // Most common interval (in days)
	Threshold    int  // 1.5x mode (trigger threshold)
	SampleSize   int  // Number of intervals analyzed
	HasGapSeason bool // Whether gap seasons (>60 days) were filtered
}

// CalculateUpdatePattern analyzes episode air date intervals
// to determine the normal update frequency.
func CalculateUpdatePattern(intervals []int) *UpdateInterval {
	if len(intervals) == 0 {
		return &UpdateInterval{Mode: 7, Threshold: 10, SampleSize: 0} // Default: weekly
	}

	// Filter out gap seasons (>60 days indicates season break)
	filtered := make([]int, 0, len(intervals))
	for _, interval := range intervals {
		if interval <= 60 {
			filtered = append(filtered, interval)
		}
	}

	// If no valid intervals, use default
	if len(filtered) == 0 {
		return &UpdateInterval{Mode: 7, Threshold: 10, SampleSize: len(intervals), HasGapSeason: true}
	}

	// Calculate mode (most common value)
	mode := calculateMode(filtered)
	threshold := int(float64(mode) * 1.5)

	return &UpdateInterval{
		Mode:         mode,
		Threshold:    threshold,
		SampleSize:   len(filtered),
		HasGapSeason: len(filtered) < len(intervals),
	}
}

// calculateMode finds the most common value in a slice
func calculateMode(values []int) int {
	if len(values) == 0 {
		return 7 // Default to weekly
	}

	// Count frequency
	freq := make(map[int]int)
	for _, v := range values {
		freq[v]++
	}

	// Find most frequent
	mode := values[0]
	maxCount := freq[mode]
	for v, count := range freq {
		if count > maxCount {
			mode = v
			maxCount = count
		}
	}

	return mode
}

// GetLastNEpisodesIntervals gets intervals from the last N episodes
// Returns slice of day gaps between consecutive episodes
func GetLastNEpisodesIntervals(intervals []int, n int) []int {
	if len(intervals) < n {
		n = len(intervals)
	}
	if n < 1 {
		return nil
	}

	// Take last N intervals
	start := len(intervals) - n
	if start < 0 {
		start = 0
	}

	result := make([]int, n)
	copy(result, intervals[start:])
	return result
}
