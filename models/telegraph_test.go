package models

import (
	"testing"
	"time"
)

func TestTelegraphPost_GetFullURL(t *testing.T) {
	tests := []struct {
		name          string
		telegraphURL  string
		telegraphPath string
		expected      string
	}{
		{
			name:          "Full URL exists",
			telegraphURL:  "https://telegra.ph/test-article-12-34",
			telegraphPath: "test-article-12-34",
			expected:      "https://telegra.ph/test-article-12-34",
		},
		{
			name:          "No full URL, construct from path",
			telegraphURL:  "",
			telegraphPath: "test-article-12-34",
			expected:      "https://telegra.ph/test-article-12-34",
		},
		{
			name:          "Empty path",
			telegraphURL:  "",
			telegraphPath: "",
			expected:      "https://telegra.ph/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post := &TelegraphPost{
				TelegraphURL:  tt.telegraphURL,
				TelegraphPath: tt.telegraphPath,
			}
			if got := post.GetFullURL(); got != tt.expected {
				t.Errorf("TelegraphPost.GetFullURL() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGenerateContentHash(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantLen int
	}{
		{
			name:    "Simple content",
			content: "test content",
			wantLen: 32, // MD5 hash is always 32 characters
		},
		{
			name:    "Empty content",
			content: "",
			wantLen: 32,
		},
		{
			name:    "Long content",
			content: "a" + "b" + "c" + "d" + "e" + "f" + "g" + "h" + "i" + "j",
			wantLen: 32,
		},
		{
			name:    "Special characters",
			content: "!@#$%^&*()_+-=[]{}|;':\",./<>?",
			wantLen: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateContentHash(tt.content)
			if len(got) != tt.wantLen {
				t.Errorf("GenerateContentHash() length = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestGenerateContentHash_Consistency(t *testing.T) {
	content := "test content"

	// Generate hash multiple times
	hash1 := GenerateContentHash(content)
	hash2 := GenerateContentHash(content)
	hash3 := GenerateContentHash(content)

	// All hashes should be identical
	if hash1 != hash2 || hash2 != hash3 {
		t.Errorf("GenerateContentHash() not consistent: %v, %v, %v", hash1, hash2, hash3)
	}
}

func TestGenerateContentHash_Uniqueness(t *testing.T) {
	content1 := "content 1"
	content2 := "content 2"

	hash1 := GenerateContentHash(content1)
	hash2 := GenerateContentHash(content2)

	// Different content should produce different hashes
	if hash1 == hash2 {
		t.Errorf("GenerateContentHash() produced same hash for different content")
	}
}

func TestTelegraphPost_IsRecent(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		createdAt   time.Time
		expected    bool
		description string
	}{
		{
			name:        "Created 1 hour ago",
			createdAt:   now.Add(-1 * time.Hour),
			expected:    true,
			description: "Post created within 24 hours",
		},
		{
			name:        "Created 23 hours ago",
			createdAt:   now.Add(-23 * time.Hour),
			expected:    true,
			description: "Post created within 24 hours",
		},
		{
			name:        "Created 25 hours ago",
			createdAt:   now.Add(-25 * time.Hour),
			expected:    false,
			description: "Post created more than 24 hours ago",
		},
		{
			name:        "Created just now",
			createdAt:   now,
			expected:    true,
			description: "Post just created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post := &TelegraphPost{CreatedAt: tt.createdAt}
			if got := post.IsRecent(); got != tt.expected {
				t.Errorf("TelegraphPost.IsRecent() = %v, want %v (%s)", got, tt.expected, tt.description)
			}
		})
	}
}

func TestTelegraphPost_Validate(t *testing.T) {
	tests := []struct {
		name    string
		post    *TelegraphPost
		wantErr bool
	}{
		{
			name: "Valid post",
			post: &TelegraphPost{
				Title:         "Test Title",
				TelegraphPath: "test-path",
				ContentHash:   "abc123",
			},
			wantErr: false,
		},
		{
			name: "Empty title",
			post: &TelegraphPost{
				Title:         "",
				TelegraphPath: "test-path",
				ContentHash:   "abc123",
			},
			wantErr: true,
		},
		{
			name: "Empty telegraph path",
			post: &TelegraphPost{
				Title:         "Test Title",
				TelegraphPath: "",
				ContentHash:   "abc123",
			},
			wantErr: true,
		},
		{
			name: "Empty content hash",
			post: &TelegraphPost{
				Title:         "Test Title",
				TelegraphPath: "test-path",
				ContentHash:   "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.post.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("TelegraphPost.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTelegraphPost_GetShortURL(t *testing.T) {
	tests := []struct {
		name          string
		telegraphPath string
		expected      string
	}{
		{
			name:          "Valid path",
			telegraphPath: "test-article-12-34",
			expected:      "https://telegra.ph/test-article-12-34",
		},
		{
			name:          "Empty path",
			telegraphPath: "",
			expected:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post := &TelegraphPost{TelegraphPath: tt.telegraphPath}
			if got := post.GetShortURL(); got != tt.expected {
				t.Errorf("TelegraphPost.GetShortURL() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTelegraphPost_WasCreatedToday(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		createdAt   time.Time
		expected    bool
		description string
	}{
		{
			name:        "Created today",
			createdAt:   now,
			expected:    true,
			description: "Post created today",
		},
		{
			name:        "Created earlier today",
			createdAt:   now.Add(-5 * time.Hour),
			expected:    true,
			description: "Post created earlier today",
		},
		{
			name:        "Created yesterday",
			createdAt:   now.Add(-24 * time.Hour),
			expected:    false,
			description: "Post created yesterday",
		},
		{
			name:        "Created tomorrow",
			createdAt:   now.Add(24 * time.Hour),
			expected:    false,
			description: "Post created tomorrow (should not happen in practice)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post := &TelegraphPost{CreatedAt: tt.createdAt}
			if got := post.WasCreatedToday(); got != tt.expected {
				t.Errorf("TelegraphPost.WasCreatedToday() = %v, want %v (%s)", got, tt.expected, tt.description)
			}
		})
	}
}

func TestTelegraphPost_TableName(t *testing.T) {
	post := TelegraphPost{}
	if got := post.TableName(); got != "telegraph_posts" {
		t.Errorf("TelegraphPost.TableName() = %v, want %v", got, "telegraph_posts")
	}
}
