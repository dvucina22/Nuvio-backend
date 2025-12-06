package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/transaction-service/pkg/models"
)

type TransactionRepository interface {
	CreateSale(ctx context.Context, tx *models.Transaction, products []*models.TransactionProduct) error
}

type transactionRepo struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) CreateSale(
	ctx context.Context,
	trx *models.Transaction,
	products []*models.TransactionProduct,
) error {
	dbTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin db transaction: %w", err)
	}
	defer dbTx.Rollback()

	insertTxQuery := `
		INSERT INTO transaction.transactions (
			user_id,
			bank_card_id,
			type,
			status,
			pan_masked,
			card_first_digit,
			card_expiration_yy,
			card_expiration_mm,
			processing_code,
			amount,
			currency_code,
			stan,
			transaction_time,
			transaction_date,
			rrn,
			terminal_tid,
			merchant_mid,
			host_type,
			original_transaction_id,
			request_payload
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20
		)
		RETURNING id, created_at, updated_at
	`

	var bankCardID interface{}
	if trx.BankCardID != nil {
		bankCardID = *trx.BankCardID
	} else {
		bankCardID = nil
	}

	var originalTxID interface{}
	if trx.OriginalTransactionID != nil {
		originalTxID = *trx.OriginalTransactionID
	} else {
		originalTxID = nil
	}

	err = dbTx.QueryRowContext(
		ctx,
		insertTxQuery,
		trx.UserID,
		bankCardID,
		trx.Type,
		trx.Status,
		trx.PANMasked,
		trx.CardFirstDigit,
		trx.CardExpirationYY,
		trx.CardExpirationMM,
		trx.ProcessingCode,
		trx.Amount,
		trx.CurrencyCode,
		trx.STAN,
		trx.TransactionTime,
		trx.TransactionDate,
		trx.RRN,
		trx.TerminalTID,
		trx.MerchantMID,
		trx.HostType,
		originalTxID,
		trx.RequestPayload,
	).Scan(
		&trx.ID,
		&trx.CreatedAt,
		&trx.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert sale transaction: %w", err)
	}

	if len(products) > 0 {
		insertProductQuery := `
			INSERT INTO transaction.transaction_products (
				transaction_id,
				product_id,
				unit_price,
				quantity,
				product_name,
				product_sku
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`

		for _, p := range products {
			p.TransactionID = trx.ID

			err = dbTx.QueryRowContext(
				ctx,
				insertProductQuery,
				p.TransactionID,
				p.ProductID,
				p.UnitPrice,
				p.Quantity,
				p.ProductName,
				p.ProductSKU,
			).Scan(&p.ID)
			if err != nil {
				return fmt.Errorf("failed to insert transaction product (product_id=%d): %w", p.ProductID, err)
			}
		}
	}

	if err := dbTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit sale transaction: %w", err)
	}

	return nil
}
