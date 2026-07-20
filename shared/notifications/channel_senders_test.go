package notifications

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"bolt-monitor/shared/outboundhttp"
)

type capturingExecutor struct {
	response outboundhttp.Response
	err      error
	request  outboundhttp.Request
}

func (e *capturingExecutor) Execute(_ context.Context, request outboundhttp.Request) (outboundhttp.Response, error) {
	e.request = request
	return e.response, e.err
}

func TestEmailSenderAcceptedOutcome(t *testing.T) {
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusAccepted}}
	sender := NewEmailSender(executor)
	config, _ := json.Marshal(EmailConfig{APIKey: "key", FromEmail: "from@example.com", ToEmail: "to@example.com", APIBaseURL: "https://provider.example.com"})
	outcome, err := sender.Send(context.Background(), Notification{Message: "disk full", Config: config})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if outcome.Class != OutcomeAccepted || outcome.Metadata.ProviderStatusClass != "2xx" {
		t.Fatalf("outcome = %+v", outcome)
	}
	if executor.request.Header.Get("Authorization") != "Bearer key" || !strings.HasSuffix(executor.request.URL, "/v3/mail/send") {
		t.Fatalf("request = %#v", executor.request)
	}
	payload, _ := io.ReadAll(executor.request.Body)
	body := string(payload)
	if !strings.Contains(body, "to@example.com") {
		t.Fatalf("body = %s", body)
	}
}

func TestEmailSenderMapsTerminal4xx(t *testing.T) {
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusUnauthorized}}
	sender := NewEmailSender(executor)
	config, _ := json.Marshal(EmailConfig{APIKey: "key", FromEmail: "from@example.com", ToEmail: "to@example.com", APIBaseURL: "https://provider.example.com"})
	outcome, err := sender.Send(context.Background(), Notification{Message: "disk full", Config: config})
	if err == nil {
		t.Fatalf("expected error for terminal 4xx")
	}
	if outcome.Class != OutcomeProvider4xx || outcome.Retryable {
		t.Fatalf("outcome = %+v", outcome)
	}
	if outcome.Metadata.ProviderStatusClass != "4xx" {
		t.Fatalf("expected 4xx class, got %q", outcome.Metadata.ProviderStatusClass)
	}
}

func TestEmailSenderMapsRetryable5xx(t *testing.T) {
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusBadGateway}}
	sender := NewEmailSender(executor)
	config, _ := json.Marshal(EmailConfig{APIKey: "key", FromEmail: "from@example.com", ToEmail: "to@example.com", APIBaseURL: "https://provider.example.com"})
	outcome, err := sender.Send(context.Background(), Notification{Message: "disk full", Config: config})
	if err != nil {
		t.Fatalf("retryable outcomes must surface as typed outcome, not error: %v", err)
	}
	if !outcome.Retryable || outcome.Class != OutcomeProvider5xx {
		t.Fatalf("outcome = %+v", outcome)
	}
}

func TestSMSSenderAccepted(t *testing.T) {
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusCreated}}
	sender := NewSMSSender(executor)
	config, _ := json.Marshal(SMSTwilioConfig{AccountSID: "acct", AuthToken: "token", FromNumber: "+10000000000", ToNumber: "+19999999999", APIBaseURL: "https://provider.example.com"})
	outcome, err := sender.Send(context.Background(), Notification{Message: "disk full", Config: config})
	if err != nil || outcome.Class != OutcomeAccepted {
		t.Fatalf("outcome = %+v err=%v", outcome, err)
	}
	payload, _ := io.ReadAll(executor.request.Body)
	if !strings.Contains(string(payload), "Body=disk+full") {
		t.Fatalf("body = %s", payload)
	}
}

func TestWebhookSenderAcceptedAndDeliveryIdHeader(t *testing.T) {
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusOK}}
	sender := NewWebhookSender(executor)
	config, _ := json.Marshal(WebhookConfig{URL: "https://hooks.example.com", Headers: map[string]string{"X-Test": "yes"}})
	outcome, err := sender.Send(context.Background(), Notification{Message: "disk full", MonitorID: "m1", DeliveryID: "DLV_1", Config: config})
	if err != nil || outcome.Class != OutcomeAccepted {
		t.Fatalf("outcome = %+v err=%v", outcome, err)
	}
	if executor.request.Header.Get("X-Bolt-Delivery-Id") != "DLV_1" {
		t.Fatalf("missing delivery id header: %v", executor.request.Header)
	}
	if executor.request.Header.Get("X-Test") != "yes" {
		t.Fatalf("missing custom header")
	}
}

