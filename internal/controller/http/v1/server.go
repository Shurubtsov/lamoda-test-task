package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/Shurubtsov/lamoda-test-task/internal/domain/models"
)

type ReservationUsecase interface {
	ProductReservation(products []models.Product) error
}

type server struct {
	mu           sync.Mutex
	productInUse map[string]string
}

func NewServer(pie map[string]string) *server {
	return &server{productInUse: pie}
}

func (s *server) ReservationHandler(w http.ResponseWriter, r *http.Request) {
	var products []models.Product
	if err := json.NewDecoder(r.Body).Decode(&products); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		msg := fmt.Sprint("сервер не смог десереализовать тело запроса: ", err.Error())
		w.Write([]byte(msg))
		return
	}
	defer func() {
		for _, product := range products {
			s.mu.Lock()
			delete(s.productInUse, product.Code)
			s.mu.Unlock()
		}
	}()
}
