package response

type Pagination struct {
	Page       int    `json:"page,omitempty"`
	Size       int    `json:"size"`
	Total      int    `json:"total,omitempty"`
	Items      any    `json:"items,omitempty"`
	NextCursor string `json:"nextCursor,omitempty"`
}
