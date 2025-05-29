package model

import "time"

type ClicksByDay struct {
	DateLabel string `json:"date"`
	Clicks    int    `json:"clicks"`
}

type GroupedMetric struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type Click struct {
	ID          string    `json:"id"`
	LinkID      string    `json:"link_id"`
	IP          string    `json:"ip"`
	Referrer    string    `json:"referrer"`
	UserAgent   string    `json:"user_agent"`
	DeviceType  string    `json:"device_type"`
	Country     string    `json:"country"`
	Conversion  bool      `json:"conversion"`
	IsHighValue bool      `json:"is_high_value"`
	CreatedAt   time.Time `json:"created_at"`
}
