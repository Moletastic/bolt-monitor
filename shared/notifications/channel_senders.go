package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bolt-monitor/shared/outboundhttp"
)

const (
	defaultEmailAPIBaseURL     = "https://api.sendgrid.com"
	defaultSMSAPIBaseURL       = "https://api.twilio.com"
	defaultPagerDutyAPIBaseURL = "https://events.pagerduty.com"

	maxRetryAfterSeconds = 300
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
	executor HTTPExecutor
}

type SMSSender struct {
	executor HTTPExecutor
}

type WebhookSender struct {
	executor HTTPExecutor
}

type PagerDutySender struct {
	executor HTTPExecutor
}

func NewEmailSender(executors ...HTTPExecutor) *EmailSender {
	return &EmailSender{executor: senderExecutor(executors)}
}

func NewSMSSender(executors ...HTTPExecutor) *SMSSender {
	return &SMSSender{executor: senderExecutor(executors)}
}

func NewWebhookSender(executors ...HTTPExecutor) *WebhookSender {
	return &WebhookSender{executor: senderExecutor(executors)}
}

func NewPagerDutySender(executors ...HTTPExecutor) *PagerDutySender {
	return &PagerDutySender{executor: senderExecutor(executors)}
}

func (s *EmailSender) ChannelType() string     { return "email" }
func (s *SMSSender) ChannelType() string       { return "sms" }
func (s *WebhookSender) ChannelType() string   { return "webhook" }
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

func (s *EmailSender) Send(ctx context.Context, notification Notification) (SendOutcome, error) {
	var cfg EmailConfig
	if err := unmarshalConfig(notification.Config, &cfg); err != nil {
		return SendOutcome{Class: OutcomeInvalidConfig}, err
	}
	if strings.TrimSpace(cfg.APIKey) == "" || strings.TrimSpace(cfg.FromEmail) == "" || strings.TrimSpace(cfg.ToEmail) == "" {
		return SendOutcome{Class: OutcomeInvalidConfig, Retryable: false}, fmt.Errorf("email sender is missing required config")
	}
	payload := map[string]any{
		"personalizations": []map[string]any{{"to": []map[string]string{{"email": cfg.ToEmail}}}},
		"from":             map[string]string{"email": cfg.FromEmail},
		"subject":          buildSubject(cfg.SubjectPrefix, notification),
		"content":          []map[string]string{{"type": "text/plain", "value": notification.Message}},
	}
	out, err := sendJSONRequest(ctx, s.executor, http.MethodPost, joinEndpoint(firstNonEmpty(cfg.APIBaseURL, defaultEmailAPIBaseURL), "/v3/mail/send"), payload, map[string]string{"Authorization": "Bearer " + cfg.APIKey}, notification)
	return out, err
}

func (s *SMSSender) Send(ctx context.Context, notification Notification) (SendOutcome, error) {
	var cfg SMSTwilioConfig
	if err := unmarshalConfig(notification.Config, &cfg); err != nil {
		return SendOutcome{Class: OutcomeInvalidConfig}, err
	}
	if strings.TrimSpace(cfg.AccountSID) == "" || strings.TrimSpace(cfg.AuthToken) == "" || strings.TrimSpace(cfg.FromNumber) == "" || strings.TrimSpace(cfg.ToNumber) == "" {
		return SendOutcome{Class: OutcomeInvalidConfig, Retryable: false}, fmt.Errorf("sms sender is missing required config")
	}
	form := url.Values{"To": {cfg.ToNumber}, "From": {cfg.FromNumber}, "Body": {notification.Message}}.Encode()
	headers := make(http.Header)
	headers.Set("Content-Type", "application/x-www-form-urlencoded")
	headers.Set("Authorization", "Basic "+basicAuth(cfg.AccountSID, cfg.AuthToken))
	return sendRawRequest(ctx, s.executor, outboundhttp.Request{Method: http.MethodPost, URL: joinEndpoint(firstNonEmpty(cfg.APIBaseURL, defaultSMSAPIBaseURL), "/2010-04-01/Accounts/"+url.PathEscape(cfg.AccountSID)+"/Messages.json"), Header: headers, Body: strings.NewReader(form), Timeout: outboundhttp.NotificationTimeout}, notification)
}

