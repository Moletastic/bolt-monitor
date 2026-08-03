package main

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandleRequestReturnsGatewayOrFallbackRequestID(t *testing.T) {
	handler := newMonitorHandler(newFakeMonitorRepository())
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/unknown", RequestContext: events.APIGatewayV2HTTPRequestContext{RequestID: "gateway-request-1"}}
	response, err := handler.handleRequest(context.Background(), request)
	if err != nil || response.Headers["X-Request-Id"] != "gateway-request-1" {
		t.Fatalf("response = %+v, err = %v", response, err)
	}
	fallback, err := handler.handleRequest(context.Background(), events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/unknown"})
	if err != nil || fallback.Headers["X-Request-Id"] == "" {
		t.Fatalf("fallback = %+v, err = %v", fallback, err)
	}
}
