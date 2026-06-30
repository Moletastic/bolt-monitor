package response

import (
	"encoding/json"
	"strings"
	"testing"
)

type sample struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestOkEnvelopeOmitsReasonAndMessage(t *testing.T) {
	body, err := Ok(sample{Name: "ok", Value: 1}).MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if parsed["status"] != "success" {
		t.Fatalf("status = %v", parsed["status"])
	}
	if _, ok := parsed["reason"]; ok {
		t.Fatalf("reason key present: %s", body)
	}
	if _, ok := parsed["message"]; ok {
		t.Fatalf("message key present: %s", body)
	}
	if _, ok := parsed["pagination"]; ok {
		t.Fatalf("pagination key present: %s", body)
	}
	data, ok := parsed["data"].(map[string]any)
	if !ok {
		t.Fatalf("data not object: %s", body)
	}
	if data["name"] != "ok" || data["value"].(float64) != 1 {
		t.Fatalf("data payload = %+v", data)
	}
}

func TestErrEnvelopeOmitsDataAndPagination(t *testing.T) {
	body, err := Err[sample]("NOT_FOUND", map[string]any{"id": "abc"}).MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if parsed["status"] != "error" {
		t.Fatalf("status = %v", parsed["status"])
	}
	if _, ok := parsed["data"]; ok {
		t.Fatalf("data key present: %s", body)
	}
	if _, ok := parsed["pagination"]; ok {
		t.Fatalf("pagination key present: %s", body)
	}
	reason, ok := parsed["reason"].(map[string]any)
	if !ok {
		t.Fatalf("reason missing: %s", body)
	}
	if reason["code"] != "NOT_FOUND" {
		t.Fatalf("reason.code = %v", reason["code"])
	}
}

func TestOkEnvelopeWithMessageIncludesMessage(t *testing.T) {
	body, err := Ok(sample{Name: "ok"}, "all good").MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if !strings.Contains(string(body), `"message":"all good"`) {
		t.Fatalf("message missing: %s", body)
	}
}

func TestOkPaginatedEnvelopeIncludesPagination(t *testing.T) {
	body, err := OkPaginated([]sample{{Name: "a"}, {Name: "b"}}, 1, 2, 2).MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	pagination, ok := parsed["pagination"].(map[string]any)
	if !ok {
		t.Fatalf("pagination missing: %s", body)
	}
	if pagination["page"].(float64) != 1 || pagination["size"].(float64) != 2 || pagination["total"].(float64) != 2 {
		t.Fatalf("pagination = %+v", pagination)
	}
	items, ok := pagination["items"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("items = %+v", pagination["items"])
	}
}

func TestStatusStringSerialization(t *testing.T) {
	if StatusSuccess.String() != "success" {
		t.Fatalf("StatusSuccess = %q", StatusSuccess)
	}
	if StatusError.String() != "error" {
		t.Fatalf("StatusError = %q", StatusError)
	}
	successBytes, err := json.Marshal(StatusSuccess)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if string(successBytes) != `"success"` {
		t.Fatalf("StatusSuccess json = %s", successBytes)
	}
}

func TestReasonShape(t *testing.T) {
	body, err := json.Marshal(Reason{Code: "BOOM", Details: map[string]any{"x": 1}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(body), `"code":"BOOM"`) {
		t.Fatalf("code missing: %s", body)
	}
}

func TestPaginationShape(t *testing.T) {
	body, err := json.Marshal(Pagination{Page: 2, Size: 10, Total: 30, Items: []int{1, 2}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(body), `"page":2`) {
		t.Fatalf("page missing: %s", body)
	}
}