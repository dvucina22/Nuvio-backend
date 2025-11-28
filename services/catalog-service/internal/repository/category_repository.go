package repository

import (
	"database/sql"

	"github.com/catalog-service/pkg/models/products"
)

type CategoryRepository interface {
	GetCategories() ([]products.Category, error)
}

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) GetCategories() ([]products.Category, error) {
	rows, err := r.db.Query(`SELECT id, name FROM catalog.categories ORDER BY name`)
	if err != nil {
		return []products.Category{}, err
	}

	defer rows.Close()
	var categories []products.Category
	for rows.Next() {
		var c products.Category

		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			continue
		}
		categories = append(categories, c)
	}

	return categories, nil
}
