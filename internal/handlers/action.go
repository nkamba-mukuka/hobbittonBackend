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

type SendMoneyRequest struct {
	RecipientEmail string  `json:"recipient_email"`
	Amount         float64 `json:"amount"`
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

func (h *TransactionHandler) SendMoney(w http.ResponseWriter, r *http.Request) {
	senderID := r.Context().Value("user_id").(uint)

	var req SendMoneyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start a transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}

	// Get sender's current balance
	var senderTransactions []models.Transaction
	if err := tx.Where("user_id = ?", senderID).Find(&senderTransactions).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Error fetching sender's transactions", http.StatusInternalServerError)
		return
	}
	senderBalance := models.CalculateBalance(senderTransactions)

	// Check if sender has enough balance
	if senderBalance < req.Amount {
		tx.Rollback()
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	// Get recipient
	var recipient models.User
	if err := tx.Where("email = ?", req.RecipientEmail).First(&recipient).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Recipient not found", http.StatusNotFound)
		return
	}

	// Create withdrawal transaction for sender
	senderTransaction := models.Transaction{
		UserID: senderID,
		Type:   models.Withdrawal,
		Amount: req.Amount,
		Date:   time.Now(),
	}
	if err := tx.Create(&senderTransaction).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Error creating sender's transaction", http.StatusInternalServerError)
		return
	}

	// Create deposit transaction for recipient
	recipientTransaction := models.Transaction{
		UserID: recipient.ID,
		Type:   models.Deposit,
		Amount: req.Amount,
		Date:   time.Now(),
	}
	if err := tx.Create(&recipientTransaction).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Error creating recipient's transaction", http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":               "Money sent successfully",
		"sender_transaction":    senderTransaction,
		"recipient_transaction": recipientTransaction,
	})
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
