package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"devzen-shownote-generator/internal/utils"
)

const (
	httpTimeout = 10 * time.Second
)

// Trello API response structures
type TrelloListResponse struct {
	Cards []TrelloCard `json:"cards"`
}

type TrelloCard struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Desc             string `json:"desc"`
	DateLastActivity string `json:"dateLastActivity"`
}

type TrelloAction struct {
	Type string           `json:"type"`
	Date string           `json:"date"`
	Data TrelloActionData `json:"data"`
}

type TrelloActionData struct {
	Card       *TrelloCardRef `json:"card,omitempty"`
	ListBefore *TrelloListRef `json:"listBefore,omitempty"`
	ListAfter  *TrelloListRef `json:"listAfter,omitempty"`
}

type TrelloCardRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TrelloListRef struct {
	ID string `json:"id"`
}

// Trello webhook event structure
type TrelloWebhookEvent struct {
	Action TrelloWebhookAction `json:"action"`
}

type TrelloWebhookAction struct {
	Type string           `json:"type"`
	Data TrelloActionData `json:"data"`
}

// buildTrelloURL builds a Trello API URL with authentication
func (h *Handler) buildTrelloURL(path string, params map[string]string) string {
	u, _ := url.Parse("https://api.trello.com" + path)
	q := u.Query()
	q.Set("key", h.cfg.TrelloAppKey)
	q.Set("token", h.cfg.TrelloReadToken)

	for k, v := range params {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// fetchJSON performs an HTTP GET and unmarshals JSON response
func (h *Handler) fetchJSON(urlStr string, target interface{}) error {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// getThemeURLsByCardID fetches URLs from a card's description
func (h *Handler) getThemeURLsByCardID(cardID string) ([]string, error) {
	cardURL := h.buildTrelloURL(fmt.Sprintf("/1/cards/%s", cardID), nil)
	var card TrelloCard
	if err := h.fetchJSON(cardURL, &card); err != nil {
		return nil, err
	}

	return utils.ExtractURLsUnique(card.Desc), nil
}

// postMessageToTelegram sends a message to the Telegram channel
func (h *Handler) postMessageToTelegram(title string, urls []string) error {
	urlsStr := strings.Join(urls, "\n")
	message := fmt.Sprintf("%s\n%s", title, urlsStr)

	telegramURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s",
		h.cfg.TelegramBotToken,
		url.QueryEscape(h.cfg.TelegramDevZenChannel),
		url.QueryEscape(message))

	log.Println("sendMessageToTelegramChannel URL: " + telegramURL)

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(telegramURL)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Can't send message to Telegram. Telegram response: %s", string(body))
		return fmt.Errorf("telegram returned status %d", resp.StatusCode)
	}

	log.Println("Sent message to Telegram successfully")
	return nil
}
