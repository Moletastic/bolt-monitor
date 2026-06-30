package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultEmailAPIBaseURL     = "https://api.sendgrid.com"
	defaultSMSAPIBaseURL       = "https://api.twilio.com"
	defaultPagerDutyAPIBaseURL = "https://events.pagerduty.com"
)

var errMissingEndpointURL = errors.New("endpoint URL is required")

type EmailConfig struct {
	APIKey        string `json:"apiKey"`
	FromEmail     string `json:"fromEmail"`
	ToEmail       string `json:"toEmail"`
	SubjectPrefix string `json:"subjectPrefix,omitempty"`
	APIBaseURL    string `json:"apiBaseUrl,omitempty"`
}

type SMSTwilioConfig struct {
	AccountSID string `json:"accountSid"`
	AuthToken  string `json:"authToken"`
	FromNumber string `json:"fromNumber"`
	ToNumber   string `json:"toNumber"`
	APIBaseURL string `json:"apiBaseUrl,omitempty"`
}

type WebhookConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

type PagerDutyConfig struct {
	RoutingKey string `json:"routingKey"`
	DedupKey   string `json:"dedupKey,omitempty"`
	Severity   string `json:"severity,omitempty"`
	Source     string `json:"source,omitempty"`
	Component  string `json:"component,omitempty"`
	Group      string `json:"group,omitempty"`
	Class      string `json:"class,omitempty"`
	APIBaseURL string `json:"apiBaseUrl,omitempty"`
}

type EmailSender struct {
	httpClient *http.Client
}

type SMSSender struct {
	httpClient *http.Client
}

type WebhookSender struct {
	httpClient *http.Client
}

type PagerDutySender struct {
	httpClient *http.Client
}

