package models

type Storage struct {
	ID      uint   `json:"id"`
	Name    string `json:"name,omitempty"`
	Aviable bool   `json:"aviable,omitempty"`
}