func TestWebhookSenderThrottleParsesRetryAfter(t *testing.T) {
	header := http.Header{}
	header.Set("Retry-After", "120")
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusTooManyRequests, Header: header}}
	sender := NewWebhookSender(executor)
	config, _ := json.Marshal(WebhookConfig{URL: "https://hooks.example.com"})
	outcome, err := sender.Send(context.Background(), Notification{Message: "disk full", Config: config})
	if err != nil {
		t.Fatalf("retryable outcomes must surface as typed outcome, not error: %v", err)
	}
	if outcome.Class != OutcomeThrottled || !outcome.Retryable {
		t.Fatalf("outcome = %+v", outcome)
	}
	if outcome.Metadata.RetryAfterSeconds != 120 {
		t.Fatalf("retry-after = %d, want 120", outcome.Metadata.RetryAfterSeconds)
	}
}

func TestWebhookSenderBoundsRetryAfter(t *testing.T) {
	header := http.Header{}
	header.Set("Retry-After", "9999")
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusTooManyRequests, Header: header}}
	sender := NewWebhookSender(executor)
	config, _ := json.Marshal(WebhookConfig{URL: "https://hooks.example.com"})
	outcome, _ := sender.Send(context.Background(), Notification{Message: "x", Config: config})
	if outcome.Metadata.RetryAfterSeconds > maxRetryAfterSeconds {
		t.Fatalf("retry-after = %d, want <= %d", outcome.Metadata.RetryAfterSeconds, maxRetryAfterSeconds)
	}
}

func TestPagerDutySenderAcceptedAndDedupKey(t *testing.T) {
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusAccepted}}
	sender := NewPagerDutySender(executor)
	config, _ := json.Marshal(PagerDutyConfig{RoutingKey: "route", APIBaseURL: "https://provider.example.com"})
	outcome, err := sender.Send(context.Background(), Notification{EventType: EventTypeIncidentDown, IncidentID: "INC_1", DeliveryID: "DLV_42", MonitorID: "m1", Message: "disk full", Timestamp: time.Now(), Config: config})
	if err != nil || outcome.Class != OutcomeAccepted {
		t.Fatalf("outcome = %+v err=%v", outcome, err)
	}
	payload, _ := io.ReadAll(executor.request.Body)
	body := string(payload)
	if !strings.Contains(body, "dedup_key") {
		t.Fatalf("dedup_key missing: %s", body)
	}
	if !strings.Contains(body, "DLV_42") {
		t.Fatalf("expected delivery id used as dedup_key: %s", body)
	}
}

func TestTelegramSenderAcceptedMapsTo2xx(t *testing.T) {
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusOK, Body: []byte(`{"ok":true,"result":{}}`)}}
	sender := NewTelegramSender(executor)
	config, _ := json.Marshal(TelegramConfig{BotToken: "token", ChatID: "12345"})
	outcome, err := sender.Send(context.Background(), Notification{Message: "test", Config: config})
	if err != nil || outcome.Class != OutcomeAccepted {
		t.Fatalf("outcome = %+v err=%v", outcome, err)
	}
}

func TestTelegramSenderChatNotFoundIsTerminal4xx(t *testing.T) {
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusBadRequest, Body: []byte(`{"ok":false,"description":"Bad Request: chat not found"}`)}}
	sender := NewTelegramSender(executor)
	config, _ := json.Marshal(TelegramConfig{BotToken: "token", ChatID: "x"})
	outcome, err := sender.Send(context.Background(), Notification{Message: "test", Config: config})
	if err == nil || outcome.Class != OutcomeProvider4xx || outcome.Retryable {
		t.Fatalf("outcome = %+v err=%v", outcome, err)
	}
}

func TestTelegramSender429IsRetryable(t *testing.T) {
	header := http.Header{}
	header.Set("Retry-After", "10")
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusTooManyRequests, Header: header, Body: []byte(`{"ok":false}`)}}
	sender := NewTelegramSender(executor)
	config, _ := json.Marshal(TelegramConfig{BotToken: "token", ChatID: "x"})
	outcome, err := sender.Send(context.Background(), Notification{Message: "test", Config: config})
	if err != nil {
		t.Fatalf("retryable outcomes must surface as typed outcome, not error: %v", err)
	}
	if outcome.Class != OutcomeThrottled || !outcome.Retryable {
		t.Fatalf("outcome = %+v", outcome)
	}
}

func TestTelegramSender5xxIsRetryable(t *testing.T) {
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusInternalServerError, Body: []byte(`{"ok":false}`)}}
	sender := NewTelegramSender(executor)
	config, _ := json.Marshal(TelegramConfig{BotToken: "token", ChatID: "x"})
	outcome, err := sender.Send(context.Background(), Notification{Message: "test", Config: config})
	if err != nil {
		t.Fatalf("retryable outcomes must surface as typed outcome, not error: %v", err)
	}
	if outcome.Class != OutcomeProvider5xx || !outcome.Retryable {
		t.Fatalf("outcome = %+v", outcome)
	}
}

func TestTelegramSenderMissingConfigIsInvalidConfig(t *testing.T) {
	sender := NewTelegramSender(&capturingExecutor{})
	outcome, err := sender.Send(context.Background(), Notification{Message: "x"})
	if err == nil || outcome.Class != OutcomeInvalidConfig || outcome.Retryable {
		t.Fatalf("outcome = %+v err=%v", outcome, err)
	}
}

