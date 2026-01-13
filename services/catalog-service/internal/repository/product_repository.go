package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/catalog-service/pkg/models/products"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ProductRepository interface {
	GetFilteredProducts(filter *products.ProductFilter) ([]products.ProductMinimal, error)
	ExistsByID(ctx context.Context, productID int) (bool, error)
	GetProductByID(userID *uuid.UUID, productID int) (*products.Product, error)
	DeleteProductByID(productID int) error
	UpdateProductByID(productID int, product *products.UpdateProduct) error
	CreateProduct(product *products.CreateProduct) error
	GetPrimaryImages(ctx context.Context, productIds []int64) (map[int64]string, error)
}
type productRepo struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) GetFilteredProducts(filter *products.ProductFilter) ([]products.ProductMinimal, error) {

	query := `
		SELECT 
			p.id, p.name, p.description, p.model_number, p.sku, p.base_price, p.is_active,
			p.created_at, p.updated_at, p.quantity,

			b.id, b.name,
			c.id, c.name, c.parent_id
		FROM catalog.products p
		LEFT JOIN catalog.brands b ON b.id = p.brand_id
		LEFT JOIN catalog.categories c ON c.id = p.category_id
	`

	var where []string
	var args []any
	arg := 1

	if filter.IsFavorite != nil && *filter.IsFavorite {
		if filter.IsFavoriteUserID == nil {
			return nil, fmt.Errorf("favorite filter requires IsFavoriteUserID")
		}

		query += `
			INNER JOIN catalog.user_favorites uf 
				ON uf.product_id = p.id AND uf.user_id = $` + fmt.Sprint(arg)

		args = append(args, *filter.IsFavoriteUserID)
		arg++
	}

	if filter.Search != nil && *filter.Search != "" {
		where = append(where,
			fmt.Sprintf(`(
				p.name ILIKE '%%' || $%d || '%%' OR 
				p.model_number ILIKE '%%' || $%d || '%%' OR 
				p.sku ILIKE '%%' || $%d || '%%'
			)`, arg, arg+1, arg+2),
		)
		args = append(args, *filter.Search, *filter.Search, *filter.Search)
		arg += 3
	}

	if len(filter.BrandIDs) > 0 {
		where = append(where, fmt.Sprintf("p.brand_id = ANY($%d)", arg))
		args = append(args, pq.Array(filter.BrandIDs))
		arg++
	}

	if len(filter.CategoryIDs) > 0 {
		where = append(where, fmt.Sprintf("p.category_id = ANY($%d)", arg))
		args = append(args, pq.Array(filter.CategoryIDs))
		arg++
	}

	if filter.IsActive != nil {
		where = append(where, fmt.Sprintf("p.is_active = $%d", arg))
		args = append(args, *filter.IsActive)
		arg++
	}

	if filter.IsInStock != nil {
		if *filter.IsInStock {
			where = append(where, "p.quantity > 0")
		} else {
			where = append(where, "p.quantity = 0")
		}
	}

	if len(filter.Attributes) > 0 {
		for _, attr := range filter.Attributes {
			subQuery := fmt.Sprintf(`
				EXISTS (
					SELECT 1 
					FROM catalog.product_attributes pa
					JOIN catalog.attributes a ON a.id = pa.attribute_id
					WHERE pa.product_id = p.id
					AND a.name = $%d
					AND a.value IN (SELECT unnest($%d::varchar[]))
				)`, arg, arg+1)

			where = append(where, subQuery)
			args = append(args, attr.Name, pq.Array(attr.Values))
			arg += 2
		}
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	switch filter.Sort {
	case "price_asc":
		query += " ORDER BY p.base_price ASC"
	case "price_desc":
		query += " ORDER BY p.base_price DESC"
	case "newest":
		query += " ORDER BY p.created_at DESC"
	default:
		query += " ORDER BY p.id"
	}

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productsList := make(map[int64]*products.ProductMinimal)

	for rows.Next() {
		var p products.ProductMinimal
		var brand products.Brand
		var category products.Category

		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.ModelNumber, &p.SKU, &p.BasePrice, &p.IsActive,
			&p.CreatedAt, &p.UpdatedAt, &p.Quantity,
			&brand.ID, &brand.Name,
			&category.ID, &category.Name, &category.ParentID,
		)
		if err != nil {
			return nil, err
		}

		p.BrandName = brand.Name
		p.CategoryName = category.Name

		productsList[p.ID] = &p
	}

	if len(productsList) == 0 {
		return []products.ProductMinimal{}, nil
	}

	productIDs := make([]int64, 0, len(productsList))
	for id := range productsList {
		productIDs = append(productIDs, id)
	}

	imageQuery := `
		SELECT pi.product_id, pi.url
		FROM catalog.product_images pi
		WHERE pi.product_id = ANY($1)
		AND pi.is_primary = TRUE
	`

	imageRows, err := r.db.Query(imageQuery, pq.Array(productIDs))
	if err != nil {
		return nil, err
	}
	defer imageRows.Close()

	for imageRows.Next() {
		var productID int64
		var imageURL string

		err := imageRows.Scan(&productID, &imageURL)
		if err != nil {
			return nil, err
		}

		productsList[productID].ImageURL = imageURL
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
		var pa products.ProductAttribute
		var productID int64

		err := attrRows.Scan(
			&productID,
			&pa.ID,
			&pa.Name,
			&pa.Value,
		)
		if err != nil {
			return nil, err
		}

		productsList[productID].Attributes = append(productsList[productID].Attributes, pa)
	}

	result := make([]products.ProductMinimal, 0, len(productsList))
	for _, p := range productsList {
		result = append(result, *p)
	}

	return result, nil
}

