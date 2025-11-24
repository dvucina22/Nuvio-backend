package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/catalog-service/pkg/models"
	"github.com/lib/pq"
)

type ProductRepository interface {
	GetFilteredProducts(filter *models.ProductFilter) ([]models.Product, error)
	ExistsByID(ctx context.Context, productID int) (bool, error)
}

type productRepo struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) GetFilteredProducts(filter *models.ProductFilter) ([]models.Product, error) {

	query := `
		SELECT 
			p.id, p.name, p.description, p.model_number, p.sku, p.base_price, p.is_active,
			p.created_at, p.updated_at,

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

	products := make(map[int64]*models.Product)

	for rows.Next() {
		var p models.Product
		var brand models.Brand
		var category models.Category

		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.ModelNumber, &p.SKU, &p.BasePrice, &p.IsActive,
			&p.CreatedAt, &p.UpdatedAt,
			&brand.ID, &brand.Name,
			&category.ID, &category.Name, &category.ParentID,
		)
		if err != nil {
			return nil, err
		}

		p.Brand = brand
		p.Category = category

		products[p.ID] = &p
	}

	if len(products) == 0 {
		return []models.Product{}, nil
	}

	productIDs := make([]int64, 0, len(products))
	for id := range products {
		productIDs = append(productIDs, id)
	}

	imageQuery := `
		SELECT id, product_id, url, is_primary
		FROM catalog.product_images
		WHERE product_id = ANY($1)
	`

	imgRows, err := r.db.Query(imageQuery, pq.Array(productIDs))
	if err != nil {
		return nil, err
	}
	defer imgRows.Close()

	for imgRows.Next() {
		var img models.ProductImage
		var productID int64

		err := imgRows.Scan(
			&img.ID, &productID, &img.URL, &img.IsPrimary,
		)
		if err != nil {
			return nil, err
		}

		products[productID].Images = append(products[productID].Images, img)
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
		var pa models.ProductAttribute
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

		products[productID].Attributes = append(products[productID].Attributes, pa)
	}

	result := make([]models.Product, 0, len(products))
	for _, p := range products {
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
