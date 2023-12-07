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

// responder отправляет ответы клиенту
type responder struct {
	w http.ResponseWriter
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
	responder := &responder{w: w}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var products []models.Product
	if err := json.NewDecoder(r.Body).Decode(&products); err != nil {
		msg := "unable to deserialize the request body"
		responder.sendResponse(http.StatusUnprocessableEntity, msg, err)
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
		responder.sendResponse(
			http.StatusUnprocessableEntity,
			"can't continue reservation of products on storage",
			ErrAllProductsNotValid,
			responseOption("not_valid", notValidProducts),
		)
		return
	}
	// очищаем кэш используемых системой товаров
	defer func() {
		for _, product := range products {
			s.mu.Lock()
			delete(s.productInUse, product.Code)
			s.mu.Unlock()
		}
	}()
	reservedProducts, err := s.reservationUC.ProductReservation(ctx, products)
	if err != nil {
		responder.sendResponse(
			http.StatusInternalServerError,
			"reservation was ended with error",
			err,
		)
		return
	}

	if len(notValidProducts) > 0 {
		responder.sendResponse(
			http.StatusMultiStatus,
			"not at all products was reserved",
			ErrNotValidatedProducts,
			responseOption("reserved_products", reservedProducts),
			responseOption("not_valid", notValidProducts),
		)
		return
	}

	responder.sendResponse(
		http.StatusOK,
		"reservation successful complete",
		nil,
		responseOption("reserved_products", reservedProducts),
	)
}

func (s *server) ExemptionHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	ctx := context.Background()
	responder := &responder{w: w}

	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var products []models.Product
	if err := json.NewDecoder(r.Body).Decode(&products); err != nil {
		msg := "unable to deserialize the request body"
		responder.sendResponse(http.StatusUnprocessableEntity, msg, err)
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
		responder.sendResponse(
			http.StatusUnprocessableEntity,
			"can't continue exemption of products on storage",
			ErrAllProductsNotValid,
			responseOption("not_valid", notValidProducts),
		)
		return
	}
	// очищаем кэш используемых системой товаров
	defer func() {
		for _, product := range products {
			s.mu.Lock()
			delete(s.productInUse, product.Code)
			s.mu.Unlock()
		}
	}()
	exemptedProducts, err := s.exemptionUC.ProductExemption(ctx, products)
	if err != nil {
		responder.sendResponse(
			http.StatusInternalServerError,
			"exemption was ended with error",
			err,
		)
		return
	}

	if len(notValidProducts) > 0 {
		responder.sendResponse(
			http.StatusMultiStatus,
			"not at all products were exempt",
			ErrNotValidatedProducts,
			responseOption("exempted_products", exemptedProducts),
			responseOption("not_valid", notValidProducts),
		)
		return
	}

	responder.sendResponse(
		http.StatusOK,
		"exemption successful complete",
		nil,
		responseOption("exempted_products", exemptedProducts),
	)
}

// func (s *server) ReceivingProductsHandler(w http.ResponseWriter, r *http.Request) {
// 	logger := logging.GetLogger()
// 	ctx := context.Background()

// 	if r.Method != http.MethodGet {
// 		w.WriteHeader(http.StatusMethodNotAllowed)
// 		return
// 	}
// }

func (r *responder) sendResponse(code int, msg string, err error, args ...responseOpts) {
	response := make(map[string]any)
	response["status"] = http.StatusText(code)
	response["message"] = msg
	if err != nil {
		response["error"] = err.Error()
	}
	for _, arg := range args {
		response[arg.text] = arg.value
	}
	data, err := json.Marshal(response)
	if err != nil {
		r.w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprint("unable to serialize the response: ", err.Error())
		r.w.Write([]byte(msg))
	}
	r.w.Write(data)
}

type responseOpts struct {
	text  string
	value any
}

func responseOption(text string, val any) responseOpts {
	return responseOpts{
		text:  text,
		value: val,
	}
}
