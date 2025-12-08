package repository

import (
	"database/sql"

	"github.com/catalog-service/pkg/models/products"
)

type AttributesRepository interface {
	GetAttributes() ([]products.AttributeValues, error)
}

type attributesRepository struct {
	db *sql.DB
}

func NewAttributesRepository(db *sql.DB) AttributesRepository {
	return &attributesRepository{db: db}
}

func (r *attributesRepository) GetAttributes() ([]products.AttributeValues, error) {
	const attrQuery = `
        SELECT id, name, value
        FROM catalog.attributes
        ORDER BY name, value;
    `

	rows, err := r.db.Query(attrQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grouped := make(map[string][]products.AttributeItem)

	for rows.Next() {
		var (
			id    int64
			name  string
			value string
		)
		if err := rows.Scan(&id, &name, &value); err != nil {
			return nil, err
		}

		grouped[name] = append(grouped[name], products.AttributeItem{
			ID:    id,
			Value: value,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	brandRows, err := r.db.Query(`SELECT id, name FROM catalog.brands ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer brandRows.Close()

	for brandRows.Next() {
		var (
			id   int64
			name string
		)
		if err := brandRows.Scan(&id, &name); err != nil {
			return nil, err
		}

		grouped["brand"] = append(grouped["brand"], products.AttributeItem{
			ID:    id,
			Value: name,
		})
	}

	if err := brandRows.Err(); err != nil {
		return nil, err
	}

	categoryRows, err := r.db.Query(`SELECT id, name FROM catalog.categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer categoryRows.Close()

	for categoryRows.Next() {
		var (
			id   int64
			name string
		)
		if err := categoryRows.Scan(&id, &name); err != nil {
			return nil, err
		}

		grouped["category"] = append(grouped["category"], products.AttributeItem{
			ID:    id,
			Value: name,
		})
	}

	if err := categoryRows.Err(); err != nil {
		return nil, err
	}

	var filters []products.AttributeValues
	for name, items := range grouped {
		filters = append(filters, products.AttributeValues{
			Name:  name,
			Items: items,
		})
	}

	return filters, nil
}
