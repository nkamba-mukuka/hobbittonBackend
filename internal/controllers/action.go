package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hobbiton-wallet-backend/internal/models"
	"gorm.io/gorm"
)

type TransactionHandler struct {
	db *gorm.DB
}

func NewTransactionHandler(db *gorm.DB) *TransactionHandler {
	return &TransactionHandler{
		db: db,
	}
}

type CreateTransactionRequest struct {
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)

	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	transaction := models.Transaction{
		UserID: userID,
		Type:   models.TransactionType(req.Type),
		Amount: req.Amount,
		Date:   time.Now(),
	}

	if err := h.db.Create(&transaction).Error; err != nil {
		http.Error(w, "Error creating transaction", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(transaction)
}

func (h *TransactionHandler) GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)

	var transactions []models.Transaction
	if err := h.db.Where("user_id = ?", userID).Order("date desc").Find(&transactions).Error; err != nil {
		http.Error(w, "Error fetching transactions", http.StatusInternalServerError)
		return
	}

	balance := models.CalculateBalance(transactions)

	response := struct {
		Balance      float64              `json:"balance"`
		Transactions []models.Transaction `json:"transactions"`
	}{
		Balance:      balance,
		Transactions: transactions,
	}

	json.NewEncoder(w).Encode(response)
}

//controllers
