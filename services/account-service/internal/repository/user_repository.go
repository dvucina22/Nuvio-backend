package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/account-service/pkg/models"
)

type UserRepository interface {
	GetUserInfo(userID string) (*models.UserMinimal, error)
	UpdateUserInfo(userID string, user *models.UpdateUser) error
	UpdateUserPassword(userID, hashedPassword string) error
	UpdateUserProfilePicture(userID string, profilePictureURL *string) error
	GetAllUsers() ([]models.UserAdmin, error)
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) GetUserInfo(userID string) (*models.UserMinimal, error) {
	const q = `SELECT id, first_name, last_name, email, profile_picture_url, gender, phone_number
	 FROM account.users WHERE id = $1`
	var u models.UserMinimal
	err := r.db.QueryRow(q, userID).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.ProfilePictureURL, &u.Gender, &u.PhoneNumber)
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
        phone_number = COALESCE($3, phone_number) ,
		gender = COALESCE($4, gen der),
		updated_at = NOW()
    WHERE id = $5`

	_, err := r.db.Exec(q, user.FirstName, user.LastName, user.PhoneNumber, user.Gender, userID)
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

func (r *userRepo) GetAllUsers() ([]models.UserAdmin, error) {
	const q = `
        SELECT 
            u.id,
            u.email,
            u.phone_number,
            u.first_name,
            u.last_name,
            u.is_active,
            u.created_at,
            u.updated_at,
            u.last_login_at,
            u.profile_picture_url,
            COALESCE(
                json_agg(
                    json_build_object('id', r2.id, 'name', r2.name)
                ) FILTER (WHERE r2.id IS NOT NULL),
                '[]'
            ) AS roles
        FROM account.users u
        LEFT JOIN account.user_roles_map m ON m.user_id = u.id
        LEFT JOIN account.user_roles r2 ON r2.id = m.role_id
        GROUP BY u.id
        ORDER BY u.created_at DESC;
    `

	rows, err := r.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.UserAdmin

	for rows.Next() {
		var u models.UserAdmin
		var rolesJSON []byte

		err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.PhoneNumber,
			&u.FirstName,
			&u.LastName,
			&u.IsActive,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.LastLoginAt,
			&u.ProfilePictureURL,
			&rolesJSON,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(rolesJSON, &u.Roles); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, nil
}
