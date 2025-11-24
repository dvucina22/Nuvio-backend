package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/catalog-service/pkg/models"
	"github.com/lib/pq"
)

type ProductRepository interface {
	GetFilteredProducts(filter *models.ProductFilter) ([]models.Product, error)
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

	if filter.Search != nil && *filter.Search != "" {
		where = append(where,
			fmt.Sprintf(`(
				p.name ILIKE '%%' || $%d || '%%' OR 
				p.model_number ILIKE '%%' || $%d || '%%' OR 
				p.sku ILIKE '%%' || $%d || '%%'
			)`, arg, arg, arg),
		)
		args = append(args, *filter.Search)
		arg++
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

	if len(filter.AttributeFilters) > 0 {
		for _, attrFilter := range filter.AttributeFilters {
			subQuery := fmt.Sprintf(`
				EXISTS (
					SELECT 1 
					FROM catalog.product_attribute_values pav
					WHERE pav.product_id = p.id
					AND pav.attribute_id = $%d
					AND pav.value_id = ANY($%d)
				)`, arg, arg+1)
			where = append(where, subQuery)
			args = append(args, attrFilter.AttributeID, pq.Array(attrFilter.ValueIDs))
			arg += 2
		}
	}

	if len(filter.VariantAttributeFilters) > 0 {
		variantConditions := make([]string, 0, len(filter.VariantAttributeFilters))
		for _, vaf := range filter.VariantAttributeFilters {
			variantConditions = append(variantConditions,
				fmt.Sprintf(`
					EXISTS (
						SELECT 1 
						FROM catalog.product_variant_attribute_values pvav
						WHERE pvav.variant_id = pv.id
						AND pvav.attribute_id = $%d
						AND pvav.value_id = ANY($%d)
					)`, arg, arg+1))
			args = append(args, vaf.AttributeID, pq.Array(vaf.ValueIDs))
			arg += 2
		}

		subQuery := fmt.Sprintf(`
			EXISTS (
				SELECT 1
				FROM catalog.product_variants pv
				WHERE pv.product_id = p.id
				AND %s
			)`, strings.Join(variantConditions, " AND "))
		where = append(where, subQuery)
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

	variantQuery := `
		SELECT 
			v.id, v.product_id, v.variant_name, v.sku, v.price, v.is_active,
			COALESCE(i.quantity, 0)
		FROM catalog.product_variants v
		LEFT JOIN catalog.inventory i ON i.variant_id = v.id
		WHERE v.product_id = ANY($1)
	`

	varRows, err := r.db.Query(variantQuery, pq.Array(productIDs))
	if err != nil {
		return nil, err
	}
	defer varRows.Close()

	for varRows.Next() {
		var v models.ProductVariant
		var productID int64
		var inv models.Inventory

		err := varRows.Scan(
			&v.ID, &productID, &v.Name, &v.SKU, &v.Price, &v.IsActive,
			&inv.Quantity,
		)
		if err != nil {
			return nil, err
		}

		v.Inventory = inv
		products[productID].Variants = append(products[productID].Variants, v)
	}

	imageQuery := `
		SELECT id, product_id, variant_id, url, is_primary
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
			&img.ID, &productID, &img.VariantID, &img.URL, &img.IsPrimary,
		)
		if err != nil {
			return nil, err
		}

		p := products[productID]

		if img.VariantID == nil {
			p.Images = append(p.Images, img)
			continue
		}

		for i := range p.Variants {
			if p.Variants[i].ID == *img.VariantID {
				p.Variants[i].Images = append(p.Variants[i].Images, img)
				break
			}
		}
	}

	attrQuery := `
		SELECT pav.product_id, pav.attribute_id, pav.value_id, a.name, v.value
		FROM catalog.product_attribute_values pav
		JOIN catalog.attribute_names a ON a.id = pav.attribute_id
		JOIN catalog.attribute_values v ON v.id = pav.value_id
		WHERE pav.product_id = ANY($1)
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
			&pa.AttributeID,
			&pa.ValueID,
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
