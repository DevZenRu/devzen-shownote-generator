package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"devzen-shownote-generator/internal/config"
	"devzen-shownote-generator/internal/utils"
)

// Theme represents a podcast theme/topic with timestamp and URLs
type Theme struct {
	Title             string
	URLs              []string
	ReadableStartTime string
	RelativeStartMs   int64
}

// Handler holds the configuration and provides HTTP handlers
type Handler struct {
	cfg *config.Config
}

// New creates a new Handler with the given configuration
func New(cfg *config.Config) *Handler {
	return &Handler{cfg: cfg}
}

// HandleGenerate handles GET /generate requests
func (h *Handler) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse optional manual start time
	var manualStartMs int64
	startParam := r.URL.Query().Get("start")
	if startParam != "" {
		parsed, err := strconv.ParseInt(startParam, 10, 64)
		if err != nil {
			http.Error(w, "Invalid start parameter", http.StatusBadRequest)
			return
		}
		manualStartMs = parsed
	}

	themes, err := h.generateShownotes(manualStartMs)
	if err != nil {
		log.Printf("Error generating shownotes: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html := h.generateHTML(themes)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// HandleTrelloHook handles GET and POST /trellohook requests
func (h *Handler) HandleTrelloHook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Webhook verification
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}
	defer r.Body.Close()

	log.Println(string(body))

	// Parse webhook event
	var event map[string]interface{}
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error parsing webhook JSON: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check if it's an updateCard action
	action, ok := event["action"].(map[string]interface{})
	if !ok {
		w.WriteHeader(http.StatusOK)
		return
	}

	actionType, ok := action["type"].(string)
	if !ok || actionType != "updateCard" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check if card moved to "In Discussion"
	data, ok := action["data"].(map[string]interface{})
	if !ok {
		w.WriteHeader(http.StatusOK)
		return
	}

	listBefore, _ := data["listBefore"].(map[string]interface{})
	listAfter, _ := data["listAfter"].(map[string]interface{})
	listBeforeID, _ := listBefore["id"].(string)
	listAfterID, _ := listAfter["id"].(string)

	if listBeforeID == h.cfg.TrelloToDiscussListID && listAfterID == h.cfg.TrelloInDiscussionListID {
		card, _ := data["card"].(map[string]interface{})
		title, _ := card["name"].(string)
		cardID, _ := card["id"].(string)

		urls, err := h.getThemeURLsByCardID(cardID)
		if err != nil {
			log.Printf("Error fetching card URLs: %v", err)
		} else {
			if err := h.postMessageToTelegram(title, urls); err != nil {
				log.Printf("Error posting to Telegram: %v", err)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// generateShownotes fetches cards from Trello and generates themes with timestamps
func (h *Handler) generateShownotes(manualStartMs int64) ([]Theme, error) {
	// Fetch discussed cards
	discussedURL := h.buildTrelloURL(fmt.Sprintf("/1/lists/%s", h.cfg.TrelloDiscussedListID),
		map[string]string{
			"fields":      "name",
			"cards":       "open",
			"card_fields": "name,desc",
		})

	var response map[string]interface{}
	if err := h.fetchJSON(discussedURL, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch discussed cards: %w", err)
	}

	cards, ok := response["cards"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: cards not found")
	}

	// Get recording start time
	recordingStartMs, err := h.getTimestampOfRecordingStartedEvent(manualStartMs)
	if err != nil {
		return nil, fmt.Errorf("failed to get recording start time: %w", err)
	}

	// Process each card
	var themes []Theme
	for _, cardData := range cards {
		card, ok := cardData.(map[string]interface{})
		if !ok {
			continue
		}

		cardID, _ := card["id"].(string)
		name, _ := card["name"].(string)
		desc, _ := card["desc"].(string)

		urls := utils.ExtractURLs(desc)
		urls = utils.UniqueStrings(urls)

		// Get timestamp when theme started
		themeStartMs, err := h.getTimestampOfThemeStartedEvent(cardID)
		if err != nil {
			log.Printf("WARN - %v", err)
			themes = append(themes, Theme{
				Title:             name,
				URLs:              urls,
				ReadableStartTime: "TIMESTAMP_IS_MISSING",
				RelativeStartMs:   0,
			})
			continue
		}

		relativeStartStr := utils.FormatDuration(recordingStartMs, themeStartMs)
		themes = append(themes, Theme{
			Title:             name,
			URLs:              urls,
			ReadableStartTime: relativeStartStr,
			RelativeStartMs:   themeStartMs,
		})
	}

	// Sort by timestamp
	sort.Slice(themes, func(i, j int) bool {
		return themes[i].RelativeStartMs <= themes[j].RelativeStartMs
	})

	return themes, nil
}

// getTimestampOfRecordingStartedEvent gets the recording start timestamp
func (h *Handler) getTimestampOfRecordingStartedEvent(manualStartMs int64) (int64, error) {
	if manualStartMs != 0 {
		return manualStartMs, nil
	}

	cardURL := h.buildTrelloURL(fmt.Sprintf("/1/cards/%s", h.cfg.TrelloRecordingStartedCardID), nil)
	var card map[string]interface{}
	if err := h.fetchJSON(cardURL, &card); err != nil {
		return 0, fmt.Errorf("failed to fetch recording card: %w", err)
	}

	dateStr, ok := card["dateLastActivity"].(string)
	if !ok {
		return 0, fmt.Errorf("dateLastActivity not found in recording card")
	}

	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse date: %w", err)
	}

	return t.UnixMilli(), nil
}

// getTimestampOfThemeStartedEvent gets the timestamp when a card moved to "In Discussion"
func (h *Handler) getTimestampOfThemeStartedEvent(cardID string) (int64, error) {
	actionsURL := h.buildTrelloURL(fmt.Sprintf("/1/cards/%s/actions", cardID), nil)
	var actions []interface{}
	if err := h.fetchJSON(actionsURL, &actions); err != nil {
		return 0, fmt.Errorf("failed to fetch card actions: %w", err)
	}

	var possibleTimestamps []int64
	for _, actionData := range actions {
		action, ok := actionData.(map[string]interface{})
		if !ok {
			continue
		}

		actionType, _ := action["type"].(string)
		if actionType != "updateCard" {
			continue
		}

		data, ok := action["data"].(map[string]interface{})
		if !ok {
			continue
		}

		listAfter, _ := data["listAfter"].(map[string]interface{})
		listBefore, _ := data["listBefore"].(map[string]interface{})
		listAfterID, _ := listAfter["id"].(string)
		listBeforeID, _ := listBefore["id"].(string)

		// Check if moved from "To Discuss" or "Backlog" to "In Discussion"
		if (listBeforeID == h.cfg.TrelloToDiscussListID && listAfterID == h.cfg.TrelloInDiscussionListID) ||
			(listBeforeID == h.cfg.TrelloBacklogListID && listAfterID == h.cfg.TrelloInDiscussionListID) {

			dateStr, _ := action["date"].(string)
			t, err := time.Parse(time.RFC3339, dateStr)
			if err != nil {
				continue
			}
			possibleTimestamps = append(possibleTimestamps, t.UnixMilli())
		}
	}

	if len(possibleTimestamps) == 0 {
		return 0, fmt.Errorf("can't find card movements of %s 'to discuss' -> 'in discussion'. No timestamp found", cardID)
	}

	if len(possibleTimestamps) > 1 {
		log.Printf("WARN - More than one card movement of %s 'to discuss' -> 'in discussion'. Using the latest timestamp.", cardID)
		// Find max timestamp
		maxTs := possibleTimestamps[0]
		for _, ts := range possibleTimestamps[1:] {
			if ts > maxTs {
				maxTs = ts
			}
		}
		return maxTs, nil
	}

	return possibleTimestamps[0], nil
}

// generateHTML generates the HTML output for themes
func (h *Handler) generateHTML(themes []Theme) string {
	var sb strings.Builder
	sb.WriteString("<ul>\n")

	for _, theme := range themes {
		escapedTitle := html.EscapeString(theme.Title)
		sb.WriteString(fmt.Sprintf("<li>[%s] ", theme.ReadableStartTime))

		switch len(theme.URLs) {
		case 0:
			sb.WriteString(fmt.Sprintf("%s</li>\n", escapedTitle))
		case 1:
			sb.WriteString(fmt.Sprintf(`<a href="%s">%s</a></li>\n`, theme.URLs[0], escapedTitle))
		default:
			sb.WriteString(escapedTitle + "\n")
			sb.WriteString("<ul>\n")
			for _, url := range theme.URLs {
				urlTitle := utils.TitleOf(url)
				sb.WriteString(fmt.Sprintf(`<li><a href="%s">%s</a></li>\n`, url, urlTitle))
			}
			sb.WriteString("</ul>\n")
			sb.WriteString("</li>\n")
		}
	}

	sb.WriteString("</ul>\n")
	return sb.String()
}
