package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/transaction-service/pkg/models"
)

type TransactionRepository interface {
	CreateSale(ctx context.Context, tx *models.Transaction, products []*models.TransactionProduct) error
	GetTransactionByID(ctx context.Context, userID uuid.UUID, txID int64) (*models.Transaction, error)
	GetTransactionByIDAdmin(ctx context.Context, txID int64) (*models.Transaction, error)
	VoidTransaction(ctx context.Context, voidTx *models.Transaction) error
	HasVoidForOriginal(ctx context.Context, originalTxID int64) (bool, error)
	GetTransactionWithProducts(ctx context.Context, userID uuid.UUID, txID int64) (*models.Transaction, []*models.TransactionProduct, error)
    GetUserTransactionDetail(ctx context.Context, userID uuid.UUID, txID int64) (*models.TransactionDetail, error)
    GetAdminTransactionDetail(ctx context.Context, txID int64) (*models.AdminTransactionDetail, error)
    GetFilteredUserTransactions(ctx context.Context, filter *models.TransactionFilter) ([]*models.TransactionListItem, int64, error)
	GetFilteredAdminTransactions(ctx context.Context, filter *models.AdminTransactionFilter) ([]*models.AdminTransactionListItem, int64, error)
}

type transactionRepo struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) CreateSale(ctx context.Context, trx *models.Transaction, products []*models.TransactionProduct) error {
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
			request_payload,
			response_payload,
			response_code,
			auth_code
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23
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
		trx.ResponsePayload,
		trx.ResponseCode,
		trx.AuthCode,
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

func (r *transactionRepo) GetTransactionByID(ctx context.Context, userID uuid.UUID, txID int64) (*models.Transaction, error) {
	q := `
        SELECT 
            id,
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
            request_payload,
            created_at,
            updated_at
        FROM transaction.transactions
        WHERE id = $1 AND user_id = $2
        LIMIT 1
    `

	var (
		tx           models.Transaction
		bankCardID   sql.NullInt64
		originalTxID sql.NullInt64
	)

	err := r.db.QueryRowContext(ctx, q, txID, userID).Scan(
		&tx.ID,
		&tx.UserID,
		&bankCardID,
		&tx.Type,
		&tx.Status,
		&tx.PANMasked,
		&tx.CardFirstDigit,
		&tx.CardExpirationYY,
		&tx.CardExpirationMM,
		&tx.ProcessingCode,
		&tx.Amount,
		&tx.CurrencyCode,
		&tx.STAN,
		&tx.TransactionTime,
		&tx.TransactionDate,
		&tx.RRN,
		&tx.TerminalTID,
		&tx.MerchantMID,
		&tx.HostType,
		&originalTxID,
		&tx.RequestPayload,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if bankCardID.Valid {
		id := bankCardID.Int64
		tx.BankCardID = &id
	} else {
		tx.BankCardID = nil
	}

	if originalTxID.Valid {
		id := originalTxID.Int64
		tx.OriginalTransactionID = &id
	} else {
		tx.OriginalTransactionID = nil
	}

	return &tx, nil
}

func (r *transactionRepo) GetTransactionByIDAdmin(ctx context.Context, txID int64) (*models.Transaction, error) {
	q := `
        SELECT 
            id,
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
            request_payload,
            created_at,
            updated_at
        FROM transaction.transactions
        WHERE id = $1
        LIMIT 1
    `

	var (
		tx           models.Transaction
		bankCardID   sql.NullInt64
		originalTxID sql.NullInt64
	)

	err := r.db.QueryRowContext(ctx, q, txID).Scan(
		&tx.ID,
		&tx.UserID,
		&bankCardID,
		&tx.Type,
		&tx.Status,
		&tx.PANMasked,
		&tx.CardFirstDigit,
		&tx.CardExpirationYY,
		&tx.CardExpirationMM,
		&tx.ProcessingCode,
		&tx.Amount,
		&tx.CurrencyCode,
		&tx.STAN,
		&tx.TransactionTime,
		&tx.TransactionDate,
		&tx.RRN,
		&tx.TerminalTID,
		&tx.MerchantMID,
		&tx.HostType,
		&originalTxID,
		&tx.RequestPayload,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if bankCardID.Valid {
		id := bankCardID.Int64
		tx.BankCardID = &id
	} else {
		tx.BankCardID = nil
	}

	if originalTxID.Valid {
		id := originalTxID.Int64
		tx.OriginalTransactionID = &id
	} else {
		tx.OriginalTransactionID = nil
	}

	return &tx, nil
}

func (r *transactionRepo) VoidTransaction(ctx context.Context, voidTx *models.Transaction) error {
	dbTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer dbTx.Rollback()

	var bankCardID interface{}
	if voidTx.BankCardID != nil {
		bankCardID = *voidTx.BankCardID
	} else {
		bankCardID = nil
	}

	if voidTx.OriginalTransactionID == nil {
		return fmt.Errorf("original transaction id is nil")
	}
	originalTxID := *voidTx.OriginalTransactionID

	insertQ := `
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
            request_payload,
			response_payload,
			response_code,
			auth_code
        )
        VALUES (
            $1,$2,$3,$4,
            $5,$6,$7,$8,$9,
            $10,$11,
            $12,$13,$14,$15,
            $16,$17,$18,
            $19,$20, $21, $22, $23
        )
        RETURNING id, created_at, updated_at
    `

	err = dbTx.QueryRowContext(
		ctx,
		insertQ,
		voidTx.UserID,
		bankCardID,
		voidTx.Type,
		voidTx.Status,
		voidTx.PANMasked,
		voidTx.CardFirstDigit,
		voidTx.CardExpirationYY,
		voidTx.CardExpirationMM,
		voidTx.ProcessingCode,
		voidTx.Amount,
		voidTx.CurrencyCode,
		voidTx.STAN,
		voidTx.TransactionTime,
		voidTx.TransactionDate,
		voidTx.RRN,
		voidTx.TerminalTID,
		voidTx.MerchantMID,
		voidTx.HostType,
		originalTxID,
		voidTx.RequestPayload,
		voidTx.ResponsePayload,
		voidTx.ResponseCode,
		voidTx.AuthCode,
	).Scan(
		&voidTx.ID,
		&voidTx.CreatedAt,
		&voidTx.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert void transaction: %w", err)
	}

	if err := dbTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit void tx: %w", err)
	}

	return nil
}

func (r *transactionRepo) HasVoidForOriginal(ctx context.Context, originalTxID int64) (bool, error) {
	q := `
        SELECT EXISTS (
            SELECT 1 
            FROM transaction.transactions
            WHERE original_transaction_id = $1
              AND type = 'VOID'
        )
    `

	var exists bool
	if err := r.db.QueryRowContext(ctx, q, originalTxID).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check void existence: %w", err)
	}
	return exists, nil
}

func (r *transactionRepo) GetTransactionWithProducts(ctx context.Context, userID uuid.UUID, txID int64) (*models.Transaction, []*models.TransactionProduct, error) {
	tx, err := r.GetTransactionByID(ctx, userID, txID)
	if err != nil {
		return nil, nil, err
	}
	if tx == nil {
		return nil, nil, nil
	}

	q := `
        SELECT 
            id,
            transaction_id,
            product_id,
            unit_price,
            quantity,
            product_name,
            product_sku
        FROM transaction.transaction_products
        WHERE transaction_id = $1
        ORDER BY id
    `

	rows, err := r.db.QueryContext(ctx, q, txID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query transaction products: %w", err)
	}
	defer rows.Close()

	var products []*models.TransactionProduct

	for rows.Next() {
		var (
			p        models.TransactionProduct
			prodName sql.NullString
			prodSKU  sql.NullString
		)

		if err := rows.Scan(
			&p.ID,
			&p.TransactionID,
			&p.ProductID,
			&p.UnitPrice,
			&p.Quantity,
			&prodName,
			&prodSKU,
		); err != nil {
			return nil, nil, fmt.Errorf("failed to scan transaction product: %w", err)
		}

		if prodName.Valid {
			name := prodName.String
			p.ProductName = &name
		}
		if prodSKU.Valid {
			sku := prodSKU.String
			p.ProductSKU = &sku
		}

		products = append(products, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iteration error for transaction products: %w", err)
	}

	return tx, products, nil
}

func (r *transactionRepo) GetUserTransactionDetail(ctx context.Context, userID uuid.UUID, txID int64) (*models.TransactionDetail, error) {
    q := `
        SELECT 
            t.id,
            t.status,
            t.amount,
            t.currency_code,
            t.pan_masked,
            t.card_expiration_yy,
            t.card_expiration_mm,
            t.created_at,
            t.transaction_date,
            t.transaction_time
        FROM transaction.transactions t
        WHERE t.id = $1 AND t.user_id = $2 AND t.type = 'SALE'
        LIMIT 1
    `
    
    var detail models.TransactionDetail
    err := r.db.QueryRowContext(ctx, q, txID, userID).Scan(
        &detail.ID,
        &detail.Status,
        &detail.Amount,
        &detail.CurrencyCode,
        &detail.PANMasked,
        &detail.CardExpirationYY,
        &detail.CardExpirationMM,
        &detail.CreatedAt,
        &detail.TransactionDate,
        &detail.TransactionTime,
    )
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get transaction detail: %w", err)
    }
    
    productsQ := `
        SELECT 
            id,
            product_id,
            unit_price,
            quantity,
            product_name,
            product_sku
        FROM transaction.transaction_products
        WHERE transaction_id = $1
        ORDER BY id
    `
    
    rows, err := r.db.QueryContext(ctx, productsQ, txID)
    if err != nil {
        return nil, fmt.Errorf("failed to query products: %w", err)
    }
    defer rows.Close()
    
    for rows.Next() {
        var p models.TransactionProductDetail
        var name, sku sql.NullString
        
        if err := rows.Scan(
            &p.ID,
            &p.ProductID,
            &p.UnitPrice,
            &p.Quantity,
            &name,
            &sku,
        ); err != nil {
            return nil, fmt.Errorf("failed to scan product: %w", err)
        }
        
        if name.Valid {
            p.ProductName = &name.String
        }
        if sku.Valid {
            p.ProductSKU = &sku.String
        }
        
        p.LineTotal = p.UnitPrice * int64(p.Quantity)
        detail.Products = append(detail.Products, p)
    }
    
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iteration error: %w", err)
    }
    
    return &detail, nil
}

