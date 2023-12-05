package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/Shurubtsov/lamoda-test-task/internal/domain/models"
	"github.com/Shurubtsov/lamoda-test-task/pkg/logging"
)

type middleware struct {
	ProductInUse map[string]string
	mu           sync.RWMutex
}

func New() *middleware {
	pc := make(map[string]string)
	return &middleware{ProductInUse: pc}
}

func (m *middleware) SyncProducts(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.GetLogger()
		ipv4Addr := r.RemoteAddr
		logger.Info().Str("ip", ipv4Addr).Msg("getting request from")

		var products []models.Product
		if err := json.NewDecoder(r.Body).Decode(&products); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("сервер не смог десереализовать тело запроса"))
			return
		}

		codesInUse := make([]string, 0, len(products))
		for _, product := range products {
			m.mu.RLock()
			ip, ok := m.ProductInUse[product.Code]
			if !ok {
				m.mu.Lock()
				m.ProductInUse[product.Code] = ipv4Addr
				m.mu.Unlock()
				continue
			}
			m.mu.RUnlock()
			if ip != ipv4Addr {
				codesInUse = append(codesInUse, product.Code)
			}
		}

		if len(codesInUse) > 0 {
			var buf strings.Builder
			buf.WriteString("один или несколько товаров уже используются другой системой:\n")
			for _, code := range codesInUse {
				buf.WriteString(fmt.Sprintf("{%s}\n", code))
			}
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(buf.String()))
			return
		}
		next(w, r)
	}
}
