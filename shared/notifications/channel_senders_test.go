package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestEmailSenderSend(t *testing.T) {
	var authHeader string
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/mail/send" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		authHeader = r.Header.Get("Authorization")
		payload, _ := io.ReadAll(r.Body)
		body = string(payload)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	sender := NewEmailSender()
	config, _ := json.Marshal(EmailConfig{APIKey: "key", FromEmail: "from@example.com", ToEmail: "to@example.com", APIBaseURL: server.URL})
	err := sender.Send(context.Background(), Notification{Message: "disk full", Config: config})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if authHeader != "Bearer key" {
		t.Fatalf("auth header = %q", authHeader)
	}
	if !strings.Contains(body, "to@example.com") {
		t.Fatalf("body = %s", body)
	}
}

func TestSMSSenderSend(t *testing.T) {
	var authHeader string
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		payload, _ := io.ReadAll(r.Body)
		body = string(payload)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	sender := NewSMSSender()
	config, _ := json.Marshal(SMSTwilioConfig{AccountSID: "acct", AuthToken: "token", FromNumber: "+10000000000", ToNumber: "+19999999999", APIBaseURL: server.URL})
	err := sender.Send(context.Background(), Notification{Message: "disk full", Config: config})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if authHeader == "" {
		t.Fatal("missing Authorization header")
	}
	if !strings.Contains(body, "Body=disk full") {
		t.Fatalf("body = %s", body)
	}
}

func TestWebhookSenderSend(t *testing.T) {
	var gotHeader string
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Test")
		payload, _ := io.ReadAll(r.Body)
		gotBody = string(payload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewWebhookSender()
	config, _ := json.Marshal(WebhookConfig{URL: server.URL, Headers: map[string]string{"X-Test": "yes"}})
	err := sender.Send(context.Background(), Notification{Message: "disk full", MonitorID: "m1", Config: config})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if gotHeader != "yes" {
		t.Fatalf("header = %q", gotHeader)
	}
	if !strings.Contains(gotBody, "disk full") {
		t.Fatalf("body = %s", gotBody)
	}
}

func TestPagerDutySenderSend(t *testing.T) {
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := io.ReadAll(r.Body)
		gotBody = string(payload)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	sender := NewPagerDutySender()
	config, _ := json.Marshal(PagerDutyConfig{RoutingKey: "route", APIBaseURL: server.URL})
	err := sender.Send(context.Background(), Notification{EventType: EventTypeIncidentDown, IncidentID: "INC_1", MonitorID: "m1", Message: "disk full", Timestamp: time.Now(), Config: config})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if !strings.Contains(gotBody, "trigger") {
		t.Fatalf("body = %s", gotBody)
	}
	if !strings.Contains(gotBody, "route") {
		t.Fatalf("body = %s", gotBody)
	}
}

func TestTelegramSenderChatNotFoundErrorIsActionable(t *testing.T) {
	sender := NewTelegramSender()
	sender.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(`{"ok":false,"error_code":400,"description":"Bad Request: chat not found"}`)),
		}, nil
	})}
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
