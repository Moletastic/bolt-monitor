package response

type Reason struct {
	Code    string         `json:"code"`
	Details map[string]any `json:"details"`
}
