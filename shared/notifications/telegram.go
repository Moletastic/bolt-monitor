package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"bolt-monitor/shared/outboundhttp"
)

var errMissingBotToken = errors.New("telegram bot token is required")
var errMissingChatID = errors.New("telegram chat ID is required")

type TelegramConfig struct {
	BotToken     string `json:"botToken"`
	ChatID       string `json:"chatId"`
	SendSilently bool   `json:"sendSilently"`
}

type TelegramSender struct {
	executor HTTPExecutor
}

func NewTelegramSender(executors ...HTTPExecutor) *TelegramSender {
	return &TelegramSender{executor: senderExecutor(executors)}
}

func (s *TelegramSender) Send(ctx context.Context, notification Notification) error {
	var config TelegramConfig
	if len(notification.Config) > 0 {
		if err := json.Unmarshal(notification.Config, &config); err != nil {
			return fmt.Errorf("invalid telegram config: %w", err)
		}
	}

	if strings.TrimSpace(config.BotToken) == "" {
		return errMissingBotToken
	}
	if strings.TrimSpace(config.ChatID) == "" {
		return errMissingChatID
	}

	escapedText := escapeMarkdownV2(notification.Message)

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.BotToken)
	payload := map[string]interface{}{
		"chat_id":    config.ChatID,
		"text":       escapedText,
		"parse_mode": "MarkdownV2",
	}
	if config.SendSilently {
		payload["disable_notification"] = true
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	response, err := s.executor.Execute(ctx, outboundhttp.Request{Method: http.MethodPost, URL: apiURL, Header: http.Header{"Content-Type": {"application/json"}}, Body: bytes.NewReader(body), Timeout: outboundhttp.NotificationTimeout})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		var tgResp telegramResponse
		if err := json.Unmarshal(response.Body, &tgResp); err == nil && strings.Contains(strings.ToLower(tgResp.Description), "chat not found") {
			return errors.New("telegram chat not found: use the numeric chat ID and make sure the bot has access to that chat")
		}
		return errors.New("telegram API returned non-success status")
	}

	var tgResp telegramResponse
	if err := json.Unmarshal(response.Body, &tgResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	if !tgResp.Ok {
		return fmt.Errorf("telegram error: %s", tgResp.Description)
	}

	return nil
}

func (s *TelegramSender) ChannelType() string {
	return "telegram"
}

func (s *TelegramSender) ValidateConfig(config json.RawMessage) error {
	if len(config) == 0 {
		return errors.New("config is required")
	}
	var cfg TelegramConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if strings.TrimSpace(cfg.BotToken) == "" {
		return errMissingBotToken
	}
	return nil
}

type telegramResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
}

var markdownV2Escaper = regexp.MustCompile(`[_*\[\]()~>#+\-=|{}.!\\]`)

func escapeMarkdownV2(text string) string {
	return markdownV2Escaper.ReplaceAllString(text, `\$0`)
}

func (s *TelegramSender) DetectChatID(ctx context.Context, botToken string) (string, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", botToken)

	response, err := s.executor.Execute(ctx, outboundhttp.Request{Method: http.MethodGet, URL: apiURL, Timeout: outboundhttp.NotificationTimeout})
	if err != nil {
		return "", fmt.Errorf("failed to get updates: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return "", errors.New("telegram API returned non-success status")
	}

	var updates telegramGetUpdatesResponse
	if err := json.Unmarshal(response.Body, &updates); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(updates.Result) == 0 {
		return "", errors.New("no updates found - send a message to your bot first")
	}

	chat := updates.Result[0].Message.Chat
	return fmt.Sprintf("%d", chat.ID), nil
}

type telegramGetUpdatesResponse struct {
	Ok     bool `json:"ok"`
	Result []struct {
		Message struct {
			Chat struct {
				ID int `json:"id"`
			} `json:"chat"`
		} `json:"message"`
	} `json:"result"`
}
