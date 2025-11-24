package repository

import "database/sql"

type FavoritesRepository interface {
	AddToFavorites(userID string, productID int) error
	RemoveFromFavorites(userID string, productID int) error
	IsFavorited(userID string, productID int) (bool, error)
}

type favoritesRepo struct {
	db *sql.DB
}

func NewFavoritesRepository(db *sql.DB) FavoritesRepository {
	return &favoritesRepo{db: db}
}

func (r *favoritesRepo) AddToFavorites(userID string, productID int) error {
	query := `
		INSERT INTO catalog.user_favorites (user_id, product_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, product_id) DO NOTHING
	`
	_, err := r.db.Exec(query, userID, productID)
	return err
}

func (r *favoritesRepo) RemoveFromFavorites(userID string, productID int) error {
	query := `
		DELETE FROM catalog.user_favorites
		WHERE user_id = $1 AND product_id = $2
	`
	_, err := r.db.Exec(query, userID, productID)
	return err
}

func (r *favoritesRepo) IsFavorited(userID string, productID int) (bool, error) {
	query := `
		SELECT 1 FROM catalog.user_favorites	
		WHERE user_id = $1 AND product_id = $2
	`
	var exists int
	err := r.db.QueryRow(query, userID, productID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
