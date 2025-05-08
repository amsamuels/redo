package model

type CreateLinkRequest struct {
	UserID      string `json:"id"`
	Slug        string `json:"slug"`
	Destination string `json:"destination"`
}

type Link struct {
	LinkID      string `json:"id"`
	Slug        string `json:"slug"`
	Destination string `json:"destination"`
	CreatedAt   string `json:"created_at"`
}
