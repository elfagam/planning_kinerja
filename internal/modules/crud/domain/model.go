package domain

import "time"

type Record struct {
	ID          int64          `json:"id"`
	Code        string         `json:"code"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Attributes  map[string]any `json:"attributes"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type Payload struct {
	Code        string         `json:"code"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Attributes  map[string]any `json:"attributes"`
}

type ListFilter struct {
	Query  string
	Page   int
	Limit  int
	Offset int
}
