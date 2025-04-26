package model

type CreateLinkRequest struct {
	Slug        string `json:"slug"`
	Destination string `json:"destination"`
}

type Link struct {
	Slug        string `json:"slug"`
	Destination string `json:"destination"`
	CreatedAt   string `json:"created_at"`
}