func (r *transactionRepo) GetAdminTransactionDetail(ctx context.Context, txID int64) (*models.AdminTransactionDetail, error) {
    q := `
        SELECT 
            t.id,
            t.user_id,
            t.bank_card_id,
            t.type,
            t.status,
            t.pan_masked,
            t.card_first_digit,
            t.card_expiration_yy,
            t.card_expiration_mm,
            t.processing_code,
            t.amount,
            t.currency_code,
            t.stan,
            t.transaction_time,
            t.transaction_date,
            t.rrn,
            t.terminal_tid,
            t.merchant_mid,
            t.host_type,
            t.original_transaction_id,
            t.response_code,
            t.auth_code,
            t.created_at,
            t.updated_at
        FROM transaction.transactions t
        WHERE t.id = $1
        LIMIT 1
    `
    
    var detail models.AdminTransactionDetail
    var bankCardID, originalTxID sql.NullInt64
    var responseCode, authCode sql.NullString
    
    err := r.db.QueryRowContext(ctx, q, txID).Scan(
        &detail.ID,
        &detail.UserID,
        &bankCardID,
        &detail.Type,
        &detail.Status,
        &detail.PANMasked,
        &detail.CardFirstDigit,
        &detail.CardExpirationYY,
        &detail.CardExpirationMM,
        &detail.ProcessingCode,
        &detail.Amount,
        &detail.CurrencyCode,
        &detail.STAN,
        &detail.TransactionTime,
        &detail.TransactionDate,
        &detail.RRN,
        &detail.TerminalTID,
        &detail.MerchantMID,
        &detail.HostType,
        &originalTxID,
        &responseCode,
        &authCode,
        &detail.CreatedAt,
        &detail.UpdatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get admin transaction detail: %w", err)
    }
    
    if bankCardID.Valid {
        id := bankCardID.Int64
        detail.BankCardID = &id
    }
    if originalTxID.Valid {
        id := originalTxID.Int64
        detail.OriginalTransactionID = &id
    }
    if responseCode.Valid {
        detail.ResponseCode = &responseCode.String
    }
    if authCode.Valid {
        detail.AuthCode = &authCode.String
    }
    
    if detail.Type == "SALE" {
        productsQ := `
            SELECT 
                id,
                product_id,
                unit_price,
                quantity,
                product_name,
                product_sku
            FROM transaction.transaction_products
            WHERE transaction_id = $1
            ORDER BY id
        `
        
        rows, err := r.db.QueryContext(ctx, productsQ, txID)
        if err != nil {
            return nil, fmt.Errorf("failed to query products: %w", err)
        }
        defer rows.Close()
        
        for rows.Next() {
            var p models.TransactionProductDetail
            var name, sku sql.NullString
            
            if err := rows.Scan(
                &p.ID,
                &p.ProductID,
                &p.UnitPrice,
                &p.Quantity,
                &name,
                &sku,
            ); err != nil {
                return nil, fmt.Errorf("failed to scan product: %w", err)
            }
            
            if name.Valid {
                p.ProductName = &name.String
            }
            if sku.Valid {
                p.ProductSKU = &sku.String
            }
            
            p.LineTotal = p.UnitPrice * int64(p.Quantity)
            detail.Products = append(detail.Products, p)
        }
        
        if err := rows.Err(); err != nil {
            return nil, fmt.Errorf("iteration error: %w", err)
        }
    }
    
    return &detail, nil
}

