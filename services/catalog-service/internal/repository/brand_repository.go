package repository

import (
	"database/sql"

	"github.com/catalog-service/pkg/models/products"
)

type BrandRepository interface {
	GetBrands() ([]products.Brand, error)
}

type brandRepository struct {
	db *sql.DB
}

func NewBrandRepository(db *sql.DB) BrandRepository {
	return &brandRepository{db: db}
}

func (r *brandRepository) GetBrands() ([]products.Brand, error) {
	rows, err := r.db.Query(`SELECT id, name FROM catalog.brands ORDER BY name`)
	if err != nil {
		return []products.Brand{}, err
	}
	defer rows.Close()

	var brands []products.Brand
	for rows.Next() {
		var b products.Brand
		if err := rows.Scan(&b.ID, &b.Name); err != nil {
			continue
		}
		brands = append(brands, b)
	}

	return brands, nil
}
