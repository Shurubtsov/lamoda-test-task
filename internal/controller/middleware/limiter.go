package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		var products []models.Product
		if err := json.Unmarshal(body, &products); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			msg := fmt.Sprint("unable to deserialize the request body: ", err.Error())
			w.Write([]byte(msg))
			return
		}

		codesInUse := make([]string, 0, len(products))
		m.mu.RLock()
		for _, product := range products {
			ip, ok := m.ProductInUse[product.Code]
			if !ok {
				go func() {
					m.mu.Lock()
					m.ProductInUse[product.Code] = ipv4Addr
					m.mu.Unlock()
				}()
				continue
			}
			if ip != ipv4Addr {
				codesInUse = append(codesInUse, product.Code)
			}
		}
		m.mu.RUnlock()

		if len(codesInUse) > 0 {
			var buf strings.Builder
			buf.WriteString("one or more products are already in use by another system:\n")
			for _, code := range codesInUse {
				buf.WriteString(fmt.Sprintf("{%s}\n", code))
			}
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(buf.String()))
			return
		}
		w.Header().Add("Content-Type", "application/json")
		next(w, r)
	}
}
