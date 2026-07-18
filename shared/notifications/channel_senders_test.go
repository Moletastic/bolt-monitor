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
