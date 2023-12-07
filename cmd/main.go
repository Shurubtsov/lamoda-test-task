package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Shurubtsov/lamoda-test-task/internal/adapters/db"
	"github.com/Shurubtsov/lamoda-test-task/internal/config"
	v1 "github.com/Shurubtsov/lamoda-test-task/internal/controller/http/v1"
	"github.com/Shurubtsov/lamoda-test-task/internal/controller/middleware"
	"github.com/Shurubtsov/lamoda-test-task/internal/domain/service"
	"github.com/Shurubtsov/lamoda-test-task/internal/domain/usecase"
	"github.com/Shurubtsov/lamoda-test-task/pkg/client/postgresql"
	"github.com/Shurubtsov/lamoda-test-task/pkg/logging"
)

func main() {
	logger := logging.GetLogger()
	logger.Info().Msg("initialize dependencies")
	cfg := config.GetConfig()

	pgClient, err := postgresql.NewClient(context.TODO(), 5)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed create new pgx client")
	}
	defer pgClient.Close()

	if err := db.Migrate(cfg.Service.MigrationsPath, cfg.Service.MigrationVersion); err != nil {
		logger.Fatal().Err(err).Msg("migration failed")
	}

	repo := db.New(pgClient)

	storageService := service.NewStorageService(repo)
	productService := service.NewProductService(repo)

	reservationUC := usecase.NewReservation(storageService, productService, repo)

	middleware := middleware.New()
	server := v1.NewServer(reservationUC, productService, productService, middleware.ProductInUse)

	mux := http.NewServeMux()
	mux.HandleFunc("/product/reservation", middleware.SyncProducts(server.ReservationHandler))
	mux.HandleFunc("/product/exemption", middleware.SyncProducts(server.ExemptionHandler))
	mux.HandleFunc("/storage/products", server.ReceivingProductsHandler)

	srv := http.Server{
		Addr:    cfg.Service.Address,
		Handler: mux,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
