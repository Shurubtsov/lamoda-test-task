package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sync"

	"github.com/Shurubtsov/lamoda-test-task/internal/domain/models"
	"github.com/Shurubtsov/lamoda-test-task/pkg/logging"
)

var (
	ErrNotValidatedProducts = errors.New("some of products was not validated")
	ErrAllProductsNotValid  = errors.New("all products not validated")
)

type ReservationUsecase interface {
	ProductReservation(ctx context.Context, products []models.Product) ([]models.Product, error)
}

type server struct {
	mu            sync.Mutex
	productInUse  map[string]string
	reservationUC ReservationUsecase
}

func NewServer(ruc ReservationUsecase, pie map[string]string) *server {
	return &server{productInUse: pie, reservationUC: ruc}
}

func (s *server) ReservationHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	ctx := context.Background()

	var products []models.Product
	if err := json.NewDecoder(r.Body).Decode(&products); err != nil {
		msg := "unable to deserialize the request body"
		response(w, nil, nil, msg, http.StatusUnprocessableEntity, err)
		return
	}

	notValidProducts := make([]models.Product, 0, len(products))
	products = slices.DeleteFunc(products, func(p models.Product) bool {
		if err := p.Validate(); err != nil {
			logger.Warn().Err(err).Str("code", p.Code).Msg("one of product not validated")
			notValidProducts = append(notValidProducts, p)
			return true
		}
		return false
	})
	if !(len(products) > 0) {
		response(w, &notValidProducts, nil, "can't continue reservation of products on storage", http.StatusUnprocessableEntity, ErrAllProductsNotValid)
		return
	}
	defer func() {
		for _, product := range products {
			s.mu.Lock()
			delete(s.productInUse, product.Code)
			s.mu.Unlock()
		}
	}()
	filledProducts, err := s.reservationUC.ProductReservation(ctx, products)
	if err != nil {
		response(w, nil, nil, "reservation was ended with error", http.StatusInternalServerError, err)
		return
	}

	if len(notValidProducts) > 0 {
		response(w, &notValidProducts, &filledProducts, "not at all products were reserved", http.StatusMultiStatus, ErrNotValidatedProducts)
		return
	}

	response(w, &notValidProducts, &filledProducts, "reservation successful complete", http.StatusOK, nil)
}

func response(w http.ResponseWriter, notValidProducts *[]models.Product, successfulIDs *[]models.Product, msg string, code int, err error) {
	var response models.ReservationResponse

	response.Status = http.StatusText(code)
	response.Message = msg
	w.WriteHeader(code)

	if err != nil {
		response.Error = err.Error()
	}

	if notValidProducts != nil {
		response.NotValid = *notValidProducts
	}

	if successfulIDs != nil {
		response.Successful = *successfulIDs
	}

	if err := json.NewEncoder(w).Encode(&response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprint("unable to serialize the response: ", err.Error())
		w.Write([]byte(msg))
	}
}