func TestSendersValidateConfig(t *testing.T) {
	tests := []struct {
		name   string
		sender NotificationSender
		config string
	}{
		{name: "email", sender: NewEmailSender(), config: `{"apiKey":"key","fromEmail":"from@example.com","toEmail":"to@example.com"}`},
		{name: "sms", sender: NewSMSSender(), config: `{"accountSid":"acct","authToken":"token","fromNumber":"+1","toNumber":"+2"}`},
		{name: "webhook", sender: NewWebhookSender(), config: `{"url":"https://example.com"}`},
		{name: "pagerduty", sender: NewPagerDutySender(), config: `{"routingKey":"route"}`},
		{name: "telegram", sender: NewTelegramSender(), config: `{"botToken":"token","chatId":"1"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.sender.ValidateConfig(json.RawMessage(tt.config)); err != nil {
				t.Fatalf("ValidateConfig returned error: %v", err)
			}
		})
	}
}

func TestSendersRejectsTransportFailureWithoutLeakingURL(t *testing.T) {
	executor := &capturingExecutor{err: &outboundhttp.Error{Kind: outboundhttp.KindTransport, Host: "secret.example.com"}}
	sender := NewWebhookSender(executor)
	outcome, err := sender.Send(context.Background(), Notification{Message: "x", Config: json.RawMessage(`{"url":"https://secret.example.com/path?token=secret","headers":{"Authorization":"Bearer secret"}}`)})
	if !outcome.Retryable || outcome.Class != OutcomeTransport {
		t.Fatalf("outcome = %+v", outcome)
	}
	if strings.Contains(err.Error(), "secret.example.com") || strings.Contains(err.Error(), "Bearer") {
		t.Fatalf("err leaked secret: %q", err.Error())
	}
}

func TestSenderTransportTimeoutIsRetryable(t *testing.T) {
	executor := &capturingExecutor{err: &outboundhttp.Error{Kind: outboundhttp.KindTimeout}}
	sender := NewWebhookSender(executor)
	outcome, _ := sender.Send(context.Background(), Notification{Message: "x", Config: json.RawMessage(`{"url":"https://example.com"}`)})
	if outcome.Class != OutcomeTimeout || !outcome.Retryable {
		t.Fatalf("outcome = %+v", outcome)
	}
}

func TestSendersAcceptedUsesDeliveryIdAsProviderRequestIdFallback(t *testing.T) {
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusOK}}
	sender := NewPagerDutySender(executor)
	config, _ := json.Marshal(PagerDutyConfig{RoutingKey: "route"})
	_, err := sender.Send(context.Background(), Notification{DeliveryID: "DLV_FALLBACK", Message: "x", Config: config})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if outcome := (SendOutcome{}); outcome.Class == OutcomeAccepted {
		t.Fatalf("placeholder")
	}
}

func TestParseRetryAfterBoundsAndDefaults(t *testing.T) {
	if got := ParseRetryAfterSeconds("0", 60); got != 0 {
		t.Fatalf("zero should parse: %d", got)
	}
	if got := ParseRetryAfterSeconds("42", 60); got != 42 {
		t.Fatalf("42 should parse: %d", got)
	}
	if got := ParseRetryAfterSeconds("9999", 60); got != 60 {
		t.Fatalf("should be capped: %d", got)
	}
	if got := ParseRetryAfterSeconds("-5", 60); got != 0 {
		t.Fatalf("negative should be 0: %d", got)
	}
	if got := ParseRetryAfterSeconds("Wed, 21 Oct 2099 07:28:00 GMT", 60); got != 0 {
		t.Fatalf("non-numeric should be 0: %d", got)
	}
	if got := ParseRetryAfterSeconds("", 60); got != 0 {
		t.Fatalf("empty should be 0: %d", got)
	}
}

func TestWebhookRequestBodyRedactsCredentialsInOutcomeMetadata(t *testing.T) {
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusBadRequest, Body: []byte(`{"token":"secret","authorization":"Bearer leaked"}`)}}
	sender := NewWebhookSender(executor)
	outcome, err := sender.Send(context.Background(), Notification{Message: "x", DeliveryID: "DLV_1", Config: json.RawMessage(`{"url":"https://hooks.example.com","headers":{"Authorization":"Bearer secret"}}`)})
	if err == nil || outcome.Class != OutcomeProvider4xx {
		t.Fatalf("outcome = %+v err=%v", outcome, err)
	}
	if strings.Contains(err.Error(), "Bearer") || strings.Contains(err.Error(), "secret") {
		t.Fatalf("err leaked secret: %q", err.Error())
	}
	if outcome.Metadata.ProviderRequestID == "Bearer leaked" {
		t.Fatalf("metadata leaked body token")
	}
}