func (r *productRepo) ExistsByID(ctx context.Context, productID int) (bool, error) {
	const q = `SELECT 1 FROM catalog.products WHERE id = $1`

	var exists int
	err := r.db.QueryRowContext(ctx, q, productID).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *productRepo) GetProductByID(userID *uuid.UUID, productID int) (*products.Product, error) {
	const productQuery = `
        SELECT 
            p.id, p.name, p.description, p.model_number, p.sku, p.base_price, p.is_active,
            p.created_at, p.updated_at, p.quantity,
            b.id, b.name,
            c.id, c.name, c.parent_id
        FROM catalog.products p
        LEFT JOIN catalog.brands b ON b.id = p.brand_id
        LEFT JOIN catalog.categories c ON c.id = p.category_id
        WHERE p.id = $1
    `

	var p products.Product
	var brand products.Brand
	var category products.Category

	err := r.db.QueryRow(productQuery, productID).Scan(
		&p.ID, &p.Name, &p.Description, &p.ModelNumber, &p.SKU, &p.BasePrice, &p.IsActive,
		&p.CreatedAt, &p.UpdatedAt, &p.Quantity,
		&brand.ID, &brand.Name,
		&category.ID, &category.Name, &category.ParentID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	p.Brand = brand
	p.Category = category

	const imagesQuery = `
        SELECT id, url, is_primary
        FROM catalog.product_images
        WHERE product_id = $1
        ORDER BY is_primary DESC, id ASC
    `
	imageRows, err := r.db.Query(imagesQuery, p.ID)
	if err != nil {
		return nil, err
	}
	defer imageRows.Close()

	for imageRows.Next() {
		var img products.ProductImage
		if err := imageRows.Scan(&img.ID, &img.URL, &img.IsPrimary); err != nil {
			return nil, err
		}
		p.Images = append(p.Images, img)
	}

	const attrsQuery = `
        SELECT pa.product_id, a.id, a.name, a.value
        FROM catalog.product_attributes pa
        JOIN catalog.attributes a ON a.id = pa.attribute_id
        WHERE pa.product_id = $1
    `
	attrRows, err := r.db.Query(attrsQuery, p.ID)
	if err != nil {
		return nil, err
	}
	defer attrRows.Close()

	for attrRows.Next() {
		var pid int64
		var attr products.ProductAttribute
		if err := attrRows.Scan(&pid, &attr.ID, &attr.Name, &attr.Value); err != nil {
			return nil, err
		}
		p.Attributes = append(p.Attributes, attr)
	}

	if userID == nil {
		p.IsFavorite = false
		return &p, nil
	}

	const favQuery = `
    SELECT EXISTS(
        SELECT 1 
        FROM catalog.user_favorites 
        WHERE user_id = $1 AND product_id = $2
    )`

	var isFavorite bool
	err = r.db.QueryRow(favQuery, *userID, p.ID).Scan(&isFavorite)
	if err != nil {
		return nil, err
	}

	p.IsFavorite = isFavorite
	return &p, nil
}

func (r *productRepo) DeleteProductByID(productID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		DELETE FROM catalog.products
		WHERE id = $1
	`, productID)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *productRepo) UpdateProductByID(productID int, product *products.UpdateProduct) error {
	const updateProductQ = `
		UPDATE catalog.products SET
			name         = COALESCE($1, name),
			description  = COALESCE($2, description),
			model_number = COALESCE($3, model_number),
			sku          = COALESCE($4, sku),
			base_price   = COALESCE($5, base_price),
			is_active    = COALESCE($6, is_active),
			brand_id     = COALESCE($7, brand_id),
			category_id  = COALESCE($8, category_id),
			quantity     = COALESCE($9, quantity),
			updated_at   = NOW()
		WHERE id = $10
	`

	const deleteAttributesQ = `
		DELETE FROM catalog.product_attributes
		WHERE product_id = $1
	`

	const insertAttributeQ = `
		INSERT INTO catalog.product_attributes (product_id, attribute_id)
		VALUES ($1, $2)
	`

	const deleteImagesQ = `
		DELETE FROM catalog.product_images
		WHERE product_id = $1
	`

	const insertImageQ = `
		INSERT INTO catalog.product_images (product_id, url, is_primary, created_at)
		VALUES ($1, $2, $3, NOW())
	`

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	if _, err = tx.Exec(
		updateProductQ,
		toNullString(product.Name),
		toNullString(product.Description),
		toNullString(product.ModelNumber),
		toNullString(product.SKU),
		toNullFloat64(product.BasePrice),
		toNullBool(product.IsActive),
		toNullInt64(product.BrandID),
		toNullInt64(product.CategoryID),
		toNullInt64(product.Quantity),
		productID,
	); err != nil {
		_ = tx.Rollback()
		return err
	}

	if product.AttributeIds != nil {
		if _, err = tx.Exec(deleteAttributesQ, productID); err != nil {
			_ = tx.Rollback()
			return err
		}

		for _, attrID := range *product.AttributeIds {
			if _, err = tx.Exec(insertAttributeQ, productID, attrID); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}

	if product.ImageURLs != nil {
		if _, err = tx.Exec(deleteImagesQ, productID); err != nil {
			_ = tx.Rollback()
			return err
		}

		for i, url := range *product.ImageURLs {
			isPrimary := (i == 0)
			if _, err = tx.Exec(insertImageQ, productID, url, isPrimary); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *productRepo) CreateProduct(product *products.CreateProduct) error {
	const insertProduct = `
		INSERT INTO catalog.products (
			name,
			description,
			model_number,
			sku,
			base_price,
			is_active,
			brand_id,
			category_id,
			quantity,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, TRUE, $6, $7, $8, NOW(), NOW())
		RETURNING id
	`

	const insertProductAttribute = `
		INSERT INTO catalog.product_attributes (product_id, attribute_id)
		VALUES ($1, $2)
	`

	const insertProductImage = `
		INSERT INTO catalog.product_images (product_id, url, is_primary)
		VALUES ($1, $2, $3)
	`

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	var productID int64

	err = tx.QueryRow(
		insertProduct,
		product.Name,
		product.Description,
		product.ModelNumber,
		product.SKU,
		product.BasePrice,
		product.BrandID,
		product.CategoryID,
		product.Quantity,
	).Scan(&productID)
	if err != nil {
		return err
	}

	if len(product.AttributeIds) > 0 {
		for _, attrID := range product.AttributeIds {
			if _, err = tx.Exec(insertProductAttribute, productID, attrID); err != nil {
				return err
			}
		}
	}

	if len(product.ImageURLs) > 0 {
		for i, url := range product.ImageURLs {
			isPrimary := (i == 0)
			if _, err = tx.Exec(insertProductImage, productID, url, isPrimary); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *productRepo) GetPrimaryImages(ctx context.Context, productIds []int64) (map[int64]string, error) {
	if len(productIds) == 0 {
		return make(map[int64]string), nil
	}
	
	query := `
		SELECT pi.product_id, pi.url
		FROM catalog.product_images pi
		WHERE pi.product_id = ANY($1)
		AND pi.is_primary = TRUE
	`
	
	rows, err := r.db.QueryContext(ctx, query, pq.Array(productIds))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	result := make(map[int64]string)
	
	for rows.Next() {
		var productID int64
		var imageURL string
		
		if err := rows.Scan(&productID, &imageURL); err != nil {
			return nil, err
		}
		
		result[productID] = imageURL
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return result, nil
}

func toNullString(v *string) any {
	if v == nil {
		return nil
	}
	return *v
}

func toNullFloat64(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}

func toNullBool(v *bool) any {
	if v == nil {
		return nil
	}
	return *v
}

func toNullInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}
