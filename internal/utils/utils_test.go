package utils

import (
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
