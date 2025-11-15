package handlers

import (
	"bytes"
	"net/url"
	"strings"
	"testing"
)

func TestGenerateHTMLWithMultipleURLsHasRealNewlines(t *testing.T) {
	h := &Handler{}
	themes := []Theme{
		{
			Title:             "Test Theme",
			URLs:              []string{"https://example.com/1", "https://example.com/2"},
			ReadableStartTime: "00:10:00",
			RelativeStartMs:   600000,
		},
	}

	html := h.generateHTML(themes)

	// Check that we have actual newline bytes (0x0a), not literal \n
	if strings.Contains(html, "\\n") {
		t.Errorf("HTML contains literal \\n instead of actual newlines")
	}

	// Verify actual newline bytes exist
	if !bytes.Contains([]byte(html), []byte{0x0a}) {
		t.Errorf("HTML should contain actual newline bytes")
	}

	// Count actual newlines
	newlineCount := strings.Count(html, "\n")
	if newlineCount < 5 {
		t.Errorf("Expected at least 5 newlines, got %d", newlineCount)
	}

	t.Logf("HTML output:\n%s", html)
	t.Logf("HTML bytes: %q", html)
}

func TestBuildTelegramURL(t *testing.T) {
	message := `Книжный Клуб. Глава 4 книги "The Manager's Path". В этот раз кратко.
https://amzn.to/2Jnkbs2
https://medium.com/@mrabkin/the-art-of-the-awkward-1-1-f4e1dcbd1c5c
https://noidea.dog/glue
https://www.youtube.com/watch?v=DTAXQNJLskk
https://www.amazon.co.uk/Making-Manager-What-Everyone-Looks/dp/0753552892/
https://www.nonviolentcommunication.com/`

	botToken := "123"
	channel := "@devzen_live"

	result := "https://api.telegram.org/bot" + botToken + "/sendMessage?chat_id=" +
		url.QueryEscape(channel) + "&text=" + url.QueryEscape(message)

	// Check that the URL is properly formatted
	if !strings.HasPrefix(result, "https://api.telegram.org/bot123/sendMessage") {
		t.Errorf("URL should start with telegram API endpoint")
	}

	// Check that URL encoding is applied
	if !strings.Contains(result, "%D0%9A%D0%BD%D0%B8%D0%B6") {
		t.Errorf("Cyrillic characters should be URL encoded")
	}

	// Check that special characters are encoded
	if strings.Contains(result, " ") {
		t.Errorf("Spaces should be URL encoded")
	}
}

func TestGenerateHTMLWithNoURLs(t *testing.T) {
	h := &Handler{}
	themes := []Theme{
		{
			Title:             "Test Theme",
			URLs:              []string{},
			ReadableStartTime: "00:10:00",
			RelativeStartMs:   600000,
		},
	}

	html := h.generateHTML(themes)

	if !strings.Contains(html, "Test Theme") {
		t.Error("HTML should contain theme title")
	}

	if !strings.Contains(html, "00:10:00") {
		t.Error("HTML should contain timestamp")
	}

	if strings.Contains(html, "<a href") {
		t.Error("HTML should not contain links when no URLs")
	}
}

func TestGenerateHTMLWithOneURL(t *testing.T) {
	h := &Handler{}
	themes := []Theme{
		{
			Title:             "Test Theme",
			URLs:              []string{"https://example.com"},
			ReadableStartTime: "00:10:00",
			RelativeStartMs:   600000,
		},
	}

	html := h.generateHTML(themes)

	if !strings.Contains(html, `<a href="https://example.com">Test Theme</a>`) {
		t.Error("HTML should contain link to single URL with theme title as text")
	}
}

func TestGenerateHTMLWithMultipleURLs(t *testing.T) {
	h := &Handler{}
	themes := []Theme{
		{
			Title:             "Test Theme",
			URLs:              []string{"https://example1.com", "https://example2.com"},
			ReadableStartTime: "00:10:00",
			RelativeStartMs:   600000,
		},
	}

	html := h.generateHTML(themes)

	if !strings.Contains(html, "Test Theme") {
		t.Error("HTML should contain theme title")
	}

	// Should create nested list
	if !strings.Contains(html, "<ul>") || strings.Count(html, "<ul>") < 2 {
		t.Error("HTML should contain nested list for multiple URLs")
	}

	if !strings.Contains(html, "https://example1.com") {
		t.Error("HTML should contain first URL")
	}

	if !strings.Contains(html, "https://example2.com") {
		t.Error("HTML should contain second URL")
	}
}
