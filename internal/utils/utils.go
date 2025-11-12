package utils

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	maxDownloadSize = 10 * 1024 * 1024 // 10MB
	httpTimeout     = 10 * time.Second
)

var (
	urlRegex   = regexp.MustCompile(`(?i)(https?|ftp|gopher|telnet|file):((//)|(\\\\))+[\w\d:#@%/;$()~_?\+-=\\\\.&]*`)
	titleRegex = regexp.MustCompile(`(?i)<\s*title.*?>(.*?)</title>`)
)

// ExtractURLs extracts all URLs from the given text
func ExtractURLs(text string) []string {
	return urlRegex.FindAllString(text, -1)
}

// TitleOf fetches the HTML title from the given URL
// Returns the URL itself if fetching fails or no title is found
// Limits download to 10MB and uses 10s timeout
func TitleOf(url string) string {
	client := &http.Client{
		Timeout: httpTimeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return url
	}
	defer resp.Body.Close()

	// Limit reading to maxDownloadSize
	limitedReader := io.LimitReader(resp.Body, maxDownloadSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return url
	}

	matches := titleRegex.FindSubmatch(body)
	if len(matches) > 1 {
		title := string(matches[1])
		// Remove newlines and excessive whitespace
		title = strings.ReplaceAll(title, "\n", " ")
		title = strings.ReplaceAll(title, "\r", " ")
		title = strings.TrimSpace(title)
		return html.EscapeString(title)
	}

	return url
}

// FormatDuration formats milliseconds duration as HH:mm:ss
func FormatDuration(startMs, endMs int64) string {
	duration := time.Duration(endMs-startMs) * time.Millisecond
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// UniqueStrings returns a slice with duplicate strings removed
func UniqueStrings(input []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, str := range input {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}

	return result
}
