package main

import "bolt-monitor/shared/api/response"

func errEnvelopeForTest() response.Envelope[any] {
	return response.Err[any]("HEALTH_UNAVAILABLE", nil)
}
