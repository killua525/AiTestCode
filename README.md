VPS Telegram Bot (Go)

This project is a Telegram bot for basic VPS operations and monitoring. It is intended to run on a Linux VPS with root privileges when using system update and tool installation commands.

Key features:
- Status monitoring: CPU, memory, disk, uptime
- Common ops: system update, install base tools (vim/curl/htop)
- Menu UI: separate Monitoring and Ops panels
- Admin-only access using chat ID

Configuration (environment variables):
- TELEGRAM_BOT_TOKEN: Telegram bot token
- ADMIN_CHAT_ID: Telegram chat ID allowed to use the bot (optional; if empty, any chat can use it)
- POLL_TIMEOUT_SECONDS: long-polling timeout, default 30

How to run:
- Install Go 1.21+
- Set the environment variables above
- Option A (go install): go install github.com/killua525/AiTestCode/cmd/bot@latest
- Run: bot
- Option B (local dev):
	- go mod tidy
	- go run ./cmd/bot

Notes:
- Operations use apt-get and require root permissions.
- Replace the base tools list in internal/ops/ops.go if you need different packages.
