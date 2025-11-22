package repository

import (
	"database/sql"

	"github.com/account-service/pkg/models"
)

type UserRepository interface {
	GetUserInfo(userID string) (*models.UserMinimal, error)
	UpdateUserInfo(userID string, user *models.UpdateUser) error
	UpdateUserPassword(userID, hashedPassword string) error
	UpdateUserProfilePicture(userID string, profilePictureURL *string) error
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) GetUserInfo(userID string) (*models.UserMinimal, error) {
	const q = `SELECT id, first_name, last_name, email, profile_picture_url, gender
	 FROM account.users WHERE id = $1`
	var u models.UserMinimal
	err := r.db.QueryRow(q, userID).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.ProfilePictureURL, &u.Gender)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) UpdateUserInfo(userID string, user *models.UpdateUser) error {
	const q = `UPDATE account.users SET 
        first_name = COALESCE($1, first_name), 
        last_name = COALESCE($2, last_name), 
        email = COALESCE($3, email), 
        phone_number = COALESCE($4, phone_number) ,
		gender = COALESCE($5, gender),
		updated_at = NOW()
    WHERE id = $6`

	_, err := r.db.Exec(q, user.FirstName, user.LastName, user.Email, user.PhoneNumber, user.Gender, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepo) UpdateUserPassword(userID, hashedPassword string) error {
	const q = `UPDATE account.users SET password_hash = $1 WHERE id = $2`

	_, err := r.db.Exec(q, hashedPassword, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepo) UpdateUserProfilePicture(userID string, profilePictureURL *string) error {
	const q = `UPDATE account.users SET profile_picture_url = $1 WHERE id = $2`

	_, err := r.db.Exec(q, profilePictureURL, userID)
	if err != nil {
		return err
	}

	return nil
}
