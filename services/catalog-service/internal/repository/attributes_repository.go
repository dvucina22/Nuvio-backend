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
        SELECT name, value
        FROM catalog.attributes
        ORDER BY name, value;
    `

	rows, err := r.db.Query(attrQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grouped := make(map[string][]string)

	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		grouped[name] = append(grouped[name], value)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	brandRows, err := r.db.Query(`SELECT name FROM catalog.brands ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer brandRows.Close()

	for brandRows.Next() {
		var name string
		if err := brandRows.Scan(&name); err != nil {
			return nil, err
		}
		grouped["brand"] = append(grouped["brand"], name)
	}

	if err := brandRows.Err(); err != nil {
		return nil, err
	}

	categoryRows, err := r.db.Query(`SELECT name FROM catalog.categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer categoryRows.Close()

	for categoryRows.Next() {
		var name string
		if err := categoryRows.Scan(&name); err != nil {
			return nil, err
		}
		grouped["category"] = append(grouped["category"], name)
	}

	if err := categoryRows.Err(); err != nil {
		return nil, err
	}

	filters := make([]products.AttributeValues, 0, len(grouped))
	for name, values := range grouped {
		filters = append(filters, products.AttributeValues{
			Name:   name,
			Values: values,
		})
	}

	return filters, nil
}
