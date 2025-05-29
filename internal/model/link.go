package model

type CreateLinkRequest struct {
	Slug        string `json:"slug"`
	Destination string `json:"destination"`
}

type Link struct {
	LinkID      string `json:"id"`
	Slug        string `json:"slug"`
	ShortCode   string `json:"short_code"`
	ClickCount  int    `json:"clicks"`
	Is_active   bool   `json:"is_active"`
	Destination string `json:"destination"`
	CreatedAt   string `json:"created_at"`
}
