package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"

	"github.com/Shurubtsov/lamoda-test-task/internal/domain/models"
)

const headerIP = "X-Forwarded-For"

type middleware struct {
	productCodes map[string]string
	mu           sync.RWMutex
}

func New() *middleware {
	pc := make(map[string]string)
	return &middleware{productCodes: pc}
}

func (m *middleware) SyncReservation(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get(headerIP)
		ipv4Addr := net.ParseIP(header).String()

		var data models.ReservationRequest
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("сервер не смог десереализовать тело запроса"))
			return
		}

		for _, product := range data.Products {
			m.mu.RLock()
			ip, ok := m.productCodes[product.Code]
			if !ok {
				m.mu.Lock()
				m.productCodes[product.Code] = ipv4Addr
				m.mu.Unlock()
				continue
			}
			m.mu.RUnlock()
			if ip != ipv4Addr {
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte("один или несколько товаров уже используются другой системой"))
				return
			}
		}
		next(w, r)
	}
}
