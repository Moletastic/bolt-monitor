package errors

import (
	"net/http"

	"bolt-monitor/shared/api/response"
)

func Respond(err error) (int, response.Envelope[any]) {
	te, ok := As(err)
	if ok {
		return StatusOf(te.Code), response.Err[any](string(te.Code), te.Details)
	}
	return http.StatusInternalServerError, response.Err[any](string(CodeInternal), nil)
}
