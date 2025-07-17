package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Joshdike/subscriptions_aggregator/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
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

	// Create a new handler with the database connection pool
	h := handlers.New(pool)

	// Define routes and their handler functions
	r.Post("/subscriptions", h.CreateSubscription)
	r.Get("/subscriptions", h.GetSubscriptions)
	r.Get("/subscriptions/{id}", h.GetSubscriptionByID)
	r.Put("/subscriptions/{id}", h.UpdateSubscription)
	r.Delete("/subscriptions/{id}", h.DeleteSubscription)

	r.Get("/costs/{id}?from={from}&to={to}&service={service}", h.GetCostByDateRange)

	// Get the port from environment variable and start the server
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	fmt.Println("Server starting ...")
	// Start the HTTP server
	err = http.ListenAndServe(port, r)
	if err != nil {
		log.Fatal(err)
	}

}