func (s *WebhookSender) Send(ctx context.Context, notification Notification) (SendOutcome, error) {
	var cfg WebhookConfig
	if err := unmarshalConfig(notification.Config, &cfg); err != nil {
		return SendOutcome{Class: OutcomeInvalidConfig}, err
	}
	if strings.TrimSpace(cfg.URL) == "" {
		return SendOutcome{Class: OutcomeInvalidConfig, Retryable: false}, errMissingEndpointURL
	}
	body, err := json.Marshal(map[string]any{
		"eventType":  notification.EventType,
		"tenantId":   notification.TenantID,
		"serviceId":  notification.ServiceID,
		"monitorId":  notification.MonitorID,
		"incidentId": notification.IncidentID,
		"message":    notification.Message,
		"timestamp":  notification.Timestamp.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return SendOutcome{Class: OutcomeInvalidConfig, Retryable: false}, err
	}
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	for key, value := range cfg.Headers {
		reqHeaders.Set(key, value)
	}
	if notification.DeliveryID != "" {
		reqHeaders.Set("X-Bolt-Delivery-Id", notification.DeliveryID)
	}
	return sendRawRequest(ctx, s.executor, outboundhttp.Request{Method: http.MethodPost, URL: cfg.URL, Header: reqHeaders, Body: bytes.NewReader(body), Timeout: outboundhttp.NotificationTimeout}, notification)
}

func (s *PagerDutySender) Send(ctx context.Context, notification Notification) (SendOutcome, error) {
	var cfg PagerDutyConfig
	if err := unmarshalConfig(notification.Config, &cfg); err != nil {
		return SendOutcome{Class: OutcomeInvalidConfig}, err
	}
	if strings.TrimSpace(cfg.RoutingKey) == "" {
		return SendOutcome{Class: OutcomeInvalidConfig, Retryable: false}, fmt.Errorf("pagerduty routingKey is required")
	}
	action := "trigger"
	if notification.EventType == EventTypeIncidentUp {
		action = "resolve"
	}
	dedupKey := firstNonEmpty(cfg.DedupKey, notification.IncidentID, notification.DeliveryID)
	payload := map[string]any{
		"routing_key":  cfg.RoutingKey,
		"event_action": action,
		"dedup_key":    dedupKey,
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
	out, err := sendJSONRequest(ctx, s.executor, http.MethodPost, joinEndpoint(firstNonEmpty(cfg.APIBaseURL, defaultPagerDutyAPIBaseURL), "/v2/enqueue"), payload, nil, notification)
	return out, err
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

func sendJSONRequest(ctx context.Context, executor HTTPExecutor, method, target string, payload any, headers map[string]string, notification Notification) (SendOutcome, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return SendOutcome{Class: OutcomeInvalidConfig, Retryable: false}, err
	}
	reqHeaders := make(http.Header)
	reqHeaders.Set("Content-Type", "application/json")
	for key, value := range headers {
		reqHeaders.Set(key, value)
	}
	return sendRawRequest(ctx, executor, outboundhttp.Request{Method: method, URL: target, Header: reqHeaders, Body: bytes.NewReader(body), Timeout: outboundhttp.NotificationTimeout}, notification)
}

func sendRawRequest(ctx context.Context, executor HTTPExecutor, request outboundhttp.Request, notification Notification) (SendOutcome, error) {
	out, err := executor.Execute(ctx, request)
	if err != nil {
		outcome := classifyTransportError(err)
		return outcome, err
	}
	outcome := classifyResponse(out)
	outcome.ProviderName = notification.ChannelType
	if notification.DeliveryID != "" {
		if outcome.Metadata.ProviderRequestID == "" {
			outcome.Metadata.ProviderRequestID = notification.DeliveryID
		}
	}
	if !outcome.Retryable && outcome.Class != OutcomeAccepted {
		return outcome, fmt.Errorf("provider %s returned status %s", notification.ChannelType, outcome.Metadata.ProviderStatusClass)
	}
	return outcome, nil
}

func classifyTransportError(err error) SendOutcome {
	var httpErr *outboundhttp.Error
	if errors.As(err, &httpErr) {
		switch httpErr.Kind {
		case outboundhttp.KindTimeout:
			return SendOutcome{Class: OutcomeTimeout, Retryable: true, Metadata: ProviderMetadata{ProviderStatusClass: "timeout"}}
		default:
			return SendOutcome{Class: OutcomeTransport, Retryable: true, Metadata: ProviderMetadata{ProviderStatusClass: string(httpErr.Kind)}}
		}
	}
	return SendOutcome{Class: OutcomeTransport, Retryable: true, Metadata: ProviderMetadata{ProviderStatusClass: "transport"}}
}

func classifyResponse(response outboundhttp.Response) SendOutcome {
	class := ClassifyHTTPStatus(response.StatusCode)
	retryAfter := ParseRetryAfterSeconds(response.Header.Get("Retry-After"), maxRetryAfterSeconds)
	requestID := firstNonEmpty(response.Header.Get("X-Request-Id"), response.Header.Get("X-Request-ID"), response.Header.Get("Request-Id"))
	outcome := SendOutcome{Class: class, Retryable: IsRetryableClass(class), Metadata: SafeProviderMetadata(response.StatusCode, requestID, retryAfter, maxRetryAfterSeconds)}
	return outcome
}

func senderExecutor(executors []HTTPExecutor) HTTPExecutor {
	if len(executors) > 0 && executors[0] != nil {
		return executors[0]
	}
	return outboundhttp.NewExecutor()
}

func joinEndpoint(base, endpoint string) string {
	parsed, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return ""
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + endpoint
	return parsed.String()
}

func basicAuth(username, password string) string {
	return base64Encode(username + ":" + password)
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
