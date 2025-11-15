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
		// Unescape HTML entities first (they may be present in the original title tag)
		title = html.UnescapeString(title)
		// Then escape for safe HTML output
		title = html.EscapeString(title)
		// Finally remove problematic entities that break RSS feeds
		title = removeHTMLEntities(title)
		return title
	}

	return url
}

// removeHTMLEntities removes HTML entities that break RSS feeds
func removeHTMLEntities(s string) string {
	s = strings.ReplaceAll(s, "&ndash;", "-")
	s = strings.ReplaceAll(s, "&mdash;", "-")
	s = strings.ReplaceAll(s, "&hellip;", "...")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&rsquo;", "'")
	s = strings.ReplaceAll(s, "&lsquo;", "'")
	s = strings.ReplaceAll(s, "&rdquo;", `"`)
	s = strings.ReplaceAll(s, "&ldquo;", `"`)
	// Also replace the actual unicode characters
	s = strings.ReplaceAll(s, "\u2014", "-")   // em dash —
	s = strings.ReplaceAll(s, "\u2013", "-")   // en dash –
	s = strings.ReplaceAll(s, "\u2026", "...") // ellipsis …
	s = strings.ReplaceAll(s, "\u00a0", " ")   // non-breaking space
	s = strings.ReplaceAll(s, "\u2019", "'")   // right single quote '
	s = strings.ReplaceAll(s, "\u2018", "'")   // left single quote '
	s = strings.ReplaceAll(s, "\u201d", `"`)   // right double quote "
	s = strings.ReplaceAll(s, "\u201c", `"`)   // left double quote "
	return s
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
