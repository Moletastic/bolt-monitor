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

func (s *TelegramSender) Send(ctx context.Context, notification Notification) (SendOutcome, error) {
	var config TelegramConfig
	if len(notification.Config) > 0 {
		if err := json.Unmarshal(notification.Config, &config); err != nil {
			return SendOutcome{Class: OutcomeInvalidConfig, Retryable: false}, fmt.Errorf("invalid telegram config: %w", err)
		}
	}

	if strings.TrimSpace(config.BotToken) == "" {
		return SendOutcome{Class: OutcomeInvalidConfig, Retryable: false}, errMissingBotToken
	}
	if strings.TrimSpace(config.ChatID) == "" {
		return SendOutcome{Class: OutcomeInvalidConfig, Retryable: false}, errMissingChatID
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
		return SendOutcome{Class: OutcomeInvalidConfig, Retryable: false}, fmt.Errorf("failed to marshal request: %w", err)
	}

	headers := http.Header{"Content-Type": {"application/json"}}
	response, err := s.executor.Execute(ctx, outboundhttp.Request{Method: http.MethodPost, URL: apiURL, Header: headers, Body: bytes.NewReader(body), Timeout: outboundhttp.NotificationTimeout})
	if err != nil {
		return classifyTransportError(err), err
	}
	if response.StatusCode == http.StatusOK {
		var tgResp telegramResponse
		if err := json.Unmarshal(response.Body, &tgResp); err != nil {
			return SendOutcome{Class: OutcomeTransport, Retryable: true}, fmt.Errorf("failed to parse response: %w", err)
		}
		if tgResp.Ok {
			return SendOutcome{Class: OutcomeAccepted, Metadata: ProviderMetadata{ProviderStatusClass: "2xx"}}, nil
		}
		return SendOutcome{Class: OutcomeProvider4xx, Retryable: false, Metadata: SafeProviderMetadata(response.StatusCode, "", 0, maxRetryAfterSeconds)}, errors.New("telegram rejected the request")
	}
	if response.StatusCode == http.StatusTooManyRequests {
		return SendOutcome{Class: OutcomeThrottled, Retryable: true, Metadata: SafeProviderMetadata(response.StatusCode, "", ParseRetryAfterSeconds(response.Header.Get("Retry-After"), maxRetryAfterSeconds), maxRetryAfterSeconds)}, nil
	}
	if response.StatusCode >= 500 {
		return SendOutcome{Class: OutcomeProvider5xx, Retryable: true, Metadata: SafeProviderMetadata(response.StatusCode, "", 0, maxRetryAfterSeconds)}, nil
	}
	var tgResp telegramResponse
	_ = json.Unmarshal(response.Body, &tgResp)
	if strings.Contains(strings.ToLower(tgResp.Description), "chat not found") {
		return SendOutcome{Class: OutcomeProvider4xx, Retryable: false, Metadata: SafeProviderMetadata(response.StatusCode, "", 0, maxRetryAfterSeconds)}, errors.New("telegram chat not found")
	}
	return SendOutcome{Class: ClassifyHTTPStatus(response.StatusCode), Retryable: IsRetryableClass(ClassifyHTTPStatus(response.StatusCode)), Metadata: SafeProviderMetadata(response.StatusCode, "", 0, maxRetryAfterSeconds)}, errors.New("telegram returned non-success status")
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
