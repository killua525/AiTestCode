package bot

import (
	"errors"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/killua525/AiTestCode/internal/config"
	"github.com/killua525/AiTestCode/internal/monitor"
	"github.com/killua525/AiTestCode/internal/ops"
)

type Bot struct {
	cfg    config.Config
	logger *log.Logger
	api    *tgbotapi.BotAPI
}

func New(cfg config.Config, logger *log.Logger) (*Bot, error) {
	if cfg.BotToken == "" {
		return nil, errors.New("TELEGRAM_BOT_TOKEN is required")
	}

	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}
	api.Debug = false

	return &Bot{
		cfg:    cfg,
		logger: logger,
		api:    api,
	}, nil
}

func (b *Bot) Run() error {
	b.logger.Printf("authorized on account %s", b.api.Self.UserName)

	updateCfg := tgbotapi.NewUpdate(0)
	updateCfg.Timeout = int(b.cfg.PollTimeout.Seconds())

	updates := b.api.GetUpdatesChan(updateCfg)
	for update := range updates {
		switch {
		case update.Message != nil:
			b.handleMessage(update.Message)
		case update.CallbackQuery != nil:
			b.handleCallback(update.CallbackQuery)
		}
	}

	return nil
}

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	if !b.isAllowed(message.Chat.ID) {
		b.reply(message.Chat.ID, "Unauthorized", message.MessageID)
		return
	}

	text := strings.TrimSpace(message.Text)
	switch {
	case strings.HasPrefix(text, "/start"):
		b.replyWithMenu(message.Chat.ID, mainMenuText(), mainMenuKeyboard(), message.MessageID)
	case strings.HasPrefix(text, "/help"):
		b.replyWithMenu(message.Chat.ID, helpText(), mainMenuKeyboard(), message.MessageID)
	case strings.HasPrefix(text, "/monitor"):
		b.replyWithMenu(message.Chat.ID, "*监控面板*", monitorKeyboard(), message.MessageID)
	case strings.HasPrefix(text, "/ops"):
		b.replyWithMenu(message.Chat.ID, "*运维面板*", opsKeyboard(), message.MessageID)
	case strings.HasPrefix(text, "/status"):
		b.handleStatus(message.Chat.ID, message.MessageID)
	case strings.HasPrefix(text, "/install_tools"):
		b.handleInstallTools(message.Chat.ID, message.MessageID)
	case strings.HasPrefix(text, "/list_tools"):
		b.handleListTools(message.Chat.ID, message.MessageID)
	default:
		b.replyWithMenu(message.Chat.ID, "Unknown command. Use /help", mainMenuKeyboard(), message.MessageID)
	}
}

func (b *Bot) handleCallback(query *tgbotapi.CallbackQuery) {
	if !b.isAllowed(query.Message.Chat.ID) {
		b.answerCallback(query.ID, "Unauthorized")
		return
	}

	switch query.Data {
	case "menu_main":
		b.editMenu(query.Message, mainMenuText(), mainMenuKeyboard())
	case "menu_monitor":
		b.editMenu(query.Message, "*监控面板*", monitorKeyboard())
	case "menu_ops":
		b.editMenu(query.Message, "*运维面板*", opsKeyboard())
	case "monitor_status":
		b.answerCallback(query.ID, "正在获取状态...")
		b.handleStatus(query.Message.Chat.ID, 0)
	case "ops_install_tools":
		b.answerCallback(query.ID, "开始安装工具...")
		b.handleInstallTools(query.Message.Chat.ID, 0)
	case "ops_list_tools":
		b.answerCallback(query.ID, "获取工具列表...")
		b.handleListTools(query.Message.Chat.ID, 0)
	default:
		b.answerCallback(query.ID, "Unknown action")
	}
}

func (b *Bot) isAllowed(chatID int64) bool {
	if b.cfg.AdminChatID == 0 {
		return true
	}
	return chatID == b.cfg.AdminChatID
}

func (b *Bot) reply(chatID int64, text string, replyTo int) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if replyTo > 0 {
		msg.ReplyToMessageID = replyTo
	}
	if _, err := b.api.Send(msg); err != nil {
		b.logger.Printf("send message error: %v", err)
	}
}

func (b *Bot) replyWithMenu(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup, replyTo int) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if replyTo > 0 {
		msg.ReplyToMessageID = replyTo
	}
	msg.ReplyMarkup = keyboard
	if _, err := b.api.Send(msg); err != nil {
		b.logger.Printf("send message error: %v", err)
	}
}

