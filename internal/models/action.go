package models

import (
	"time"
)

type TransactionType string

const (
	Deposit    TransactionType = "deposit"
	Withdrawal TransactionType = "withdrawal"
)

type Transaction struct {
	ID        uint            `json:"id" gorm:"primaryKey"`
	UserID    uint            `json:"user_id"`
	Type      TransactionType `json:"type"`
	Amount    float64         `json:"amount"`
	Date      time.Time      `json:"date"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	User      User           `json:"-" gorm:"foreignKey:UserID"`
}

// CalculateBalance calculates the current balance for a set of transactions
func CalculateBalance(transactions []Transaction) float64 {
	var balance float64
	for _, t := range transactions {
		switch t.Type {
		case Deposit:
			balance += t.Amount
		case Withdrawal:
			balance -= t.Amount
		}
	}
	return balance
}