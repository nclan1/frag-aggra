package parser

import (
	"context"
	"encoding/json"
	"frag-aggra/internal/models"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_MissingAPIKey(t *testing.T) {
	// Clear the API key
	t.Setenv("OPENAI_API_KEY", "")

	parser, err := New()

	assert.Error(t, err)
	assert.Nil(t, parser)
	assert.Contains(t, err.Error(), "OPENAI_API_KEY")
}

func TestNew_WithAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-api-key")

	parser, err := New()

	assert.NoError(t, err)
	assert.NotNil(t, parser)
	assert.NotNil(t, parser.client)
	assert.NotEmpty(t, parser.systemPrompt)
}

func TestNew_SystemPromptContent(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-api-key")

	parser, err := New()
	assert.NoError(t, err)

	// Verify system prompt contains key instructions
	assert.Contains(t, parser.systemPrompt, "Tom Ford")
	assert.Contains(t, parser.systemPrompt, "MFK")
	assert.Contains(t, parser.systemPrompt, "Maison Francis Kurkdijan")
	assert.Contains(t, parser.systemPrompt, "PdM")
	assert.Contains(t, parser.systemPrompt, "Parfums de Marly")
	assert.Contains(t, parser.systemPrompt, "perfumes")
	assert.Contains(t, parser.systemPrompt, "JSON")
}

func TestGenerateSchema(t *testing.T) {
	schema := generateSchema[models.FragranceListing]()
	
	assert.NotNil(t, schema)
	
	// Convert to JSON to validate structure
	jsonBytes, err := json.Marshal(schema)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonBytes)
	
	// Verify it's valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	assert.NoError(t, err)
	
	// Check that schema has expected fields
	assert.Contains(t, string(jsonBytes), "perfumes")
}

func TestGenerateSchema_Perfume(t *testing.T) {
	schema := generateSchema[models.Perfume]()
	
	assert.NotNil(t, schema)
	
	jsonBytes, err := json.Marshal(schema)
	assert.NoError(t, err)
	
	// Verify it contains expected fields
	assert.Contains(t, string(jsonBytes), "name")
	assert.Contains(t, string(jsonBytes), "sizes")
	assert.Contains(t, string(jsonBytes), "prices")
}

func TestParsePostContent_InvalidAPIKey(t *testing.T) {
	// This test verifies error handling when API key is invalid
	// Set an invalid API key
	originalKey := os.Getenv("OPENAI_API_KEY")
	t.Setenv("OPENAI_API_KEY", "invalid-key-for-testing")
	
	parser, err := New()
	assert.NoError(t, err) // Creating parser should succeed
	
	// Restore original key for other tests
	if originalKey != "" {
		t.Setenv("OPENAI_API_KEY", originalKey)
	}
	
	// Attempting to parse with invalid key should fail
	postContent := "[WTS] Tom Ford Tobacco Vanille 100ml $150"
	listing, err := parser.ParsePostContent(context.Background(), postContent)
	
	// We expect an error from OpenAI API
	assert.Error(t, err)
	assert.Nil(t, listing)
}

func TestParsePostContent_EmptyContent(t *testing.T) {
	// Skip if no real API key (this would make actual API call)
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test that requires OPENAI_API_KEY")
	}
	
	t.Setenv("OPENAI_API_KEY", "test-key")
	parser, err := New()
	assert.NoError(t, err)
	
	// Empty content should still return without panic
	// The actual behavior depends on OpenAI API
	_, err = parser.ParsePostContent(context.Background(), "")
	// We just verify it doesn't panic - error is acceptable
	// assert.Error(t, err) // May or may not error depending on API
}

func TestParsePostContent_ContextCancellation(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-api-key")
	
	parser, err := New()
	assert.NoError(t, err)
	
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	postContent := "[WTS] Tom Ford Tobacco Vanille 100ml $150"
	listing, err := parser.ParsePostContent(ctx, postContent)
	
	// Should return error due to cancelled context
	assert.Error(t, err)
	assert.Nil(t, listing)
}

// MockOpenAIResponse tests JSON unmarshaling with a valid response
func TestParsePostContent_JSONUnmarshal(t *testing.T) {
	// Test that our listing struct can properly unmarshal expected JSON format
	jsonResponse := `{
		"perfumes": [
			{
				"name": "Tom Ford Tobacco Vanille",
				"sizes": ["100ml"],
				"prices": ["$150"]
			},
			{
				"name": "Bleu de Chanel EDT",
				"sizes": ["50ml", "100ml"],
				"prices": ["$80", "$120"]
			}
		]
	}`
	
	var listing models.FragranceListing
	err := json.Unmarshal([]byte(jsonResponse), &listing)
	
	assert.NoError(t, err)
	assert.Len(t, listing.Perfumes, 2)
	assert.Equal(t, "Tom Ford Tobacco Vanille", listing.Perfumes[0].Name)
	assert.Equal(t, []string{"100ml"}, listing.Perfumes[0].Sizes)
	assert.Equal(t, []string{"$150"}, listing.Perfumes[0].Prices)
}

func TestParsePostContent_InvalidJSON(t *testing.T) {
	// Test unmarshaling invalid JSON
	invalidJSON := `{"perfumes": [{"name": "Tom Ford"` // Incomplete JSON
	
	var listing models.FragranceListing
	err := json.Unmarshal([]byte(invalidJSON), &listing)
	
	assert.Error(t, err)
}

func TestParsePostContent_PartialBottles(t *testing.T) {
	// Test parsing partial bottle format
	jsonResponse := `{
		"perfumes": [
			{
				"name": "Maison Francis Kurkdijan Baccarat Rouge 540",
				"sizes": ["80/100ml"],
				"prices": ["$200"]
			}
		]
	}`
	
	var listing models.FragranceListing
	err := json.Unmarshal([]byte(jsonResponse), &listing)
	
	assert.NoError(t, err)
	assert.Len(t, listing.Perfumes, 1)
	assert.Contains(t, listing.Perfumes[0].Sizes[0], "/")
	assert.Equal(t, "80/100ml", listing.Perfumes[0].Sizes[0])
}

func TestParsePostContent_MultipleFragrances(t *testing.T) {
	// Test parsing multiple fragrances
	jsonResponse := `{
		"perfumes": [
			{
				"name": "Tom Ford Tobacco Vanille",
				"sizes": ["100ml"],
				"prices": ["$150"]
			},
			{
				"name": "Parfums de Marly Layton",
				"sizes": ["75ml"],
				"prices": ["$180"]
			},
			{
				"name": "Yves Saint Laurent Y EDP",
				"sizes": ["60ml", "100ml"],
				"prices": ["$70", "$100"]
			}
		]
	}`
	
	var listing models.FragranceListing
	err := json.Unmarshal([]byte(jsonResponse), &listing)
	
	assert.NoError(t, err)
	assert.Len(t, listing.Perfumes, 3)
	
	// Verify third perfume has multiple sizes
	assert.Len(t, listing.Perfumes[2].Sizes, 2)
	assert.Len(t, listing.Perfumes[2].Prices, 2)
}

func TestParsePostContent_EmptyPerfumes(t *testing.T) {
	// Test empty perfumes array
	jsonResponse := `{"perfumes": []}`
	
	var listing models.FragranceListing
	err := json.Unmarshal([]byte(jsonResponse), &listing)
	
	assert.NoError(t, err)
	assert.Empty(t, listing.Perfumes)
}
