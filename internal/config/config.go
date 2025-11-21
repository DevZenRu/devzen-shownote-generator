package config

import (
	"log"
	"os"
)

// Config holds all configuration values loaded from environment variables
type Config struct {
	// Trello configuration
	TrelloAppKey                 string
	TrelloReadToken              string
	TrelloDiscussedListID        string
	TrelloToDiscussListID        string
	TrelloInDiscussionListID     string
	TrelloRecordingStartedCardID string
	TrelloBacklogListID          string

	// Telegram configuration
	TelegramBotToken      string
	TelegramDevZenChannel string

	// Server configuration
	Port string
}

// Load reads configuration from environment variables and panics if required values are missing
func Load() *Config {
	cfg := &Config{
		TrelloAppKey:                 getEnvOrPanic("TRELLO_APP_KEY"),
		TrelloReadToken:              getEnvOrPanic("TRELLO_READ_TOKEN"),
		TrelloDiscussedListID:        getEnvOrPanic("TRELLO_DISCUSSED_LIST_ID"),
		TrelloToDiscussListID:        getEnvOrPanic("TRELLO_TO_DISCUSS_LIST_ID"),
		TrelloInDiscussionListID:     getEnvOrPanic("TRELLO_IN_DISCUSSION_LIST_ID"),
		TrelloRecordingStartedCardID: getEnvOrPanic("TRELLO_RECORDING_STARTED_CARD_ID"),
		TrelloBacklogListID:          getEnvOrPanic("TRELLO_BACKLOG_LIST_ID"),
		TelegramBotToken:             getEnvOrPanic("TELEGRAM_BOT_TOKEN"),
		TelegramDevZenChannel:        "@devzen_live",
		Port:                         getEnvOrDefault("PORT", "9025"),
	}

	return cfg
}

func getEnvOrPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("ERROR: Required environment variable %s is not set", key)
	}
	return value
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
