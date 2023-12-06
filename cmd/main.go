package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Shurubtsov/lamoda-test-task/internal/adapters/db"
	v1 "github.com/Shurubtsov/lamoda-test-task/internal/controller/http/v1"
	"github.com/Shurubtsov/lamoda-test-task/internal/controller/middleware"
	"github.com/Shurubtsov/lamoda-test-task/internal/domain/service"
	"github.com/Shurubtsov/lamoda-test-task/internal/domain/usecase/reservation"
	"github.com/Shurubtsov/lamoda-test-task/pkg/client/postgresql"
	"github.com/Shurubtsov/lamoda-test-task/pkg/logging"
)

func main() {
	logger := logging.GetLogger()
	logger.Info().Msg("initialize dependencies")

	pgClient, err := postgresql.NewClient(context.TODO(), 5)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed create new pgx client")
	}
	defer pgClient.Close()
	// if err := db.Migrate("file://./migrations", 1); err != nil {
	// 	logger.Fatal().Err(err).Msg("migration failed")
	// }
	repo := db.New(pgClient)
	storageService := service.NewStorageService(repo)
	productService := service.NewProductService(repo)

	reservationUC := reservation.New(storageService, productService, repo)

	middleware := middleware.New()
	server := v1.NewServer(reservationUC, middleware.ProductInUse)

	mux := http.NewServeMux()
	mux.HandleFunc("/reservation", middleware.SyncProducts(server.ReservationHandler))

	srv := http.Server{
		Addr:    "0.0.0.0:8888",
		Handler: mux,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
