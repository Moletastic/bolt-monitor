package response

import (
	"bytes"
	"encoding/json"
)

type Envelope[T any] struct {
	Status     Status      `json:"-"`
	Data       *T          `json:"-"`
	Reason     *Reason     `json:"-"`
	Message    *string     `json:"-"`
	Pagination *Pagination `json:"-"`
}

func (e Envelope[T]) MarshalJSON() ([]byte, error) {
	payload := make(map[string]any, 5)
	payload["status"] = e.Status
	if e.Data != nil {
		payload["data"] = *e.Data
	}
	if e.Reason != nil {
		payload["reason"] = e.Reason
	}
	if e.Message != nil {
		payload["message"] = *e.Message
	}
	if e.Pagination != nil {
		payload["pagination"] = e.Pagination
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	first := true
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	for key, value := range payload {
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		if !first {
			buf.WriteByte(',')
		}
		first = false
		buf.WriteByte('"')
		buf.WriteString(key)
		buf.WriteString(`":`)
		buf.Write(encoded)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func Ok[T any](data T, message ...string) Envelope[T] {
	env := Envelope[T]{Status: StatusSuccess, Data: &data}
	if len(message) > 0 && message[0] != "" {
		msg := message[0]
		env.Message = &msg
	}
	return env
}

func Err[T any](code string, details map[string]any) Envelope[T] {
	return Envelope[T]{Status: StatusError, Reason: &Reason{Code: code, Details: details}}
}

func OkPaginated[T any](data T, page, size, total int) Envelope[T] {
	return Envelope[T]{
		Status:     StatusSuccess,
		Data:       &data,
		Pagination: &Pagination{Page: page, Size: size, Total: total, Items: data},
	}
}
