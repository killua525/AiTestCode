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
		if update.Message == nil {
			continue
		}
		b.handleMessage(update.Message)
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
		b.replyWithKeyboard(message.Chat.ID, mainMenuText(), mainReplyKeyboard(), message.MessageID)
	case strings.HasPrefix(text, "/help"):
		b.replyWithKeyboard(message.Chat.ID, helpText(), mainReplyKeyboard(), message.MessageID)
	case strings.HasPrefix(text, "/monitor"):
		b.handleStatus(message.Chat.ID, message.MessageID)
	case strings.HasPrefix(text, "/ops"):
		b.replyWithKeyboard(message.Chat.ID, "*运维面板*", opsReplyKeyboard(), message.MessageID)
	case strings.HasPrefix(text, "/install"):
		b.replyWithKeyboard(message.Chat.ID, "*安装工具*", installReplyKeyboard(), message.MessageID)
	case strings.HasPrefix(text, "/uninstall"):
		b.replyWithKeyboard(message.Chat.ID, "*卸载工具*", uninstallReplyKeyboard(), message.MessageID)
	case strings.HasPrefix(text, "/back"):
		b.replyWithKeyboard(message.Chat.ID, "*主菜单*", mainReplyKeyboard(), message.MessageID)
	case strings.HasPrefix(text, "/install_tools"):
		b.handleInstallTools(message.Chat.ID, message.MessageID)
	case strings.HasPrefix(text, "/list_tools"):
		b.handleListTools(message.Chat.ID, message.MessageID)
	case strings.HasPrefix(text, "/uninstall_tools"):
		b.handleUninstallTools(message.Chat.ID, message.MessageID)
	case strings.HasPrefix(text, "/list_uninstall"):
		b.handleListUninstallTools(message.Chat.ID, message.MessageID)
	default:
		b.replyWithKeyboard(message.Chat.ID, "Unknown command. Use /help", mainReplyKeyboard(), message.MessageID)
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

func (b *Bot) replyWithKeyboard(chatID int64, text string, keyboard tgbotapi.ReplyKeyboardMarkup, replyTo int) {
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

func (b *Bot) handleUninstallTools(chatID int64, replyTo int) {
	tools := strings.Join(ops.BaseTools(), ", ")
	if tools == "" {
		tools = "(empty)"
	}
	b.reply(chatID, fmt.Sprintf("Uninstalling base tools: %s", tools), replyTo)
	out, err := ops.UninstallBaseTools()
	if err != nil {
		b.reply(chatID, fmt.Sprintf("Uninstall failed: %v\n%s", err, out), replyTo)
		return
	}
	b.reply(chatID, "Uninstall finished.", replyTo)
}

func (b *Bot) handleListUninstallTools(chatID int64, replyTo int) {
	tools := ops.BaseTools()
	if len(tools) == 0 {
		b.reply(chatID, "Uninstall tools list is empty.", replyTo)
		return
	}
	b.reply(chatID, fmt.Sprintf("Uninstall tools: %s", strings.Join(tools, ", ")), replyTo)
}

func helpText() string {
	return strings.Join([]string{
		"*VPS Bot Commands*",
		"/monitor - monitoring panel",
		"/ops - ops panel",
		"/install - install menu",
		"/uninstall - uninstall menu",
		"/install_tools - install base tools",
		"/list_tools - show install tools list",
		"/uninstall_tools - uninstall base tools",
		"/list_uninstall - show uninstall tools list",
	}, "\n")
}

func mainMenuText() string {
	return strings.Join([]string{
		"*VPS 管理机器人*",
		"请选择功能模块：",
	}, "\n")
}

func mainReplyKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/monitor 📈"),
			tgbotapi.NewKeyboardButton("/ops 🛠️"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/help ❓"),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false
	return keyboard
}

func opsReplyKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/install 📦"),
			tgbotapi.NewKeyboardButton("/uninstall 🗑️"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/back ⬅️"),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false
	return keyboard
}

func installReplyKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/install_tools ✅"),
			tgbotapi.NewKeyboardButton("/list_tools 📋"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/back ⬅️"),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false
	return keyboard
}

func uninstallReplyKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/uninstall_tools 🗑️"),
			tgbotapi.NewKeyboardButton("/list_uninstall 📋"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/back ⬅️"),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false
	return keyboard
}
