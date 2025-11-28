package repository

import (
	"database/sql"

	"github.com/catalog-service/pkg/models/cart"
	"github.com/catalog-service/pkg/models/products"
	"github.com/lib/pq"
)

type CartRepository interface {
	AddProductToCart(userID string, productID int) error
	RemoveProductFromCart(userID string, productID int) error
	GetCartContents(userID string) (map[int]int, error)
	CartExists(userID string) (bool, error)
	GetCartProducts(userID string) ([]cart.CartProduct, error)
	EmptyCart(userID string) error
}

type cartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) AddProductToCart(userID string, productID int) error {
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
		VALUES ($1, $2, 1)
		ON CONFLICT (cart_id, product_id)
		DO UPDATE SET quantity = catalog.cart_items.quantity + 1
	`, cartID, productID)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *cartRepository) RemoveProductFromCart(userID string, productID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	var cartID int
	err = tx.QueryRow(`
		SELECT id FROM catalog.carts WHERE user_id = $1
	`, userID).Scan(&cartID)

	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	var quantity int
	err = tx.QueryRow(`
		SELECT quantity 
		FROM catalog.cart_items 
		WHERE cart_id = $1 AND product_id = $2
	`, cartID, productID).Scan(&quantity)

	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	if quantity > 1 {
		_, err = tx.Exec(`
			UPDATE catalog.cart_items
			SET quantity = quantity - 1
			WHERE cart_id = $1 AND product_id = $2
		`, cartID, productID)
	} else {
		_, err = tx.Exec(`
			DELETE FROM catalog.cart_items
			WHERE cart_id = $1 AND product_id = $2
		`, cartID, productID)
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
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
            COALESCE(c.name, '') AS category_name,
            bi.quantity,
            CASE WHEN uf.product_id IS NOT NULL THEN TRUE ELSE FALSE END AS is_favorite
        FROM catalog.cart_items bi
        JOIN catalog.carts ca ON ca.id = bi.cart_id
        JOIN catalog.products p ON p.id = bi.product_id
        JOIN catalog.brands b ON b.id = p.brand_id
        LEFT JOIN catalog.categories c ON c.id = p.category_id
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
	productIndex := make(map[int64]int)

	for rows.Next() {
		var dto cart.CartProduct

		err := rows.Scan(
			&dto.ID,
			&dto.Name,
			&dto.BasePrice,
			&dto.BrandName,
			&dto.CategoryName,
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

		dto.Attributes = []products.ProductAttribute{}

		productIndex[dto.ID] = len(result)
		result = append(result, dto)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return []cart.CartProduct{}, nil
	}

	productIDs := make([]int64, 0, len(productIndex))
	for id := range productIndex {
		productIDs = append(productIDs, id)
	}

	attrQuery := `
        SELECT pa.product_id, a.id, a.name, a.value
        FROM catalog.product_attributes pa
        JOIN catalog.attributes a ON a.id = pa.attribute_id
        WHERE pa.product_id = ANY($1)
    `

	attrRows, err := r.db.Query(attrQuery, pq.Array(productIDs))
	if err != nil {
		return nil, err
	}
	defer attrRows.Close()

	for attrRows.Next() {
		var (
			productID int64
			attrID    int64
			attrName  string
			attrValue string
		)

		if err := attrRows.Scan(&productID, &attrID, &attrName, &attrValue); err != nil {
			return nil, err
		}

		if idx, ok := productIndex[productID]; ok {
			result[idx].Attributes = append(result[idx].Attributes, products.ProductAttribute{
				ID:    attrID,
				Name:  attrName,
				Value: attrValue,
			})
		}
	}

	if err := attrRows.Err(); err != nil {
		return nil, err
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

func (r *cartRepository) EmptyCart(userID string) error {
	query := `
		DELETE FROM catalog.cart_items 
		WHERE cart_id = (SELECT id FROM catalog.carts WHERE user_id = $1)
	`
	_, err := r.db.Exec(query, userID)
	return err
}