func (r *transactionRepo) GetFilteredUserTransactions(ctx context.Context, filter *models.TransactionFilter) ([]*models.TransactionListItem, int64, error) {
	userUUID, err := uuid.Parse(*filter.UserID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid user id: %w", err)
	}
	
	var conditions []string
	var args []interface{}
	argCounter := 1
	
	conditions = append(conditions, fmt.Sprintf("t.user_id = $%d", argCounter))
	args = append(args, userUUID)
	argCounter++
	
	if len(filter.Types) > 0 {
		placeholders := make([]string, len(filter.Types))
		for i, txType := range filter.Types {
			placeholders[i] = fmt.Sprintf("$%d", argCounter)
			args = append(args, txType)
			argCounter++
		}
		conditions = append(conditions, fmt.Sprintf("t.type IN (%s)", strings.Join(placeholders, ",")))
	} else {
		conditions = append(conditions, "t.type IN ('SALE', 'VOID')")
	}
	
	if filter.Search != nil && strings.TrimSpace(*filter.Search) != "" {
		searchPattern := "%" + strings.TrimSpace(*filter.Search) + "%"
		conditions = append(conditions, fmt.Sprintf(`(
			CAST(t.id AS TEXT) LIKE $%d OR
			EXISTS (
				SELECT 1 FROM transaction.transaction_products tp 
				WHERE tp.transaction_id = t.id 
				AND tp.product_name ILIKE $%d
			)
		)`, argCounter, argCounter))
		args = append(args, searchPattern)
		argCounter++
	}
	
	dateFrom, dateTo := filter.GetParsedDates()
	
	if dateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("t.created_at >= $%d", argCounter))
		args = append(args, *dateFrom)
		argCounter++
	}
	
	if dateTo != nil {
		conditions = append(conditions, fmt.Sprintf("t.created_at <= $%d", argCounter))
		args = append(args, *dateTo)
		argCounter++
	}
	
	if len(filter.Statuses) > 0 {
		placeholders := make([]string, len(filter.Statuses))
		for i, status := range filter.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", argCounter)
			args = append(args, status)
			argCounter++
		}
		conditions = append(conditions, fmt.Sprintf("t.status IN (%s)", strings.Join(placeholders, ",")))
	}
	
	if filter.AmountMin != nil {
		conditions = append(conditions, fmt.Sprintf("t.amount >= $%d", argCounter))
		args = append(args, *filter.AmountMin)
		argCounter++
	}
	
	if filter.AmountMax != nil {
		conditions = append(conditions, fmt.Sprintf("t.amount <= $%d", argCounter))
		args = append(args, *filter.AmountMax)
		argCounter++
	}
	
	whereClause := strings.Join(conditions, " AND ")
	
	havingClause := ""
	var havingArgs []interface{}
	
	if filter.ProductCountMin != nil || filter.ProductCountMax != nil {
		var havingConditions []string
		
		if filter.ProductCountMin != nil {
			havingConditions = append(havingConditions, fmt.Sprintf("COUNT(tp.id) >= $%d", argCounter))
			havingArgs = append(havingArgs, *filter.ProductCountMin)
			argCounter++
		}
		
		if filter.ProductCountMax != nil {
			havingConditions = append(havingConditions, fmt.Sprintf("COUNT(tp.id) <= $%d", argCounter))
			havingArgs = append(havingArgs, *filter.ProductCountMax)
			argCounter++
		}
		
		if len(havingConditions) > 0 {
			havingClause = "HAVING " + strings.Join(havingConditions, " AND ")
		}
	}
	
	countArgs := append([]interface{}{}, args...)
	countArgs = append(countArgs, havingArgs...)
	
	countQ := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM (
			SELECT t.id
			FROM transaction.transactions t
			LEFT JOIN transaction.transaction_products tp ON t.id = tp.transaction_id
			WHERE %s
			GROUP BY t.id
			%s
		) AS filtered_transactions
	`, whereClause, havingClause)
	
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}
	
	dataArgs := append([]interface{}{}, args...)
	dataArgs = append(dataArgs, havingArgs...)
	
	offset := (filter.Page - 1) * filter.PageSize
	dataArgs = append(dataArgs, filter.PageSize, offset)
	
	dataQ := fmt.Sprintf(`
		SELECT 
			t.id,
			t.status,
			t.amount,
			t.currency_code,
			t.pan_masked,
			t.created_at,
			COUNT(tp.id) as product_count,
			COALESCE(array_agg(tp.product_id ORDER BY tp.id) FILTER (WHERE tp.product_id IS NOT NULL), '{}') as product_ids
		FROM transaction.transactions t
		LEFT JOIN transaction.transaction_products tp ON t.id = tp.transaction_id
		WHERE %s
		GROUP BY t.id, t.status, t.amount, t.currency_code, t.pan_masked, t.created_at
		%s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, havingClause, argCounter, argCounter+1)
	
	rows, err := r.db.QueryContext(ctx, dataQ, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query filtered transactions: %w", err)
	}
	defer rows.Close()
	
	var transactions []*models.TransactionListItem
	for rows.Next() {
		var tx models.TransactionListItem
		var productIds []int64
		
		if err := rows.Scan(
			&tx.ID,
			&tx.Status,
			&tx.Amount,
			&tx.CurrencyCode,
			&tx.PANMasked,
			&tx.CreatedAt,
			&tx.ProductCount,
			(*pq.Int64Array)(&productIds),
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction: %w", err)
		}
		
		tx.ProductIds = productIds
		transactions = append(transactions, &tx)
	}
	
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iteration error: %w", err)
	}
	
	return transactions, total, nil
}

