package repository

import (
	"database/sql"

	"github.com/catalog-service/pkg/models/cart"
)

type CartRepository interface {
	AddProductToCart(userID string, productID int, quantity int) error
	RemoveProductFromCart(userID string, productID int) error
	GetCartContents(userID string) (map[int]int, error)
	CartExists(userID string) (bool, error)
	GetCartProducts(userID string) ([]cart.CartProduct, error)
}

type cartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) AddProductToCart(userID string, productID int, quantity int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	var cartID int
	err = tx.QueryRow(`
		SELECT id FROM catalog.carts WHERE user_id = $1
	`, userID).Scan(&cartID)

	if err == sql.ErrNoRows {
		err = tx.QueryRow(`
			INSERT INTO catalog.carts (user_id)
			VALUES ($1)
			RETURNING id
		`, userID).Scan(&cartID)
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO catalog.cart_items (cart_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, product_id)
		DO UPDATE SET quantity = catalog.cart_items.quantity + EXCLUDED.quantity
	`, cartID, productID, quantity)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *cartRepository) RemoveProductFromCart(userID string, productID int) error {
	query := `
		DELETE FROM catalog.cart_items 
		WHERE product_id = $1 
		AND cart_id = (SELECT id FROM catalog.carts WHERE user_id = $2)
	`
	_, err := r.db.Exec(query, productID, userID)
	return err
}

func (r *cartRepository) GetCartContents(userID string) (map[int]int, error) {
	query := `
		SELECT bi.product_id, bi.quantity
		FROM catalog.cart_items bi
		JOIN catalog.carts b ON b.id = bi.cart_id
		WHERE b.user_id = $1
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]int)
	for rows.Next() {
		var productID, qty int
		if err := rows.Scan(&productID, &qty); err != nil {
			return nil, err
		}
		result[productID] = qty
	}

	return result, nil
}

func (r *cartRepository) GetCartProducts(userID string) ([]cart.CartProduct, error) {
	query := `
        SELECT 
            p.id,
            p.name,
            p.base_price,
            b.name AS brand_name,
            bi.quantity,
            CASE WHEN uf.product_id IS NOT NULL THEN TRUE ELSE FALSE END AS is_favorite
        FROM catalog.cart_items bi
        JOIN catalog.carts ca ON ca.id = bi.cart_id
        JOIN catalog.products p ON p.id = bi.product_id
        JOIN catalog.brands b ON b.id = p.brand_id
        LEFT JOIN catalog.user_favorites uf 
            ON uf.product_id = p.id AND uf.user_id = $1
        WHERE ca.user_id = $1
        ORDER BY p.id
    `

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []cart.CartProduct{}

	for rows.Next() {
		var dto cart.CartProduct

		err := rows.Scan(
			&dto.ID,
			&dto.Name,
			&dto.BasePrice,
			&dto.BrandName,
			&dto.Quantity,
			&dto.IsFavorite,
		)
		if err != nil {
			return nil, err
		}

		dto.ImageURL, err = r.loadPrimaryImage(dto.ID)
		if err != nil {
			return nil, err
		}

		result = append(result, dto)
	}

	return result, nil
}

func (r *cartRepository) CartExists(userID string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM catalog.carts WHERE user_id = $1
		)
	`
	err := r.db.QueryRow(query, userID).Scan(&exists)
	return exists, err
}

func (r *cartRepository) loadPrimaryImage(productID int64) (string, error) {
	var url string

	err := r.db.QueryRow(`
        SELECT url 
        FROM catalog.product_images
        WHERE product_id = $1
        ORDER BY is_primary DESC, id ASC
        LIMIT 1
    `, productID).Scan(&url)

	if err == sql.ErrNoRows {
		return "", nil
	}

	return url, err
}
