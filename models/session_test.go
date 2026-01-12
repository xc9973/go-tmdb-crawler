package models

import (
	"testing"
	"time"
)

func TestSession_Validate(t *testing.T) {
	tests := []struct {
		name    string
		session *Session
		wantErr bool
	}{
		{
			name: "Valid session",
			session: &Session{
				SessionID: "test-session-id",
				Token:     "test-token",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "Empty session ID",
			session: &Session{
				SessionID: "",
				Token:     "test-token",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "Empty token",
			session: &Session{
				SessionID: "test-session-id",
				Token:     "",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "Zero expiration time",
			session: &Session{
				SessionID: "test-session-id",
				Token:     "test-token",
				ExpiresAt: time.Time{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.session.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Session.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSession_IsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		expiresAt   time.Time
		expected    bool
		description string
	}{
		{
			name:        "Expired session",
			expiresAt:   now.Add(-1 * time.Hour),
			expected:    true,
			description: "Session expired 1 hour ago",
		},
		{
			name:        "Valid session",
			expiresAt:   now.Add(1 * time.Hour),
			expected:    false,
			description: "Session expires in 1 hour",
		},
		{
			name:        "Just expired",
			expiresAt:   now.Add(-1 * time.Second),
			expected:    true,
			description: "Session just expired",
		},
		{
			name:        "Expires now",
			expiresAt:   now,
			expected:    true,
			description: "Session expires exactly now",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{ExpiresAt: tt.expiresAt}
			if got := session.IsExpired(); got != tt.expected {
				t.Errorf("Session.IsExpired() = %v, want %v (%s)", got, tt.expected, tt.description)
			}
		})
	}
}

func TestSession_IsValid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		expiresAt   time.Time
		expected    bool
		description string
	}{
		{
			name:        "Valid session",
			expiresAt:   now.Add(1 * time.Hour),
			expected:    true,
			description: "Session is still valid",
		},
		{
			name:        "Expired session",
			expiresAt:   now.Add(-1 * time.Hour),
			expected:    false,
			description: "Session has expired",
		},
		{
			name:        "Expires in future",
			expiresAt:   now.Add(24 * time.Hour),
			expected:    true,
			description: "Session expires in 24 hours",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{ExpiresAt: tt.expiresAt}
			if got := session.IsValid(); got != tt.expected {
				t.Errorf("Session.IsValid() = %v, want %v (%s)", got, tt.expected, tt.description)
			}
		})
	}
}

func TestSession_Refresh(t *testing.T) {
	now := time.Now()
	duration := 24 * time.Hour

	session := &Session{
		SessionID:  "test-session",
		Token:      "test-token",
		ExpiresAt:  now.Add(1 * time.Hour),
		LastActive: now.Add(-1 * time.Hour),
	}

	// Store original values
	originalExpiresAt := session.ExpiresAt
	originalLastActive := session.LastActive

	// Refresh session
	session.Refresh(duration)

	// Check that ExpiresAt was extended
	expectedExpiresAt := now.Add(duration)
	if session.ExpiresAt.Before(expectedExpiresAt.Add(-1*time.Minute)) ||
		session.ExpiresAt.After(expectedExpiresAt.Add(1*time.Minute)) {
		t.Errorf("Session.Refresh() ExpiresAt = %v, want approximately %v", session.ExpiresAt, expectedExpiresAt)
	}

	// Check that ExpiresAt changed
	if session.ExpiresAt.Equal(originalExpiresAt) || session.ExpiresAt.Before(originalExpiresAt) {
		t.Errorf("Session.Refresh() should extend ExpiresAt, was %v, now %v", originalExpiresAt, session.ExpiresAt)
	}

	// Check that LastActive was updated
	if session.LastActive.Before(originalLastActive) || session.LastActive.Equal(originalLastActive) {
		t.Errorf("Session.Refresh() should update LastActive, was %v, now %v", originalLastActive, session.LastActive)
	}
}

func TestSession_TableName(t *testing.T) {
	session := Session{}
	if got := session.TableName(); got != "sessions" {
		t.Errorf("Session.TableName() = %v, want %v", got, "sessions")
	}
}

func TestSession_Refresh_DifferentDurations(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
	}{
		{
			name:     "1 hour",
			duration: 1 * time.Hour,
		},
		{
			name:     "24 hours",
			duration: 24 * time.Hour,
		},
		{
			name:     "7 days",
			duration: 7 * 24 * time.Hour,
		},
		{
			name:     "30 days",
			duration: 30 * 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{
				SessionID: "test-session",
				Token:     "test-token",
				ExpiresAt: time.Now(),
			}

			session.Refresh(tt.duration)

			// Check that expiration is approximately correct
			expected := time.Now().Add(tt.duration)
			if session.ExpiresAt.Before(expected.Add(-1*time.Minute)) ||
				session.ExpiresAt.After(expected.Add(1*time.Minute)) {
				t.Errorf("Session.Refresh() with duration %v resulted in ExpiresAt = %v, want approximately %v",
					tt.duration, session.ExpiresAt, expected)
			}
		})
	}
}
