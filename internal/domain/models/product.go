package models

type ReservationRequest struct {
	Products []Product `json:"products"`
}
type Product struct {
	Code  string
	Name  string
	Size  int
	Count int
}
