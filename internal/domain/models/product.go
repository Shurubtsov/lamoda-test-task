package models

type Product struct {
	Code  string `json:"code"`
	Name  string `json:"name,omitempty"`
	Size  uint   `json:"size,omitempty"`
	Count uint   `json:"count,omitempty"`
}
