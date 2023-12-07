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

type ExemptionUsecase interface {
	ProductExemption(ctx context.Context, products []models.Product) ([]models.Product, error)
}

type server struct {
	mu            sync.Mutex
	productInUse  map[string]string
	reservationUC ReservationUsecase
	exemptionUC   ExemptionUsecase
}

func NewServer(ruc ReservationUsecase, euc ExemptionUsecase, pie map[string]string) *server {
	return &server{
		productInUse:  pie,
		reservationUC: ruc,
		exemptionUC:   euc,
	}
}

func (s *server) ReservationHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	ctx := context.Background()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var products []models.Product
	if err := json.NewDecoder(r.Body).Decode(&products); err != nil {
		msg := "unable to deserialize the request body"
		sendResponse(w, nil, nil, nil, msg, http.StatusUnprocessableEntity, err)
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
		sendResponse(w, notValidProducts, nil, nil, "can't continue reservation of products on storage", http.StatusUnprocessableEntity, ErrAllProductsNotValid)
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
		sendResponse(w, nil, nil, nil, "reservation was ended with error", http.StatusInternalServerError, err)
		return
	}

	if len(notValidProducts) > 0 {
		sendResponse(w, notValidProducts, filledProducts, nil, "not at all products were reserved", http.StatusMultiStatus, ErrNotValidatedProducts)
		return
	}

	sendResponse(w, notValidProducts, filledProducts, nil, "reservation successful complete", http.StatusOK, nil)
}

func (s *server) ExemptionHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	ctx := context.Background()

	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var products []models.Product
	if err := json.NewDecoder(r.Body).Decode(&products); err != nil {
		msg := "unable to deserialize the request body"
		sendResponse(w, nil, nil, nil, msg, http.StatusUnprocessableEntity, err)
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
		sendResponse(w, notValidProducts, nil, nil, "can't continue exemption of products on storage", http.StatusUnprocessableEntity, ErrAllProductsNotValid)
		return
	}
	defer func() {
		for _, product := range products {
			s.mu.Lock()
			delete(s.productInUse, product.Code)
			s.mu.Unlock()
		}
	}()
	filledProducts, err := s.exemptionUC.ProductExemption(ctx, products)
	if err != nil {
		sendResponse(w, nil, nil, nil, "exemption was ended with error", http.StatusInternalServerError, err)
		return
	}

	if len(notValidProducts) > 0 {
		sendResponse(w, notValidProducts, nil, filledProducts, "not at all products were exempt", http.StatusMultiStatus, ErrNotValidatedProducts)
		return
	}

	sendResponse(w, notValidProducts, nil, filledProducts, "exemption successful complete", http.StatusOK, nil)
}

func sendResponse(
	w http.ResponseWriter,
	notValid []models.Product,
	reserved []models.Product,
	exempted []models.Product,
	msg string, code int,
	err error,
) {
	status := http.StatusText(code)
	response := models.NewProductsResponse(notValid, reserved, exempted, status, msg)
	if err != nil {
		response.Error = err.Error()
	}

	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(&response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprint("unable to serialize the response: ", err.Error())
		w.Write([]byte(msg))
	}
}
