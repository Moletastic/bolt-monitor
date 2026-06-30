package response

import "encoding/json"

type Status string

const (
	StatusSuccess Status = "success"
	StatusError   Status = "error"
)

func (s Status) String() string {
	return string(s)
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}