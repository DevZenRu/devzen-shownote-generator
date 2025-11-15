package utils

import (
	"strings"
	"testing"
)

func TestExtractURLs(t *testing.T) {
	trelloDesc := " --- Hello  foo bar, test - " +
		"https://www.theengineeringmanager.com/management-101/performance-reviews/ " +
		"https://blog.pragmaticengineer.com/performance-reviews-for-software-engineers/ "

	actualURLs := ExtractURLs(trelloDesc)

	if len(actualURLs) != 2 {
		t.Errorf("Expected 2 URLs, got %d", len(actualURLs))
	}

	expected1 := "https://www.theengineeringmanager.com/management-101/performance-reviews/"
	if actualURLs[0] != expected1 {
		t.Errorf("Expected first URL to be %s, got %s", expected1, actualURLs[0])
	}

	expected2 := "https://blog.pragmaticengineer.com/performance-reviews-for-software-engineers/"
	if actualURLs[1] != expected2 {
		t.Errorf("Expected second URL to be %s, got %s", expected2, actualURLs[1])
	}
}

func TestUniqueStrings(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b", "d"}
	result := UniqueStrings(input)

	if len(result) != 4 {
		t.Errorf("Expected 4 unique strings, got %d", len(result))
	}

	// Check all expected values are present
	expected := map[string]bool{"a": true, "b": true, "c": true, "d": true}
	for _, str := range result {
		if !expected[str] {
			t.Errorf("Unexpected string in result: %s", str)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	// Test 1 hour, 23 minutes, 45 seconds
	startMs := int64(1000000000)
	endMs := startMs + (1*3600+23*60+45)*1000

	result := FormatDuration(startMs, endMs)
	expected := "01:23:45"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestRemoveHTMLEntities(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ndash entity",
			input:    "Test &ndash; value",
			expected: "Test - value",
		},
		{
			name:     "mdash entity",
			input:    "Test &mdash; value",
			expected: "Test - value",
		},
		{
			name:     "hellip entity",
			input:    "Test&hellip;",
			expected: "Test...",
		},
		{
			name:     "nbsp entity",
			input:    "Test&nbsp;value",
			expected: "Test value",
		},
		{
			name:     "quote entities",
			input:    "&ldquo;Test&rdquo; and &lsquo;value&rsquo;",
			expected: "\"Test\" and 'value'",
		},
		{
			name:     "multiple entities",
			input:    "Language Models &amp; Agentic AI &mdash; Course",
			expected: "Language Models &amp; Agentic AI - Course",
		},
		{
			name:     "no entities",
			input:    "Test value",
			expected: "Test value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeHTMLEntities(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTitleOfWithMeridianURL(t *testing.T) {
	url := "https://www.meridiancambridge.org/language-models-course"
	title := TitleOf(url)

	t.Logf("Title: %q", title)

	// Check that mdash entity is not present
	if strings.Contains(title, "&mdash;") {
		t.Errorf("Title contains &mdash; entity: %s", title)
	}

	// Check that ndash entity is not present
	if strings.Contains(title, "&ndash;") {
		t.Errorf("Title contains &ndash; entity: %s", title)
	}

	// Should contain a regular dash instead
	if !strings.Contains(title, "-") && !strings.Contains(title, "—") {
		t.Logf("Warning: Title doesn't contain dash or em-dash character")
	}
}
