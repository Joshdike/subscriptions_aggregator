// @title Subscriptions Aggregator API
// @version 1.0
// @description API for managing user subscriptions and cost calculations
// @contact.name API Support
// @contact.email dikejoshua@gmail.com
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey AdminAuth
// @in header
// @name secret-key
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/Joshdike/subscriptions_aggregator/docs"
	"github.com/Joshdike/subscriptions_aggregator/internal/handlers"
	mw "github.com/Joshdike/subscriptions_aggregator/internal/middleware"
	"github.com/Joshdike/subscriptions_aggregator/internal/repository/pg"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	// Create a new context
	ctx := context.Background()

	// Establish a new connection pool to the database
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Check if the database connection is alive
	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	// Initialize a new router using Chi
	r := chi.NewRouter()
	r.Use(middleware.Logger)    // Middleware for logging
	r.Use(middleware.Recoverer) // Middleware for recovering from panics

	// Swagger UI route
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json")))

	// Create a new Subscription repository
	subRepo := pg.NewSubscriptionRepo(pool)

	// Create a new Subscription handler
	h := handlers.New(subRepo)

	// Define routes and their handler functions
	r.Post("/subscriptions", h.CreateSubscription)
	r.Get("/subscriptions/user/{user_id}", h.GetSubscriptionByUserID)
	r.Get("/subscriptions/{id}", h.GetSubscriptionByID)
	r.Post("/subscriptions/{id}", h.RenewOrExtendSubscription)
	r.Patch("/subscriptions/{id}", h.DeleteSubscription)

	r.Get("/costs/{user_id}", h.GetCostByDateRange)

	// Admin route
	r.With(mw.AdminSecretMiddleware(os.Getenv("SECRET_KEY"))).Get("/subscriptions", h.GetSubscriptions)

	// Get the port from environment variable and start the server
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	fmt.Println("Server starting ...")
	// Start the HTTP server
	err = http.ListenAndServe(port, r)
	if err != nil {
		log.Fatal(err)
	}

}
