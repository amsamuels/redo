package model

type CreateLinkRequest struct {
	Slug        string `json:"slug"`
	Destination string `json:"destination"`
}
