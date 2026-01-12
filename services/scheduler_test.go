package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/xc9973/go-tmdb-crawler/utils"
)

func TestValidateCronSpec(t *testing.T) {
	tests := []struct {
		name      string
		spec      string
		wantValid bool
	}{
		// Valid 6-field cron expressions (with seconds)
		{
			name:      "Valid - Every second",
			spec:      "* * * * * *",
			wantValid: true,
		},
		{
			name:      "Valid - Daily at 8:00:00 AM",
			spec:      "0 0 8 * * *",
			wantValid: true,
		},
		{
			name:      "Valid - Daily at 8:00:00, 12:00:00, 20:00:00",
			spec:      "0 0 8,12,20 * * *",
			wantValid: true,
		},
		{
			name:      "Valid - Daily at 8:30:00 PM",
			spec:      "0 30 20 * * *",
			wantValid: true,
		},
		{
			name:      "Valid - Weekly Monday 6:00:00 AM",
			spec:      "0 0 6 * * 1",
			wantValid: true,
		},
		{
			name:      "Valid - Weekly Monday 7:00:00 AM",
			spec:      "0 0 7 * * 1",
			wantValid: true,
		},
		{
			name:      "Valid - Every 5 seconds",
			spec:      "*/5 * * * * *",
			wantValid: true,
		},
		{
			name:      "Valid - Every minute at 0 seconds",
			spec:      "0 * * * * *",
			wantValid: true,
		},
		{
			name:      "Valid - Hourly at 0:00",
			spec:      "0 0 * * * *",
			wantValid: true,
		},
		{
			name:      "Valid - Monthly at 00:00:00 on 1st",
			spec:      "0 0 0 1 * *",
			wantValid: true,
		},
		{
			name:      "Valid - Using @yearly descriptor",
			spec:      "@yearly",
			wantValid: true,
		},
		{
			name:      "Valid - Using @monthly descriptor",
			spec:      "@monthly",
			wantValid: true,
		},
		{
			name:      "Valid - Using @weekly descriptor",
			spec:      "@weekly",
			wantValid: true,
		},
		{
			name:      "Valid - Using @daily descriptor",
			spec:      "@daily",
			wantValid: true,
		},
		{
			name:      "Valid - Using @hourly descriptor",
			spec:      "@hourly",
			wantValid: true,
		},

		// Invalid expressions
		{
			name:      "Invalid - 5-field cron (no seconds)",
			spec:      "0 8 * * *",
			wantValid: false,
		},
		{
			name:      "Invalid - Empty string",
			spec:      "",
			wantValid: false,
		},
		{
			name:      "Invalid - Too many fields",
			spec:      "* * * * * * *",
			wantValid: false,
		},
		{
			name:      "Invalid - Out of range seconds",
			spec:      "60 * * * * *",
			wantValid: false,
		},
		{
			name:      "Invalid - Out of range minutes",
			spec:      "0 60 * * * *",
			wantValid: false,
		},
		{
			name:      "Invalid - Out of range hours",
			spec:      "0 0 24 * * *",
			wantValid: false,
		},
		{
			name:      "Invalid - Out of range day",
			spec:      "0 0 0 32 * *",
			wantValid: false,
		},
		{
			name:      "Invalid - Out of range month",
			spec:      "0 0 0 1 13 *",
			wantValid: false,
		},
		{
			name:      "Invalid - Out of range weekday",
			spec:      "0 0 0 * * 8",
			wantValid: false,
		},
		{
			name:      "Invalid - Negative number",
			spec:      "-1 * * * * *",
			wantValid: false,
		},
		{
			name:      "Invalid - Invalid characters",
			spec:      "a b c d e f",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCronSpec(tt.spec)
			if (err == nil) != tt.wantValid {
				t.Errorf("ValidateCronSpec(%q) error = %v, wantValid %v", tt.spec, err, tt.wantValid)
			}
		})
	}
}

func TestValidateCronSpec_WithSeconds(t *testing.T) {
	// Test that 6-field expressions with seconds are accepted
	// This matches the cron.WithSeconds() option used in NewScheduler
	spec := "0 0 8 * * *" // Daily at 8:00:00 AM
	err := ValidateCronSpec(spec)
	if err != nil {
		t.Errorf("6-field cron expression should be valid, got error: %v", err)
	}
}

