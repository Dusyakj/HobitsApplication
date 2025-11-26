package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"HobitsService/internal/domain"
)

// HabitLogRepository реализация интерфейса HabitLogRepository для PostgreSQL
type HabitLogRepository struct {
	pool *pgxpool.Pool
}

// NewHabitLogRepository создает новый HabitLogRepository
func NewHabitLogRepository(pool *pgxpool.Pool) *HabitLogRepository {
	return &HabitLogRepository{pool: pool}
}

// CreateLog создает новый лог выполнения
func (r *HabitLogRepository) CreateLog(ctx context.Context, log *domain.HabitLog) (*domain.HabitLog, error) {
	query := `
		INSERT INTO habit_logs (habit_id, user_id, comment, logged_date, logged_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, habit_id, user_id, comment, logged_date, logged_at
	`

	row := r.pool.QueryRow(ctx, query,
		log.HabitID,
		log.UserID,
		log.Comment,
		log.LoggedDate,
		log.LoggedAt,
	)

	var result domain.HabitLog
	err := row.Scan(
		&result.ID,
		&result.HabitID,
		&result.UserID,
		&result.Comment,
		&result.LoggedDate,
		&result.LoggedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create habit log: %w", err)
	}

	return &result, nil
}

// GetLogByID получает лог по ID
func (r *HabitLogRepository) GetLogByID(ctx context.Context, id int) (*domain.HabitLog, error) {
	query := `
		SELECT id, habit_id, user_id, comment, logged_date, logged_at
		FROM habit_logs
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)

	var log domain.HabitLog
	err := row.Scan(
		&log.ID,
		&log.HabitID,
		&log.UserID,
		&log.Comment,
		&log.LoggedDate,
		&log.LoggedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get habit log by id: %w", err)
	}

	return &log, nil
}

// GetLogsByHabitID получает логи по привычке
func (r *HabitLogRepository) GetLogsByHabitID(ctx context.Context, habitID int) ([]*domain.HabitLog, error) {
	query := `
		SELECT id, habit_id, user_id, comment, logged_date, logged_at
		FROM habit_logs
		WHERE habit_id = $1
		ORDER BY logged_date DESC
	`

	rows, err := r.pool.Query(ctx, query, habitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs by habit_id: %w", err)
	}
	defer rows.Close()

	var logs []*domain.HabitLog
	for rows.Next() {
		var log domain.HabitLog
		err := rows.Scan(
			&log.ID,
			&log.HabitID,
			&log.UserID,
			&log.Comment,
			&log.LoggedDate,
			&log.LoggedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan habit log: %w", err)
		}
		logs = append(logs, &log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating logs: %w", err)
	}

	return logs, nil
}

// GetLogsByHabitIDAndDate получает логи за определенный период
func (r *HabitLogRepository) GetLogsByHabitIDAndDate(ctx context.Context, habitID int, from, to time.Time) ([]*domain.HabitLog, error) {
	query := `
		SELECT id, habit_id, user_id, comment, logged_date, logged_at
		FROM habit_logs
		WHERE habit_id = $1 AND logged_date >= $2 AND logged_date <= $3
		ORDER BY logged_date DESC
	`

	rows, err := r.pool.Query(ctx, query, habitID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs by habit_id and date: %w", err)
	}
	defer rows.Close()

	var logs []*domain.HabitLog
	for rows.Next() {
		var log domain.HabitLog
		err := rows.Scan(
			&log.ID,
			&log.HabitID,
			&log.UserID,
			&log.Comment,
			&log.LoggedDate,
			&log.LoggedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan habit log: %w", err)
		}
		logs = append(logs, &log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating logs: %w", err)
	}

	return logs, nil
}

// GetLogByHabitIDAndDate получает лог за конкретный день
func (r *HabitLogRepository) GetLogByHabitIDAndDate(ctx context.Context, habitID int, date time.Time) (*domain.HabitLog, error) {
	query := `
		SELECT id, habit_id, user_id, comment, logged_date, logged_at
		FROM habit_logs
		WHERE habit_id = $1 AND logged_date = $2
	`

	row := r.pool.QueryRow(ctx, query, habitID, date)

	var log domain.HabitLog
	err := row.Scan(
		&log.ID,
		&log.HabitID,
		&log.UserID,
		&log.Comment,
		&log.LoggedDate,
		&log.LoggedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get log by habit_id and date: %w", err)
	}

	return &log, nil
}

// DeleteLog удаляет лог
func (r *HabitLogRepository) DeleteLog(ctx context.Context, id int) error {
	query := "DELETE FROM habit_logs WHERE id = $1"
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete habit log: %w", err)
	}
	return nil
}

// CountLogsByHabitIDAndDate считает логи за период
func (r *HabitLogRepository) CountLogsByHabitIDAndDate(ctx context.Context, habitID int, from, to time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM habit_logs
		WHERE habit_id = $1 AND logged_date >= $2 AND logged_date <= $3
	`

	var count int
	err := r.pool.QueryRow(ctx, query, habitID, from, to).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count logs: %w", err)
	}

	return count, nil
}
