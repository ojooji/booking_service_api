package router

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/ojooji/booking-service-api/internal/config"
	"github.com/ojooji/booking-service-api/internal/handler"
	"github.com/ojooji/booking-service-api/internal/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func New(cfg *config.Config, authH *handler.AuthHandler, userH *handler.UserHandler, svcH *handler.ServiceHandler, empH *handler.EmployeeHandler, bookingH *handler.BookingHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(middleware.NewCORS(cfg.CORSAllowedOrigins))

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", authH.Register)
		r.Post("/auth/login", authH.Login)

		r.Get("/services", svcH.List)
		r.Get("/services/{id}", svcH.GetByID)
		r.Get("/employees", empH.List)
		r.Get("/employees/{id}", empH.GetByID)
		r.Get("/employees/{id}/slots", bookingH.GetSlots)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.JWTSecret))

			r.Get("/bookings", bookingH.ListMy)
			r.Post("/bookings", bookingH.Create)
			r.Get("/bookings/{id}", bookingH.GetByID)
			r.Put("/bookings/{id}/cancel", bookingH.Cancel)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AdminOnly)

				r.Get("/users", userH.List)
				r.Get("/users/{id}", userH.GetByID)
				r.Put("/users/{id}", userH.Update)
				r.Delete("/users/{id}", userH.Delete)

				r.Post("/services", svcH.Create)
				r.Put("/services/{id}", svcH.Update)
				r.Delete("/services/{id}", svcH.Delete)

				r.Post("/employees", empH.Create)
				r.Put("/employees/{id}", empH.Update)
				r.Delete("/employees/{id}", empH.Delete)

				r.Get("/bookings/all", bookingH.ListAll)
			})
		})
	})

	return r
}
