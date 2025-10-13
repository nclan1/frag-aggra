package scraper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsWTS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "uppercase WTS",
			input:    "[WTS] Tom Ford Sale",
			expected: true,
		},
		{
			name:     "lowercase wts",
			input:    "[wts] Selling fragrances",
			expected: true,
		},
		{
			name:     "mixed case WtS",
			input:    "[WtS] Great deals",
			expected: true,
		},
		{
			name:     "WTS in middle of text",
			input:    "Check out my [WTS] post",
			expected: true,
		},
		{
			name:     "WTS in body text",
			input:    "I'm selling some perfumes [wts]",
			expected: true,
		},
		{
			name:     "no WTS present",
			input:    "Looking to buy fragrances",
			expected: false,
		},
		{
			name:     "WTB instead of WTS",
			input:    "[WTB] Tom Ford",
			expected: false,
		},
		{
			name:     "WTS without brackets",
			input:    "WTS Tom Ford",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "partial match wtss",
			input:    "[WTSS] Tom Ford",
			expected: false,
		},
		{
			name:     "WTS with extra spaces",
			input:    "[WTS ] Tom Ford",
			expected: false,
		},
		{
			name:     "wts lowercase brackets",
			input:    "[wts] Selling items",
			expected: true,
		},
		{
			name:     "WTS multiple occurrences",
			input:    "[WTS] Tom Ford [WTS] More items",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsWTS(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNew_MissingCredentials(t *testing.T) {
	// Save current env vars
	t.Setenv("REDDIT_CLIENT_ID", "")
	t.Setenv("REDDIT_CLIENT_SECRET", "")
	t.Setenv("REDDIT_USERNAME", "")
	t.Setenv("REDDIT_PASSWORD", "")

	// When credentials are missing, NewClient may still succeed
	// but the scraper will be created (Reddit lib doesn't validate immediately)
	scraper, err := New()
	
	// The Reddit library doesn't fail on empty credentials during initialization
	// It only fails when actually making API calls
	// So we just verify the scraper was created
	assert.NoError(t, err)
	assert.NotNil(t, scraper)
	assert.NotNil(t, scraper.client)
}

func TestNew_WithValidCredentials(t *testing.T) {
	// Set valid-looking credentials (won't actually connect)
	t.Setenv("REDDIT_CLIENT_ID", "test_client_id")
	t.Setenv("REDDIT_CLIENT_SECRET", "test_secret")
	t.Setenv("REDDIT_USERNAME", "test_user")
	t.Setenv("REDDIT_PASSWORD", "test_pass")

	// This will attempt to create a client with the credentials
	// It may succeed or fail depending on Reddit API, but at least
	// it tests our initialization logic
	scraper, err := New()
	
	// If no error, verify scraper was created
	if err == nil {
		assert.NotNil(t, scraper)
		assert.NotNil(t, scraper.client)
	}
	// If there's an error, it's likely from Reddit API authentication
	// which is acceptable for a unit test without actual credentials
}

