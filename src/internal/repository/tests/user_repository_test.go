package repository_test

import (
	"avito-backend/src/internal/domain/models"
	"avito-backend/src/internal/repository"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewUserRepository(db)
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         string(models.EmployeeRole),
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (id,email,password_hash,role) VALUES ($1,$2,$3,$4)`)).
			WithArgs(user.ID, user.Email, user.PasswordHash, user.Role).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(user)
		require.NoError(t, err)
	})

	t.Run("Duplicate Email", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (id,email,password_hash,role) VALUES ($1,$2,$3,$4)`)).
			WithArgs(user.ID, user.Email, user.PasswordHash, user.Role).
			WillReturnError(sql.ErrConnDone) 

		err := repo.Create(user)
		require.Error(t, err)
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewUserRepository(db)
	userID := uuid.New()
	email := "test@example.com"

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "role"}).
			AddRow(userID, email, "hashed_password", string(models.EmployeeRole))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, password_hash, role FROM users WHERE email = $1`)).
			WithArgs(email).
			WillReturnRows(rows)

		user, err := repo.GetByEmail(email)
		require.NoError(t, err)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, userID, user.ID)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, password_hash, role FROM users WHERE email = $1`)).
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetByEmail(email)
		require.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewUserRepository(db)
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "role"}).
			AddRow(userID, "test@example.com", "hashed_password", string(models.EmployeeRole))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, password_hash, role FROM users WHERE id = $1`)).
			WithArgs(userID).
			WillReturnRows(rows)

		user, err := repo.GetByID(userID)
		require.NoError(t, err)
		assert.Equal(t, userID, user.ID)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, password_hash, role FROM users WHERE id = $1`)).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetByID(userID)
		require.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewUserRepository(db)
	user := &models.User{
		ID:           uuid.New(),
		Email:        "updated@example.com",
		PasswordHash: "new_hash",
		Role:         string(models.EmployeeRole),
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET email = $1, password_hash = $2, role = $3 WHERE id = $4`)).
			WithArgs(user.Email, user.PasswordHash, user.Role, user.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Update(user)
		require.NoError(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET email = $1, password_hash = $2, role = $3 WHERE id = $4`)).
			WithArgs(user.Email, user.PasswordHash, user.Role, user.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(user)
		require.NoError(t, err) 
	})
}

func TestUserRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewUserRepository(db)
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE id = $1`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Delete(userID)
		require.NoError(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE id = $1`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(userID)
		require.NoError(t, err)
	})

	t.Run("DB Error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE id = $1`)).
			WithArgs(userID).
			WillReturnError(sql.ErrConnDone)

		err := repo.Delete(userID)
		require.Error(t, err)
	})
}