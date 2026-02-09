package main

import (
	"log"
	"os"

	"vps-telegram-bot/internal/bot"
	"vps-telegram-bot/internal/config"
)

func main() {
	cfg := config.Load()
	logger := log.New(os.Stdout, "", log.LstdFlags)

	b, err := bot.New(cfg, logger)
	if err != nil {
		logger.Fatalf("init bot: %v", err)
	}

	if err := b.Run(); err != nil {
		logger.Fatalf("run bot: %v", err)
	}
}
