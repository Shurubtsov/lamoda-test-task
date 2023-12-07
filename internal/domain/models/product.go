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

type ProductsResponse struct {
	Status           string    `json:"status"`
	Message          string    `json:"message"`
	Error            string    `json:"error,omitempty"`
	ReservedProducts []Product `json:"reserved_products,omitempty"`
	ExemptedProducts []Product `json:"exempted_products,omitempty"`
	NotValid         []Product `json:"not_valid,omitempty"`
}

func NewProductsResponse(notValid []Product, reserved []Product, exempted []Product, status, msg string) ProductsResponse {
	return ProductsResponse{
		Status:           status,
		Message:          msg,
		ReservedProducts: reserved,
		ExemptedProducts: exempted,
		NotValid:         notValid,
	}
}

func (p Product) Validate() error {
	r, err := regexp.Compile("^[A-Z0-9]{3}-[A-Z0-9]{3}-[A-Z0-9]{3}-[A-Z0-9]{3}$")
	if err != nil {
		return err
	}
	if !r.MatchString(p.Code) {
		return ErrCodeNotValid
	}
	return nil
}
