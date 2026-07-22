package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	_ "github.com/ojooji/booking-service-api/docs" // swagger spec registration
	"github.com/ojooji/booking-service-api/internal/config"
	"github.com/ojooji/booking-service-api/internal/handler"
	pgxrepo "github.com/ojooji/booking-service-api/internal/repository/pgx"
	"github.com/ojooji/booking-service-api/internal/router"
	"github.com/ojooji/booking-service-api/internal/service"
)

// @title Booking Service API
// @version 1.0
// @description RESTful API for managing appointment scheduling and reservation workflows
// @contact.name API Support
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	_ = godotenv.Load() // .env is optional; real env vars take precedence

	cfg := config.Load()

	ctx := context.Background()

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	userRepo := pgxrepo.NewUserRepository(dbPool)
	serviceRepo := pgxrepo.NewServiceRepository(dbPool)
	employeeRepo := pgxrepo.NewEmployeeRepository(dbPool)
	bookingRepo := pgxrepo.NewBookingRepository(dbPool)

	authSvc := service.NewAuthService(userRepo, cfg)
	userSvc := service.NewUserService(userRepo)
	serviceSvc := service.NewServiceService(serviceRepo)
	employeeSvc := service.NewEmployeeService(employeeRepo, userRepo)
	bookingSvc := service.NewBookingService(bookingRepo, employeeRepo, serviceRepo)

	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	svcH := handler.NewServiceHandler(serviceSvc)
	empH := handler.NewEmployeeHandler(employeeSvc)
	bookingH := handler.NewBookingHandler(bookingSvc)

	r := router.New(cfg, authH, userH, svcH, empH, bookingH)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced shutdown: %v", err)
	}

	log.Println("server stopped")
}
