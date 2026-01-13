package models

import (
	"fmt"
	"time"
)

type TransactionFilter struct {
	Search *string `json:"search,omitempty"`
	
	DateFrom *string `json:"dateFrom,omitempty"`
	DateTo   *string `json:"dateTo,omitempty"`
	
	Statuses []string `json:"statuses,omitempty"`
	Types    []string `json:"types,omitempty"`
	
	AmountMin *int64 `json:"amountMin,omitempty"`
	AmountMax *int64 `json:"amountMax,omitempty"`
	
	ProductCountMin *int `json:"productCountMin,omitempty"`
	ProductCountMax *int `json:"productCountMax,omitempty"`
	
	Page     int `json:"page,omitempty"`
	PageSize int `json:"pageSize,omitempty"`
	
	UserID *string `json:"userId,omitempty"`
}

type AdminTransactionFilter struct {
	Search *string `json:"search,omitempty"`
	
	DateFrom *string `json:"dateFrom,omitempty"`
	DateTo   *string `json:"dateTo,omitempty"`
	
	Statuses []string `json:"statuses,omitempty"`
	Types    []string `json:"types,omitempty"`
	
	AmountMin *int64 `json:"amountMin,omitempty"`
	AmountMax *int64 `json:"amountMax,omitempty"`
	
	ProductCountMin *int `json:"productCountMin,omitempty"`
	ProductCountMax *int `json:"productCountMax,omitempty"`
	
	UserID *string `json:"userId,omitempty"`
	
	Page     int `json:"page,omitempty"`
	PageSize int `json:"pageSize,omitempty"`
}

func (f *TransactionFilter) Validate() error {
	var dateFrom, dateTo *time.Time
	
	if f.DateFrom != nil && *f.DateFrom != "" {
		parsed, err := time.Parse(time.RFC3339, *f.DateFrom)
		if err != nil {
			return fmt.Errorf("%w: invalid dateFrom format, use RFC3339 (e.g. 2024-01-01T00:00:00Z)", ErrInvalidFilter)
		}
		dateFrom = &parsed
	}
	
	if f.DateTo != nil && *f.DateTo != "" {
		parsed, err := time.Parse(time.RFC3339, *f.DateTo)
		if err != nil {
			return fmt.Errorf("%w: invalid dateTo format, use RFC3339 (e.g. 2024-12-31T23:59:59Z)", ErrInvalidFilter)
		}
		dateTo = &parsed
	}
	
	if dateFrom != nil && dateTo != nil && dateFrom.After(*dateTo) {
		return fmt.Errorf("%w: dateFrom cannot be after dateTo", ErrInvalidFilter)
	}
	
	if f.AmountMin != nil && f.AmountMax != nil && *f.AmountMin > *f.AmountMax {
		return fmt.Errorf("%w: amountMin cannot be greater than amountMax", ErrInvalidFilter)
	}
	
	if f.ProductCountMin != nil && f.ProductCountMax != nil && *f.ProductCountMin > *f.ProductCountMax {
		return fmt.Errorf("%w: productCountMin cannot be greater than productCountMax", ErrInvalidFilter)
	}
	
	if len(f.Types) > 0 {
		validTypes := map[string]bool{
			"SALE": true,
			"VOID": true,
		}
		
		for _, t := range f.Types {
			if !validTypes[t] {
				return fmt.Errorf("%w: invalid transaction type '%s'", ErrInvalidFilter, t)
			}
		}
	}
	
	if f.Page < 1 {
		f.Page = 1
	}
	
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	
	return nil
}

func (f *AdminTransactionFilter) Validate() error {
	var dateFrom, dateTo *time.Time
	
	if f.DateFrom != nil && *f.DateFrom != "" {
		parsed, err := time.Parse(time.RFC3339, *f.DateFrom)
		if err != nil {
			return fmt.Errorf("%w: invalid dateFrom format, use RFC3339", ErrInvalidFilter)
		}
		dateFrom = &parsed
	}
	
	if f.DateTo != nil && *f.DateTo != "" {
		parsed, err := time.Parse(time.RFC3339, *f.DateTo)
		if err != nil {
			return fmt.Errorf("%w: invalid dateTo format, use RFC3339", ErrInvalidFilter)
		}
		dateTo = &parsed
	}
	
	if dateFrom != nil && dateTo != nil && dateFrom.After(*dateTo) {
		return fmt.Errorf("%w: dateFrom cannot be after dateTo", ErrInvalidFilter)
	}
	
	if f.AmountMin != nil && f.AmountMax != nil && *f.AmountMin > *f.AmountMax {
		return fmt.Errorf("%w: amountMin cannot be greater than amountMax", ErrInvalidFilter)
	}
	
	if f.ProductCountMin != nil && f.ProductCountMax != nil && *f.ProductCountMin > *f.ProductCountMax {
		return fmt.Errorf("%w: productCountMin cannot be greater than productCountMax", ErrInvalidFilter)
	}
	
	if len(f.Types) > 0 {
		validTypes := map[string]bool{
			"SALE": true,
			"VOID": true,
		}
		
		for _, t := range f.Types {
			if !validTypes[t] {
				return fmt.Errorf("%w: invalid transaction type '%s'", ErrInvalidFilter, t)
			}
		}
	}
	
	if f.Page < 1 {
		f.Page = 1
	}
	
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	
	return nil
}

func (f *TransactionFilter) GetParsedDates() (*time.Time, *time.Time) {
	var dateFrom, dateTo *time.Time
	
	if f.DateFrom != nil && *f.DateFrom != "" {
		if parsed, err := time.Parse(time.RFC3339, *f.DateFrom); err == nil {
			dateFrom = &parsed
		}
	}
	
	if f.DateTo != nil && *f.DateTo != "" {
		if parsed, err := time.Parse(time.RFC3339, *f.DateTo); err == nil {
			dateTo = &parsed
		}
	}
	
	return dateFrom, dateTo
}

func (f *AdminTransactionFilter) GetParsedDates() (*time.Time, *time.Time) {
	var dateFrom, dateTo *time.Time
	
	if f.DateFrom != nil && *f.DateFrom != "" {
		if parsed, err := time.Parse(time.RFC3339, *f.DateFrom); err == nil {
			dateFrom = &parsed
		}
	}
	
	if f.DateTo != nil && *f.DateTo != "" {
		if parsed, err := time.Parse(time.RFC3339, *f.DateTo); err == nil {
			dateTo = &parsed
		}
	}
	
	return dateFrom, dateTo
}