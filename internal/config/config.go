package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	BotToken    string
	AdminChatID int64
	PollTimeout time.Duration
}

func Load() Config {
	cfg := Config{
		BotToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		AdminChatID: parseInt64(os.Getenv("ADMIN_CHAT_ID")),
		PollTimeout: 30 * time.Second,
	}

	if v := os.Getenv("POLL_TIMEOUT_SECONDS"); v != "" {
		if seconds, err := strconv.Atoi(v); err == nil && seconds > 0 {
			cfg.PollTimeout = time.Duration(seconds) * time.Second
		}
	}

	return cfg
}

func parseInt64(v string) int64 {
	if v == "" {
		return 0
	}
	val, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return val
}
