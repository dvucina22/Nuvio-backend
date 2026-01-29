package models

import (
	"time"
)

type TransactionStatistics struct {
	TotalRevenue            int64               `json:"totalRevenue"`
	TotalTransactions       int64               `json:"totalTransactions"`
	StatusBreakdown         []StatusStatistic   `json:"statusBreakdown"`
	AverageTransactionValue float64             `json:"averageTransactionValue"`
	RecentTransactions      []RecentTransaction `json:"recentTransactions"`
}

type StatusStatistic struct {
	Status     string  `json:"status"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

type RecentTransaction struct {
	ID           int64     `json:"id"`
	UserID       string    `json:"userId"`
	Type         string    `json:"type"`
	Status       string    `json:"status"`
	Amount       int64     `json:"amount"`
	CurrencyCode string    `json:"currencyCode"`
	CreatedAt    time.Time `json:"createdAt"`
}
