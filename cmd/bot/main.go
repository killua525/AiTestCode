package main

import (
	"log"
	"os"

	"github.com/killua525/AiTestCode/internal/bot"
	"github.com/killua525/AiTestCode/internal/config"
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
