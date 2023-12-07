package models

import "errors"

var (
	ErrNilStorageID = errors.New("storage id can't be nil")
)

type Storage struct {
	ID      *uint  `json:"id"`
	Name    string `json:"name,omitempty"`
	Aviable bool   `json:"aviable,omitempty"`
}

func (s Storage) Validate() error {
	if s.ID == nil {
		return ErrNilStorageID
	}
	return nil
}
