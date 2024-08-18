package main

import (
	"testing"
	"time"

	"github.com/justinbachtell/quote-table-go/internal/assert"
)

// Test the humanDate function
func TestHumanDate(t *testing.T) {
	// Create a slice of anonymous structs with input and expected output
	tests := []struct {
		name string
		tm time.Time
		want string
	}{

		{
			name: "UTC",
			tm: time.Date(2024, 3, 17, 10, 15, 0, 0, time.UTC),
			want: "17 Mar 2024 at 10:15",
		},
		{
			name: "Empty",
			tm: time.Time{},
			want: "",
		},
		{
			name: "CET",
			tm: time.Date(2024, 3, 17, 10, 15, 0, 0, time.FixedZone("CET", 1*60*60)),
			want: "17 Mar 2024 at 09:15",
		},
		{
			name: "UTC-6",
			tm: time.Date(2024, 3, 17, 10, 15, 0, 0, time.FixedZone("UTC-6", -6*60*60)),
			want: "17 Mar 2024 at 16:15",
		},
	}

	// Loop over the test cases
	for _, tt := range tests {
		// Run a sub-test for each case
		t.Run(tt.name, func(t *testing.T) {
			hd := humanDate(tt.tm)
			
			// Compare the actual and expected values
			assert.Equal(t, hd, tt.want)
		})
	}
}