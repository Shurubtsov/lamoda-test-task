package models

import (
	"errors"
	"regexp"
)

var ErrCodeNotValid = errors.New("code of product not valid")

type Product struct {
	Code  string `json:"code"`
	Name  string `json:"name,omitempty"`
	ID    uint   `json:"id,omitempty"`
	Size  uint   `json:"size,omitempty"`
	Count uint   `json:"count,omitempty"`
}

func (p Product) Validate() error {
	r, err := regexp.Compile("^[A-Z]{2}-[A-Z0-9]+$")
	if err != nil {
		return err
	}
	if !r.MatchString(p.Code) {
		return ErrCodeNotValid
	}
	return nil
}