func TestValidateCronSpec_WithoutSeconds(t *testing.T) {
	// Test that 5-field expressions (without seconds) are rejected
	// because we use cron.WithSeconds() option
	spec := "0 8 * * *" // 5-field expression (no seconds)
	err := ValidateCronSpec(spec)
	if err == nil {
		t.Error("5-field cron expression should be invalid when using cron.WithSeconds()")
	}
}

func TestGetDefaultCronSpecs(t *testing.T) {
	specs := GetDefaultCronSpecs()

	expectedSpecs := map[string]string{
		"daily_crawl":    "0 0 8,12,20 * * *",
		"daily_publish":  "0 30 20 * * *",
		"weekly_crawl":   "0 0 6 * * 1",
		"weekly_publish": "0 0 7 * * 1",
	}

	for jobType, expectedSpec := range expectedSpecs {
		if specs[jobType] != expectedSpec {
			t.Errorf("GetDefaultCronSpecs()[%q] = %q, want %q", jobType, specs[jobType], expectedSpec)
		}
	}

	// Verify all default specs are valid
	for jobType, spec := range specs {
		if err := ValidateCronSpec(spec); err != nil {
			t.Errorf("Default cron spec for %q is invalid: %q, error: %v", jobType, spec, err)
		}
	}
}

func TestValidateCronSpec_CronLibraryConsistency(t *testing.T) {
	// Test that ValidateCronSpec uses the same parser as cron.WithSeconds()
	specs := []string{
		"0 0 8 * * *",   // Daily at 8:00:00 AM
		"0 30 20 * * *", // Daily at 8:30:00 PM
		"0 0 6 * * 1",   // Monday 6:00:00 AM
		"*/5 * * * * *", // Every 5 seconds
		"@hourly",       // Every hour
	}

	// Create a cron with seconds (same as NewScheduler)
	c := cron.New(cron.WithSeconds())

	for _, spec := range specs {
		t.Run(spec, func(t *testing.T) {
			// Validate using our function
			if err := ValidateCronSpec(spec); err != nil {
				t.Errorf("ValidateCronSpec(%q) returned error: %v", spec, err)
			}

			// Try to add the spec to the cron (this should not fail)
			id, err := c.AddFunc(spec, func() {})
			if err != nil {
				t.Errorf("cron.AddFunc(%q) failed: %v (should match ValidateCronSpec)", spec, err)
			} else {
				// Clean up
				c.Remove(id)
			}
		})
	}
}

func TestValidateCronSpec_Descriptors(t *testing.T) {
	// Test predefined descriptors
	descriptors := []string{
		"@yearly",
		"@annually",
		"@monthly",
		"@weekly",
		"@daily",
		"@midnight",
		"@hourly",
	}

	for _, descriptor := range descriptors {
		t.Run(descriptor, func(t *testing.T) {
			if err := ValidateCronSpec(descriptor); err != nil {
				t.Errorf("Descriptor %q should be valid, got error: %v", descriptor, err)
			}
		})
	}
}

