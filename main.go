package main

import (
	"blog-api/config"
	"blog-api/handlers"
	"blog-api/storage"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {

	//New update for db
	//Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize PostgreSQL store
	store, err := storage.NewPostgresStore(cfg.GetDBConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer store.Close()

	// Initialize database (create tables)
	if err := store.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	log.Println("Database initialized successfully")

	// Initialize handlers
	postHandler := handlers.NewPostStoreHandler(store)

	// Create router
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Posts endpoints
	api.HandleFunc("/posts", postHandler.GetAllPosts).Methods("GET")
	api.HandleFunc("/posts", postHandler.CreatePost).Methods("POST")
	api.HandleFunc("/posts/{id}", postHandler.GetPost).Methods("GET")
	api.HandleFunc("/posts/{id}", postHandler.UpdatePost).Methods("PUT")
	api.HandleFunc("/posts/{id}", postHandler.DeletePost).Methods("DELETE")

	// Health check with DB connectivity test
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Test database connection
		if _, err := store.GetAll(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			jsonResponse(w, map[string]string{
				"status":  "unhealthy",
				"error":   err.Error(),
				"service": "database",
			})
			return
		}

		jsonResponse(w, map[string]string{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"database":  "connected",
		})
	}).Methods("GET")

	// Middleware
	r.Use(loggingMiddleware)
	r.Use(jsonContentTypeMiddleware)

	// CORS middleware
	r.Use(mux.CORSMethodMiddleware(r))

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in goroutine
	go func() {
		log.Printf("Server starting on :%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