func (b *Bot) editMenu(message *tgbotapi.Message, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	edit := tgbotapi.NewEditMessageText(message.Chat.ID, message.MessageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	if _, err := b.api.Send(edit); err != nil {
		b.logger.Printf("edit message error: %v", err)
	}
}

func (b *Bot) answerCallback(id string, text string) {
	cb := tgbotapi.NewCallback(id, text)
	if _, err := b.api.Request(cb); err != nil {
		b.logger.Printf("callback error: %v", err)
	}
}

func (b *Bot) handleStatus(chatID int64, replyTo int) {
	cpu, _ := monitor.CPUPercent()
	mem, _ := monitor.MemoryUsage()
	disk, _ := monitor.DiskUsage("/")
	uptime, _ := monitor.Uptime()

	text := fmt.Sprintf(
		"*Status*\nCPU: %s\nMemory: %s\nDisk: %s\nUptime: %s",
		cpu, mem, disk, uptime,
	)
	b.reply(chatID, text, replyTo)
}

func (b *Bot) handleCPU(chatID int64, replyTo int) {
	cpu, err := monitor.CPUPercent()
	if err != nil {
		b.reply(chatID, fmt.Sprintf("CPU error: %v", err), replyTo)
		return
	}
	b.reply(chatID, fmt.Sprintf("CPU: %s", cpu), replyTo)
}

func (b *Bot) handleMem(chatID int64, replyTo int) {
	mem, err := monitor.MemoryUsage()
	if err != nil {
		b.reply(chatID, fmt.Sprintf("Memory error: %v", err), replyTo)
		return
	}
	b.reply(chatID, fmt.Sprintf("Memory: %s", mem), replyTo)
}

func (b *Bot) handleDisk(chatID int64, replyTo int) {
	disk, err := monitor.DiskUsage("/")
	if err != nil {
		b.reply(chatID, fmt.Sprintf("Disk error: %v", err), replyTo)
		return
	}
	b.reply(chatID, fmt.Sprintf("Disk: %s", disk), replyTo)
}

func (b *Bot) handleUptime(chatID int64, replyTo int) {
	uptime, err := monitor.Uptime()
	if err != nil {
		b.reply(chatID, fmt.Sprintf("Uptime error: %v", err), replyTo)
		return
	}
	b.reply(chatID, fmt.Sprintf("Uptime: %s", uptime), replyTo)
}

func (b *Bot) handleInstallTools(chatID int64, replyTo int) {
	tools := strings.Join(ops.BaseTools(), ", ")
	if tools == "" {
		tools = "(empty)"
	}
	b.reply(chatID, fmt.Sprintf("Installing base tools: %s", tools), replyTo)
	out, err := ops.InstallBaseTools()
	if err != nil {
		b.reply(chatID, fmt.Sprintf("Install failed: %v\n%s", err, out), replyTo)
		return
	}
	b.reply(chatID, "Install finished.", replyTo)
}

func (b *Bot) handleListTools(chatID int64, replyTo int) {
	tools := ops.BaseTools()
	if len(tools) == 0 {
		b.reply(chatID, "Base tools list is empty.", replyTo)
		return
	}
	b.reply(chatID, fmt.Sprintf("Base tools: %s", strings.Join(tools, ", ")), replyTo)
}

func helpText() string {
	return strings.Join([]string{
		"*VPS Bot Commands*",
		"/monitor - monitoring panel",
		"/ops - ops panel",
		"/status - summary status",
		"/install_tools - install vim/curl/htop",
		"/list_tools - show base tools list",
	}, "\n")
}

func mainMenuText() string {
	return strings.Join([]string{
		"*VPS 管理机器人*",
		"请选择功能模块：",
	}, "\n")
}

func mainMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 监控", "menu_monitor"),
			tgbotapi.NewInlineKeyboardButtonData("🛠️ 运维", "menu_ops"),
		),
	)
}

func monitorKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧾 概览", "monitor_status"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ 返回", "menu_main"),
		),
	)
}

func opsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📦 安装基础工具", "ops_install_tools"),
			tgbotapi.NewInlineKeyboardButtonData("📋 工具列表", "ops_list_tools"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ 返回", "menu_main"),
		),
	)
}
