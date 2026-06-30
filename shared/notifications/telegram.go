package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var errMissingBotToken = errors.New("telegram bot token is required")
var errMissingChatID = errors.New("telegram chat ID is required")

type TelegramConfig struct {
	BotToken     string `json:"botToken"`
	ChatID       string `json:"chatId"`
	SendSilently bool   `json:"sendSilently"`
}

type TelegramSender struct {
	httpClient *http.Client
}

func NewTelegramSender() *TelegramSender {
	return &TelegramSender{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var tgResp telegramResponse
	if err := json.Unmarshal(respBody, &tgResp); err != nil {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get updates: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("telegram API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var updates telegramGetUpdatesResponse
	if err := json.Unmarshal(respBody, &updates); err != nil {
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