func (r *transactionRepo) GetFilteredAdminTransactions(ctx context.Context, filter *models.AdminTransactionFilter) ([]*models.AdminTransactionListItem, int64, error) {
	var conditions []string
	var args []interface{}
	argCounter := 1
	
	if len(filter.Types) > 0 {
		placeholders := make([]string, len(filter.Types))
		for i, txType := range filter.Types {
			placeholders[i] = fmt.Sprintf("$%d", argCounter)
			args = append(args, txType)
			argCounter++
		}
		conditions = append(conditions, fmt.Sprintf("t.type IN (%s)", strings.Join(placeholders, ",")))
	} else {
		conditions = append(conditions, "t.type IN ('SALE', 'VOID')")
	}
	
	if filter.Search != nil && strings.TrimSpace(*filter.Search) != "" {
		searchPattern := "%" + strings.TrimSpace(*filter.Search) + "%"
		conditions = append(conditions, fmt.Sprintf(`(
			CAST(t.id AS TEXT) LIKE $%d OR
			CAST(t.user_id AS TEXT) LIKE $%d OR
			EXISTS (
				SELECT 1 FROM transaction.transaction_products tp 
				WHERE tp.transaction_id = t.id 
				AND tp.product_name ILIKE $%d
			)
		)`, argCounter, argCounter, argCounter))
		args = append(args, searchPattern)
		argCounter++
	}
	
	if filter.UserID != nil && strings.TrimSpace(*filter.UserID) != "" {
		userUUID, err := uuid.Parse(*filter.UserID)
		if err == nil {
			conditions = append(conditions, fmt.Sprintf("t.user_id = $%d", argCounter))
			args = append(args, userUUID)
			argCounter++
		}
	}
	
	dateFrom, dateTo := filter.GetParsedDates()
	
	if dateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("t.created_at >= $%d", argCounter))
		args = append(args, *dateFrom)
		argCounter++
	}
	
	if dateTo != nil {
		conditions = append(conditions, fmt.Sprintf("t.created_at <= $%d", argCounter))
		args = append(args, *dateTo)
		argCounter++
	}
	
	if len(filter.Statuses) > 0 {
		placeholders := make([]string, len(filter.Statuses))
		for i, status := range filter.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", argCounter)
			args = append(args, status)
			argCounter++
		}
		conditions = append(conditions, fmt.Sprintf("t.status IN (%s)", strings.Join(placeholders, ",")))
	}
	
	if filter.AmountMin != nil {
		conditions = append(conditions, fmt.Sprintf("t.amount >= $%d", argCounter))
		args = append(args, *filter.AmountMin)
		argCounter++
	}
	
	if filter.AmountMax != nil {
		conditions = append(conditions, fmt.Sprintf("t.amount <= $%d", argCounter))
		args = append(args, *filter.AmountMax)
		argCounter++
	}
	
	whereClause := strings.Join(conditions, " AND ")
	
	havingClause := ""
	if filter.ProductCountMin != nil || filter.ProductCountMax != nil {
		var havingConditions []string
		
		if filter.ProductCountMin != nil {
			havingConditions = append(havingConditions, fmt.Sprintf("COUNT(tp.id) >= $%d", argCounter))
			args = append(args, *filter.ProductCountMin)
			argCounter++
		}
		
		if filter.ProductCountMax != nil {
			havingConditions = append(havingConditions, fmt.Sprintf("COUNT(tp.id) <= $%d", argCounter))
			args = append(args, *filter.ProductCountMax)
			argCounter++
		}
		
		if len(havingConditions) > 0 {
			havingClause = "HAVING " + strings.Join(havingConditions, " AND ")
		}
	}
	
	countQ := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM (
			SELECT t.id
			FROM transaction.transactions t
			LEFT JOIN transaction.transaction_products tp ON t.id = tp.transaction_id
			WHERE %s
			GROUP BY t.id
			%s
		) AS filtered_transactions
	`, whereClause, havingClause)
	
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count admin transactions: %w", err)
	}
	
	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)
	
	dataQ := fmt.Sprintf(`
		SELECT 
			t.id,
			t.user_id,
			t.type,
			t.status,
			t.amount,
			t.currency_code,
			t.pan_masked,
			t.response_code,
			t.auth_code,
			t.original_transaction_id,
			t.created_at,
			COUNT(tp.id) as product_count,
			COALESCE(array_agg(tp.product_id ORDER BY tp.id) FILTER (WHERE tp.product_id IS NOT NULL), '{}') as product_ids
		FROM transaction.transactions t
		LEFT JOIN transaction.transaction_products tp ON t.id = tp.transaction_id
		WHERE %s
		GROUP BY t.id, t.user_id, t.type, t.status, t.amount, t.currency_code, 
				 t.pan_masked, t.response_code, t.auth_code, t.original_transaction_id, t.created_at
		%s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, havingClause, argCounter, argCounter+1)
	
	rows, err := r.db.QueryContext(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query admin transactions: %w", err)
	}
	defer rows.Close()
	
	var transactions []*models.AdminTransactionListItem
	for rows.Next() {
		var tx models.AdminTransactionListItem
		var responseCode, authCode sql.NullString
		var originalTxID sql.NullInt64
		var productIds []int64
		
		if err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.Type,
			&tx.Status,
			&tx.Amount,
			&tx.CurrencyCode,
			&tx.PANMasked,
			&responseCode,
			&authCode,
			&originalTxID,
			&tx.CreatedAt,
			&tx.ProductCount,
			(*pq.Int64Array)(&productIds),
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan admin transaction: %w", err)
		}
		
		if responseCode.Valid {
			tx.ResponseCode = &responseCode.String
		}
		if authCode.Valid {
			tx.AuthCode = &authCode.String
		}
		if originalTxID.Valid {
			id := originalTxID.Int64
			tx.OriginalTransactionID = &id
		}
		
		tx.ProductIds = productIds
		transactions = append(transactions, &tx)
	}
	
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iteration error: %w", err)
	}
	
	return transactions, total, nil
}

