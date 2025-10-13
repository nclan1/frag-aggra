package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPerfume_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		perfume  Perfume
		expected string
	}{
		{
			name: "basic perfume",
			perfume: Perfume{
				Name:   "Tom Ford Tobacco Vanille",
				Sizes:  []string{"100ml"},
				Prices: []string{"$150"},
			},
			expected: `{"name":"Tom Ford Tobacco Vanille","sizes":["100ml"],"prices":["$150"]}`,
		},
		{
			name: "perfume with multiple sizes",
			perfume: Perfume{
				Name:   "Bleu de Chanel EDT",
				Sizes:  []string{"50ml", "100ml", "150ml"},
				Prices: []string{"$80", "$120", "$160"},
			},
			expected: `{"name":"Bleu de Chanel EDT","sizes":["50ml","100ml","150ml"],"prices":["$80","$120","$160"]}`,
		},
		{
			name: "perfume with partial bottle",
			perfume: Perfume{
				Name:   "Maison Francis Kurkdijan Baccarat Rouge 540",
				Sizes:  []string{"80/100ml"},
				Prices: []string{"$200"},
			},
			expected: `{"name":"Maison Francis Kurkdijan Baccarat Rouge 540","sizes":["80/100ml"],"prices":["$200"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.perfume)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))
		})
	}
}

func TestPerfume_JSONDeserialization(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected Perfume
		wantErr  bool
	}{
		{
			name:     "basic perfume",
			jsonData: `{"name":"Tom Ford Tobacco Vanille","sizes":["100ml"],"prices":["$150"]}`,
			expected: Perfume{
				Name:   "Tom Ford Tobacco Vanille",
				Sizes:  []string{"100ml"},
				Prices: []string{"$150"},
			},
			wantErr: false,
		},
		{
			name:     "perfume with multiple sizes",
			jsonData: `{"name":"Bleu de Chanel EDT","sizes":["50ml","100ml"],"prices":["$80","$120"]}`,
			expected: Perfume{
				Name:   "Bleu de Chanel EDT",
				Sizes:  []string{"50ml", "100ml"},
				Prices: []string{"$80", "$120"},
			},
			wantErr: false,
		},
		{
			name:     "invalid JSON",
			jsonData: `{"name":"Tom Ford"`,
			expected: Perfume{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var perfume Perfume
			err := json.Unmarshal([]byte(tt.jsonData), &perfume)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, perfume)
			}
		})
	}
}

func TestFragranceListing_JSONSerialization(t *testing.T) {
	listing := FragranceListing{
		Perfumes: []Perfume{
			{
				Name:   "Tom Ford Tobacco Vanille",
				Sizes:  []string{"100ml"},
				Prices: []string{"$150"},
			},
			{
				Name:   "Bleu de Chanel EDT",
				Sizes:  []string{"50ml", "100ml"},
				Prices: []string{"$80", "$120"},
			},
		},
	}

	jsonData, err := json.Marshal(listing)
	assert.NoError(t, err)

	var decoded FragranceListing
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, listing, decoded)
}

func TestPost_JSONSerialization(t *testing.T) {
	post := Post{
		PostID:         "abc123",
		URL:            "https://reddit.com/r/fragranceswap/abc123",
		Title:          "[WTS] Tom Ford Sale",
		Body:           "Selling Tom Ford fragrances",
		SellerUsername: "testuser",
	}

	jsonData, err := json.Marshal(post)
	assert.NoError(t, err)

	var decoded Post
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, post, decoded)
}

func TestPerfume_EmptyFields(t *testing.T) {
	perfume := Perfume{
		Name:   "",
		Sizes:  []string{},
		Prices: []string{},
	}

	jsonData, err := json.Marshal(perfume)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"name":""`)
	assert.Contains(t, string(jsonData), `"sizes":[]`)
	assert.Contains(t, string(jsonData), `"prices":[]`)
}

func TestFragranceListing_EmptyPerfumes(t *testing.T) {
	listing := FragranceListing{
		Perfumes: []Perfume{},
	}

	jsonData, err := json.Marshal(listing)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"perfumes":[]`)
}
