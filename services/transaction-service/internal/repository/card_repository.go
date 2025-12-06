package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/transaction-service/pkg/models"
)

type BankCardRepository interface {
	Create(ctx context.Context, userID uuid.UUID, card *models.BankCard, isPrimary bool) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.BankCard, error)
	FindByID(ctx context.Context, userID uuid.UUID, cardID int) (*models.BankCard, error)
	SetPrimary(ctx context.Context, userID uuid.UUID, cardID int) error
	Delete(ctx context.Context, userID uuid.UUID, cardID int) error
	UserHasCard(ctx context.Context, userID uuid.UUID, cardID int) (bool, error)
	UpdateCard(ctx context.Context, card *models.BankCard) error
}

type bankCardRepo struct {
	db *sql.DB
}

func NewBankCardRepository(db *sql.DB) BankCardRepository {
	return &bankCardRepo{db: db}
}

func (r *bankCardRepo) Create(ctx context.Context, userID uuid.UUID, card *models.BankCard, isPrimary bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if isPrimary {
		if err := r.unsetPrimaryCards(ctx, tx, userID); err != nil {
			return fmt.Errorf("failed to unset primary cards: %w", err)
		}
	}

	q := `
        INSERT INTO transaction.bank_cards (
            pan_encrypted,
            iv,
            last_four_digits,
            card_brand,
            expiration_month,
            expiration_year,
            fullname_on_card,
            card_name
        )
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
        RETURNING id, created_at, updated_at
    `

	err = tx.QueryRowContext(
		ctx, q,
		card.PANEncrypted,
		card.IV,
		card.LastFourDigits,
		card.CardBrand,
		card.ExpirationMonth,
		card.ExpirationYear,
		card.FullnameOnCard,
		card.CardName,
	).Scan(&card.ID, &card.CreatedAt, &card.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert card: %w", err)
	}

	linkQ := `
        INSERT INTO transaction.user_bank_cards (user_id, card_id, is_primary)
        VALUES ($1, $2, $3)
    `

	_, err = tx.ExecContext(ctx, linkQ, userID, card.ID, isPrimary)
	if err != nil {
		return fmt.Errorf("failed to link card to user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *bankCardRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.BankCard, error) {
	q := `
        SELECT 
            bc.id,
            bc.last_four_digits,
            bc.card_brand,
            bc.expiration_month,
            bc.expiration_year,
            bc.fullname_on_card,
            bc.card_name,
            bc.created_at,
            bc.updated_at,
            ubc.is_primary
        FROM transaction.bank_cards bc
        INNER JOIN transaction.user_bank_cards ubc ON bc.id = ubc.card_id
        WHERE ubc.user_id = $1
        ORDER BY ubc.is_primary DESC, bc.created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards: %w", err)
	}
	defer rows.Close()

	var cards []*models.BankCard

	for rows.Next() {
		var card models.BankCard

		err := rows.Scan(
			&card.ID,
			&card.LastFourDigits,
			&card.CardBrand,
			&card.ExpirationMonth,
			&card.ExpirationYear,
			&card.FullnameOnCard,
			&card.CardName,
			&card.CreatedAt,
			&card.UpdatedAt,
			&card.IsPrimary,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		cards = append(cards, &card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration error: %w", err)
	}

	return cards, nil
}

func (r *bankCardRepo) FindByID(ctx context.Context, userID uuid.UUID, cardID int) (*models.BankCard, error) {
	q := `
        SELECT 
            bc.id,
            bc.last_four_digits,
            bc.card_brand,
            bc.expiration_month,
            bc.expiration_year,
            bc.fullname_on_card,
            bc.card_name,
            bc.created_at,
            bc.updated_at,
            ubc.is_primary
        FROM transaction.bank_cards bc
        INNER JOIN transaction.user_bank_cards ubc ON bc.id = ubc.card_id
        WHERE ubc.user_id = $1 AND bc.id = $2
        LIMIT 1
    `

	var card models.BankCard

	err := r.db.QueryRowContext(ctx, q, userID, cardID).Scan(
		&card.ID,
		&card.LastFourDigits,
		&card.CardBrand,
		&card.ExpirationMonth,
		&card.ExpirationYear,
		&card.FullnameOnCard,
		&card.CardName,
		&card.CreatedAt,
		&card.UpdatedAt,
		&card.IsPrimary,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &card, nil
}

func (r *bankCardRepo) SetPrimary(ctx context.Context, userID uuid.UUID, cardID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var exists bool
	checkQ := `SELECT EXISTS(
        SELECT 1 FROM transaction.user_bank_cards 
        WHERE user_id = $1 AND card_id = $2
    )`
	err = tx.QueryRowContext(ctx, checkQ, userID, cardID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to verify card ownership: %w", err)
	}
	if !exists {
		return fmt.Errorf("card not found or does not belong to user")
	}

	if err := r.unsetPrimaryCards(ctx, tx, userID); err != nil {
		return fmt.Errorf("failed to unset primary cards: %w", err)
	}

	updateQ := `
        UPDATE transaction.user_bank_cards 
        SET is_primary = true 
        WHERE user_id = $1 AND card_id = $2
    `
	_, err = tx.ExecContext(ctx, updateQ, userID, cardID)
	if err != nil {
		return fmt.Errorf("failed to set primary card: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *bankCardRepo) Delete(ctx context.Context, userID uuid.UUID, cardID int) error {
	q := `
        DELETE FROM transaction.user_bank_cards 
        WHERE user_id = $1 AND card_id = $2
    `

	result, err := r.db.ExecContext(ctx, q, userID, cardID)
	if err != nil {
		return fmt.Errorf("failed to delete card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("card not found or does not belong to user")
	}

	return nil
}

func (r *bankCardRepo) UserHasCard(ctx context.Context, userID uuid.UUID, cardID int) (bool, error) {
	var exists bool
	q := `SELECT EXISTS(
        SELECT 1 FROM transaction.user_bank_cards 
        WHERE user_id = $1 AND card_id = $2
    )`

	err := r.db.QueryRowContext(ctx, q, userID, cardID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to verify card ownership: %w", err)
	}

	return exists, nil
}

func (r *bankCardRepo) UpdateCard(ctx context.Context, card *models.BankCard) error {
	q := `
		UPDATE transaction.bank_cards 
		SET card_name = COALESCE($1, card_name), pan_encrypted = COALESCE($2, pan_encrypted), iv = COALESCE($3, iv),
		    last_four_digits = COALESCE($4, last_four_digits), card_brand = COALESCE($5, card_brand),
		    expiration_month = COALESCE($6, expiration_month), expiration_year = COALESCE($7, expiration_year),
		    fullname_on_card = COALESCE($8, fullname_on_card), updated_at = NOW() 
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, q, card.CardName, card.PANEncrypted, card.IV, card.LastFourDigits, card.CardBrand,
		card.ExpirationMonth, card.ExpirationYear, card.FullnameOnCard, card.ID)
	if err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}

	return nil
}

func (r *bankCardRepo) unsetPrimaryCards(ctx context.Context, tx *sql.Tx, userID uuid.UUID) error {
	q := `
        UPDATE transaction.user_bank_cards 
        SET is_primary = false 
        WHERE user_id = $1
    `
	_, err := tx.ExecContext(ctx, q, userID)
	return err
}
