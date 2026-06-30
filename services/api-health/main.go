package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"bolt-monitor/shared/api/response"
)

func handler(events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	env := response.Ok(map[string]string{"status": "ok"})
	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: mustMarshal(env),
	}, nil
}

func mustMarshal(env response.Envelope[map[string]string]) string {
	body, err := env.MarshalJSON()
	if err != nil {
		return `{"status":"error","reason":{"code":"INTERNAL","details":{}}}`
	}
	return string(body)
}

func main() {
	lambda.Start(handler)
}