func NewEmailSender() *EmailSender {
	return &EmailSender{httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func NewSMSSender() *SMSSender {
	return &SMSSender{httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func NewWebhookSender() *WebhookSender {
	return &WebhookSender{httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func NewPagerDutySender() *PagerDutySender {
	return &PagerDutySender{httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func (s *EmailSender) ChannelType() string { return "email" }

func (s *SMSSender) ChannelType() string { return "sms" }

func (s *WebhookSender) ChannelType() string { return "webhook" }

func (s *PagerDutySender) ChannelType() string { return "pagerduty" }

func (s *EmailSender) ValidateConfig(config json.RawMessage) error {
	var cfg EmailConfig
	if err := unmarshalConfig(config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return errors.New("email api key is required")
	}
	if strings.TrimSpace(cfg.FromEmail) == "" {
		return errors.New("fromEmail is required")
	}
	if strings.TrimSpace(cfg.ToEmail) == "" {
		return errors.New("toEmail is required")
	}
	return nil
}

func (s *SMSSender) ValidateConfig(config json.RawMessage) error {
	var cfg SMSTwilioConfig
	if err := unmarshalConfig(config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.AccountSID) == "" {
		return errors.New("accountSid is required")
	}
	if strings.TrimSpace(cfg.AuthToken) == "" {
		return errors.New("authToken is required")
	}
	if strings.TrimSpace(cfg.FromNumber) == "" {
		return errors.New("fromNumber is required")
	}
	if strings.TrimSpace(cfg.ToNumber) == "" {
		return errors.New("toNumber is required")
	}
	return nil
}

func (s *WebhookSender) ValidateConfig(config json.RawMessage) error {
	var cfg WebhookConfig
	if err := unmarshalConfig(config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.URL) == "" {
		return errMissingEndpointURL
	}
	return nil
}

func (s *PagerDutySender) ValidateConfig(config json.RawMessage) error {
	var cfg PagerDutyConfig
	if err := unmarshalConfig(config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.RoutingKey) == "" {
		return errors.New("routingKey is required")
	}
	return nil
}

func (s *EmailSender) Send(ctx context.Context, notification Notification) error {
	var cfg EmailConfig
	if err := unmarshalConfig(notification.Config, &cfg); err != nil {
		return fmt.Errorf("invalid email config: %w", err)
	}
	payload := map[string]any{
		"personalizations": []map[string]any{{"to": []map[string]string{{"email": cfg.ToEmail}}}},
		"from":             map[string]string{"email": cfg.FromEmail},
		"subject":          buildSubject(cfg.SubjectPrefix, notification),
		"content":          []map[string]string{{"type": "text/plain", "value": notification.Message}},
	}
	_, err := sendJSONRequest(ctx, s.httpClient, http.MethodPost, strings.TrimRight(firstNonEmpty(cfg.APIBaseURL, defaultEmailAPIBaseURL), "/")+"/v3/mail/send", payload, map[string]string{"Authorization": "Bearer " + cfg.APIKey})
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

func (s *SMSSender) Send(ctx context.Context, notification Notification) error {
	var cfg SMSTwilioConfig
	if err := unmarshalConfig(notification.Config, &cfg); err != nil {
		return fmt.Errorf("invalid sms config: %w", err)
	}
	form := "To=" + cfg.ToNumber + "&From=" + cfg.FromNumber + "&Body=" + notification.Message
	url := strings.TrimRight(firstNonEmpty(cfg.APIBaseURL, defaultSMSAPIBaseURL), "/") + "/2010-04-01/Accounts/" + cfg.AccountSID + "/Messages.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(form))
	if err != nil {
		return err
	}
	req.SetBasicAuth(cfg.AccountSID, cfg.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send sms: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("send sms: status=%d body=%s", resp.StatusCode, string(body))
	}
	return nil
}

func (s *WebhookSender) Send(ctx context.Context, notification Notification) error {
	var cfg WebhookConfig
	if err := unmarshalConfig(notification.Config, &cfg); err != nil {
		return fmt.Errorf("invalid webhook config: %w", err)
	}
	_, err := sendJSONRequest(ctx, s.httpClient, http.MethodPost, cfg.URL, notification, cfg.Headers)
	if err != nil {
		return fmt.Errorf("send webhook: %w", err)
	}
	return nil
}

func (s *PagerDutySender) Send(ctx context.Context, notification Notification) error {
	var cfg PagerDutyConfig
	if err := unmarshalConfig(notification.Config, &cfg); err != nil {
		return fmt.Errorf("invalid pagerduty config: %w", err)
	}
	action := "trigger"
	if notification.EventType == EventTypeIncidentUp {
		action = "resolve"
	}
	payload := map[string]any{
		"routing_key":  cfg.RoutingKey,
		"event_action": action,
		"dedup_key":    firstNonEmpty(cfg.DedupKey, notification.IncidentID),
		"payload": map[string]any{
			"summary":        notification.Message,
			"source":         firstNonEmpty(cfg.Source, notification.MonitorID),
			"severity":       firstNonEmpty(cfg.Severity, "critical"),
			"component":      cfg.Component,
			"group":          cfg.Group,
			"class":          cfg.Class,
			"custom_details": notification,
		},
	}
	_, err := sendJSONRequest(ctx, s.httpClient, http.MethodPost, strings.TrimRight(firstNonEmpty(cfg.APIBaseURL, defaultPagerDutyAPIBaseURL), "/")+"/v2/enqueue", payload, nil)
	if err != nil {
		return fmt.Errorf("send pagerduty: %w", err)
	}
	return nil
}

func unmarshalConfig[T any](config json.RawMessage, out *T) error {
	if len(config) == 0 {
		return errors.New("config is required")
	}
	if err := json.Unmarshal(config, out); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	return nil
}

func sendJSONRequest(ctx context.Context, client *http.Client, method, url string, payload any, headers map[string]string) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status=%d body=%s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

func buildSubject(prefix string, notification Notification) string {
	base := strings.TrimSpace(notification.Message)
	if base == "" {
		base = "Bolt Monitor notification"
	}
	if strings.TrimSpace(prefix) == "" {
		return base
	}
	return strings.TrimSpace(prefix) + " " + base
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
