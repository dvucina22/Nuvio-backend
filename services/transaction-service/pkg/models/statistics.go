package models

type TransactionStatistics struct {
	TotalRevenue            int64             `json:"totalRevenue"`
	TotalTransactions       int64             `json:"totalTransactions"`
	StatusBreakdown         []StatusStatistic `json:"statusBreakdown"`
	AverageTransactionValue float64           `json:"averageTransactionValue"`
}

type StatusStatistic struct {
	Status     string  `json:"status"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}
