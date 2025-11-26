package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"HobitsService/internal/domain"
)

// UserRepository реализация интерфейса UserRepository для PostgreSQL
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository создает новый UserRepository
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// CreateUser создает нового пользователя
func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `
		INSERT INTO users (telegram_id, first_name, last_name, username, language_code, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, telegram_id, first_name, last_name, username, language_code, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		user.TelegramID,
		user.FirstName,
		user.LastName,
		user.Username,
		user.LanguageCode,
		user.CreatedAt,
		user.UpdatedAt,
	)

	var result domain.User
	err := row.Scan(
		&result.ID,
		&result.TelegramID,
		&result.FirstName,
		&result.LastName,
		&result.Username,
		&result.LanguageCode,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &result, nil
}

// GetUserByID получает пользователя по ID
func (r *UserRepository) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	query := `
		SELECT id, telegram_id, first_name, last_name, username, language_code, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)

	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.TelegramID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.LanguageCode,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// GetUserByTelegramID получает пользователя по Telegram ID
func (r *UserRepository) GetUserByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	query := `
		SELECT id, telegram_id, first_name, last_name, username, language_code, created_at, updated_at
		FROM users
		WHERE telegram_id = $1
	`

	row := r.pool.QueryRow(ctx, query, telegramID)

	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.TelegramID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.LanguageCode,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by telegram_id: %w", err)
	}

	return &user, nil
}

// UpdateUser обновляет пользователя
func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, username = $3, language_code = $4, updated_at = $5
		WHERE id = $6
		RETURNING id, telegram_id, first_name, last_name, username, language_code, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		user.FirstName,
		user.LastName,
		user.Username,
		user.LanguageCode,
		user.UpdatedAt,
		user.ID,
	)

	var result domain.User
	err := row.Scan(
		&result.ID,
		&result.TelegramID,
		&result.FirstName,
		&result.LastName,
		&result.Username,
		&result.LanguageCode,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &result, nil
}

// GetAllUsers получает всех пользователей
func (r *UserRepository) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	query := `
		SELECT id, telegram_id, first_name, last_name, username, language_code, created_at, updated_at
		FROM users
		ORDER BY id ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.TelegramID,
			&user.FirstName,
			&user.LastName,
			&user.Username,
			&user.LanguageCode,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// DeleteUser удаляет пользователя
func (r *UserRepository) DeleteUser(ctx context.Context, id int) error {
	query := "DELETE FROM users WHERE id = $1"
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
