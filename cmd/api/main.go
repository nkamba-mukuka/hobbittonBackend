package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/hobbiton-wallet-backend/internal/handlers"
	"github.com/hobbiton-wallet-backend/internal/middleware"
	"github.com/hobbiton-wallet-backend/internal/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Database connection
	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database")
	}
	err = db.AutoMigrate(&models.User{})        // Add other models here
	err = db.AutoMigrate(&models.Transaction{}) // Add other models here
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize handlers
	jwtKey := []byte(os.Getenv("JWT_SECRET"))
	authHandler := handlers.NewAuthHandler(db, jwtKey)
	transactionHandler := handlers.NewTransactionHandler(db)
	authMiddleware := middleware.NewAuthMiddleware(jwtKey)

	// Router setup
	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	r.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")

	// Protected routes
	api := r.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware.Authenticate)
	api.HandleFunc("/transactions", transactionHandler.GetUserTransactions).Methods("GET")
	api.HandleFunc("/transactions", transactionHandler.Create).Methods("POST")
	api.HandleFunc("/transactions/send", transactionHandler.SendMoney).Methods("POST")

	// CORS middleware
	handler := corsMiddleware(r)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
