package repository

import (
	"database/sql"
	"avito-backend/src/internal/domain/models"
	sq "github.com/Masterminds/squirrel"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := psql.Insert("users").
		Columns("id", "email", "password_hash", "role").
		Values(user.ID, user.Email, user.PasswordHash, user.Role)

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(sql, args...)
	return err
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := psql.Select("id", "email", "password_hash", "role").
		From("users").
		Where(sq.Eq{"email": email})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	user := &models.User{}
	err = r.db.QueryRow(sql, args...).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		return nil, err
	}

	return user, nil
}