func TestValidateCronSpec_ComplexExpressions(t *testing.T) {
	complexSpecs := []struct {
		name string
		spec string
	}{
		{
			name: "Multiple values in each field",
			spec: "0 15,30,45 8,12,20 1,15 * 1,2,3,4,5",
		},
		{
			name: "Ranges",
			spec: "0 0 8-18 * * 1-5",
		},
		{
			name: "Steps with ranges",
			spec: "0 */10 8-18 * * 1-5",
		},
		{
			name: "Last day of month (using 31)",
			spec: "0 0 0 31 * *",
		},
		{
			name: "Specific weekday",
			spec: "0 0 8 * * MON,WED,FRI",
		},
	}

	for _, tt := range complexSpecs {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateCronSpec(tt.spec); err != nil {
				t.Errorf("Complex spec %q should be valid, got error: %v", tt.spec, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidateCronSpec(b *testing.B) {
	spec := "0 0 8,12,20 * * *"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateCronSpec(spec)
	}
}

// Concurrency control tests
func TestScheduler_ConcurrencyControl(t *testing.T) {
	// Create mock services
	logger := utils.NewLogger("info", "")
	crawler := &CrawlerService{}
	publisher := &PublisherService{}

	scheduler := NewScheduler(crawler, publisher, logger)

	// Test that mutex prevents concurrent execution
	t.Run("CrawlJobMutex", func(t *testing.T) {
		// Try to acquire the lock
		if !scheduler.crawlJobMutex.TryLock() {
			// Should be able to acquire
			scheduler.crawlJobMutex.Unlock()
		} else {
			scheduler.crawlJobMutex.Unlock()
		}

		// Acquire lock
		scheduler.crawlJobMutex.Lock()

		// Try to acquire again - should fail
		if scheduler.crawlJobMutex.TryLock() {
			t.Error("Should not be able to acquire lock twice")
			scheduler.crawlJobMutex.Unlock()
		}

		scheduler.crawlJobMutex.Unlock()
	})

	t.Run("PublishJobMutex", func(t *testing.T) {
		// Acquire lock
		scheduler.publishJobMutex.Lock()

		// Try to acquire again - should fail
		if scheduler.publishJobMutex.TryLock() {
			t.Error("Should not be able to acquire lock twice")
			scheduler.publishJobMutex.Unlock()
		}

		scheduler.publishJobMutex.Unlock()
	})
}

func TestScheduler_TimeoutSettings(t *testing.T) {
	logger := utils.NewLogger("info", "")
	crawler := &CrawlerService{}
	publisher := &PublisherService{}

	scheduler := NewScheduler(crawler, publisher, logger)

	// Test default timeouts
	timeouts := scheduler.GetTimeouts()
	if timeouts["crawl_timeout"] != "30m0s" {
		t.Errorf("Default crawl timeout should be 30m, got %s", timeouts["crawl_timeout"])
	}
	if timeouts["publish_timeout"] != "10m0s" {
		t.Errorf("Default publish timeout should be 10m, got %s", timeouts["publish_timeout"])
	}

	// Test setting custom timeouts
	scheduler.SetTimeouts(15*time.Minute, 5*time.Minute)
	timeouts = scheduler.GetTimeouts()
	if timeouts["crawl_timeout"] != "15m0s" {
		t.Errorf("Crawl timeout should be 15m, got %s", timeouts["crawl_timeout"])
	}
	if timeouts["publish_timeout"] != "5m0s" {
		t.Errorf("Publish timeout should be 5m, got %s", timeouts["publish_timeout"])
	}
}

func TestScheduler_RunJobWithTimeout(t *testing.T) {
	logger := utils.NewLogger("info", "")
	crawler := &CrawlerService{}
	publisher := &PublisherService{}

	scheduler := NewScheduler(crawler, publisher, logger)

	t.Run("JobCompletesWithinTimeout", func(t *testing.T) {
		job := func() error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}

		err := scheduler.runJobWithTimeout("test_job", 1*time.Second, job)
		if err != nil {
			t.Errorf("Job should complete within timeout, got error: %v", err)
		}
	})

	t.Run("JobTimesOut", func(t *testing.T) {
		job := func() error {
			time.Sleep(2 * time.Second)
			return nil
		}

		err := scheduler.runJobWithTimeout("test_job", 100*time.Millisecond, job)
		if err == nil {
			t.Error("Job should timeout")
		}
	})

	t.Run("JobReturnsError", func(t *testing.T) {
		expectedErr := fmt.Errorf("job failed")
		job := func() error {
			return expectedErr
		}

		err := scheduler.runJobWithTimeout("test_job", 1*time.Second, job)
		if err != expectedErr {
			t.Errorf("Job should return error, got: %v", err)
		}
	})
}

func TestScheduler_StatusIncludesConcurrencyInfo(t *testing.T) {
	logger := utils.NewLogger("info", "")
	crawler := &CrawlerService{}
	publisher := &PublisherService{}

	scheduler := NewScheduler(crawler, publisher, logger)

	status := scheduler.GetStatus()

	// Check that status includes concurrency info
	if _, ok := status["crawl_job_running"]; !ok {
		t.Error("Status should include crawl_job_running")
	}
	if _, ok := status["publish_job_running"]; !ok {
		t.Error("Status should include publish_job_running")
	}
}
