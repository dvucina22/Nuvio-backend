package repository

import (
	"database/sql"

	"github.com/account-service/pkg/models"
)

type UserRepository interface {
	GetUserInfo(userID string) (*models.UserMinimal, error)
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) GetUserInfo(userID string) (*models.UserMinimal, error) {
	const q = `SELECT id, first_name, last_name, email FROM account.users WHERE id = $1`
	var u models.UserMinimal
	err := r.db.QueryRow(q, userID).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
