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

func TestEmailSenderSend(t *testing.T) {
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusAccepted}}
	sender := NewEmailSender(executor)
	config, _ := json.Marshal(EmailConfig{APIKey: "key", FromEmail: "from@example.com", ToEmail: "to@example.com", APIBaseURL: "https://provider.example.com"})
	err := sender.Send(context.Background(), Notification{Message: "disk full", Config: config})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
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

func TestSMSSenderSend(t *testing.T) {
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusCreated}}
	sender := NewSMSSender(executor)
	config, _ := json.Marshal(SMSTwilioConfig{AccountSID: "acct", AuthToken: "token", FromNumber: "+10000000000", ToNumber: "+19999999999", APIBaseURL: "https://provider.example.com"})
	err := sender.Send(context.Background(), Notification{Message: "disk full", Config: config})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if executor.request.Header.Get("Authorization") == "" {
		t.Fatal("missing Authorization header")
	}
	payload, _ := io.ReadAll(executor.request.Body)
	body := string(payload)
	if !strings.Contains(body, "Body=disk+full") {
		t.Fatalf("body = %s", body)
	}
}

func TestWebhookSenderSend(t *testing.T) {
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusOK}}
	sender := NewWebhookSender(executor)
	config, _ := json.Marshal(WebhookConfig{URL: "https://hooks.example.com", Headers: map[string]string{"X-Test": "yes"}})
	err := sender.Send(context.Background(), Notification{Message: "disk full", MonitorID: "m1", Config: config})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if executor.request.Header.Get("X-Test") != "yes" {
		t.Fatalf("header = %q", executor.request.Header.Get("X-Test"))
	}
	payload, _ := io.ReadAll(executor.request.Body)
	gotBody := string(payload)
	if !strings.Contains(gotBody, "disk full") {
		t.Fatalf("body = %s", gotBody)
	}
}

func TestPagerDutySenderSend(t *testing.T) {
	executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusAccepted}}
	sender := NewPagerDutySender(executor)
	config, _ := json.Marshal(PagerDutyConfig{RoutingKey: "route", APIBaseURL: "https://provider.example.com"})
	err := sender.Send(context.Background(), Notification{EventType: EventTypeIncidentDown, IncidentID: "INC_1", MonitorID: "m1", Message: "disk full", Timestamp: time.Now(), Config: config})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	payload, _ := io.ReadAll(executor.request.Body)
	gotBody := string(payload)
	if !strings.Contains(gotBody, "trigger") {
		t.Fatalf("body = %s", gotBody)
	}
	if !strings.Contains(gotBody, "route") {
		t.Fatalf("body = %s", gotBody)
	}
}

func TestTelegramSenderChatNotFoundErrorIsActionable(t *testing.T) {
	sender := NewTelegramSender(&capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusBadRequest, Body: []byte(`{"ok":false,"error_code":400,"description":"Bad Request: chat not found"}`)}})
	config, _ := json.Marshal(TelegramConfig{BotToken: "token", ChatID: "valid-looking-chat"})

	err := sender.Send(context.Background(), Notification{Message: "test", Config: config})

	if err == nil {
		t.Fatal("Send returned nil error")
	}
	if !strings.Contains(err.Error(), "use the numeric chat ID") {
		t.Fatalf("error = %q, want numeric chat ID guidance", err.Error())
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.sender.ValidateConfig(json.RawMessage(tt.config)); err != nil {
				t.Fatalf("ValidateConfig returned error: %v", err)
			}
		})
	}
}

func TestSenderRegistryUsesInjectedExecutorForEveryChannelType(t *testing.T) {
	tests := []struct {
		name        string
		channelType string
		config      string
	}{
		{name: "telegram", channelType: "telegram", config: `{"botToken":"token","chatId":"123"}`},
		{name: "email", channelType: "email", config: `{"apiKey":"key","fromEmail":"from@example.com","toEmail":"to@example.com","apiBaseUrl":"https://provider.example.com"}`},
		{name: "sms", channelType: "sms", config: `{"accountSid":"account","authToken":"token","fromNumber":"+10000000000","toNumber":"+19999999999","apiBaseUrl":"https://provider.example.com"}`},
		{name: "webhook", channelType: "webhook", config: `{"url":"https://hooks.example.com","headers":{"Authorization":"Bearer secret"}}`},
		{name: "pagerduty", channelType: "pagerduty", config: `{"routingKey":"routing-key","apiBaseUrl":"https://provider.example.com"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &capturingExecutor{response: outboundhttp.Response{StatusCode: http.StatusOK, Body: []byte(`{"ok":true}`)}}
			registry := NewSenderRegistry(executor)
			sender, ok := registry.Get(tt.channelType)
			if !ok {
				t.Fatalf("sender %q was not registered", tt.channelType)
			}
			if err := sender.Send(context.Background(), Notification{Message: "same-origin delivery", Config: json.RawMessage(tt.config)}); err != nil {
				t.Fatalf("Send returned error: %v", err)
			}
			if executor.request.Timeout != outboundhttp.NotificationTimeout || executor.request.Body == nil {
				t.Fatalf("request = %#v", executor.request)
			}
			if tt.channelType == "webhook" && executor.request.Header.Get("Authorization") != "Bearer secret" {
				t.Fatalf("same-origin webhook header = %q", executor.request.Header.Get("Authorization"))
			}
		})
	}
}

func TestSenderPolicyFailuresAreSanitized(t *testing.T) {
	tests := []struct {
		name string
		kind outboundhttp.Kind
	}{
		{name: "cross-origin credential redirect", kind: outboundhttp.KindRedirectBlocked},
		{name: "oversized response", kind: outboundhttp.KindResponseTooLarge},
		{name: "timeout", kind: outboundhttp.KindTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &capturingExecutor{err: &outboundhttp.Error{Kind: tt.kind, Host: "secret.example.com"}}
			sender := NewWebhookSender(executor)
			err := sender.Send(context.Background(), Notification{Message: "body-secret", Config: json.RawMessage(`{"url":"https://secret.example.com/path?token=secret","headers":{"Authorization":"Bearer secret"}}`)})
			if !outboundhttp.IsKind(err, tt.kind) {
				t.Fatalf("error = %v, want outbound kind %q", err, tt.kind)
			}
			if strings.Contains(err.Error(), "secret.example.com") || strings.Contains(err.Error(), "token=secret") || strings.Contains(err.Error(), "Bearer secret") || strings.Contains(err.Error(), "body-secret") {
				t.Fatalf("sender error leaked secret material: %q", err)
			}
		})
	}
}
