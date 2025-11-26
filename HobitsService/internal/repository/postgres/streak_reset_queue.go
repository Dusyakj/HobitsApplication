package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"HobitsService/internal/domain"
)

// StreakResetQueueRepository реализация интерфейса StreakResetQueueRepository для PostgreSQL
type StreakResetQueueRepository struct {
	pool *pgxpool.Pool
}

// NewStreakResetQueueRepository создает новый StreakResetQueueRepository
func NewStreakResetQueueRepository(pool *pgxpool.Pool) *StreakResetQueueRepository {
	return &StreakResetQueueRepository{pool: pool}
}

// CreateQueueEntry создает новую запись в очередь
func (r *StreakResetQueueRepository) CreateQueueEntry(ctx context.Context, entry *domain.StreakResetQueue) (*domain.StreakResetQueue, error) {
	query := `
		INSERT INTO streak_reset_queue (habit_id, user_id, reset_date, processed, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, habit_id, user_id, reset_date, processed, processed_at, previous_streak, created_at
	`

	row := r.pool.QueryRow(ctx, query,
		entry.HabitID,
		entry.UserID,
		entry.ResetDate,
		entry.Processed,
		entry.CreatedAt,
	)

	var result domain.StreakResetQueue
	err := row.Scan(
		&result.ID,
		&result.HabitID,
		&result.UserID,
		&result.ResetDate,
		&result.Processed,
		&result.ProcessedAt,
		&result.PreviousStreak,
		&result.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue entry: %w", err)
	}

	return &result, nil
}

// GetQueueEntryByID получает запись по ID
func (r *StreakResetQueueRepository) GetQueueEntryByID(ctx context.Context, id int) (*domain.StreakResetQueue, error) {
	query := `
		SELECT id, habit_id, user_id, reset_date, processed, processed_at, previous_streak, created_at
		FROM streak_reset_queue
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)

	var entry domain.StreakResetQueue
	err := row.Scan(
		&entry.ID,
		&entry.HabitID,
		&entry.UserID,
		&entry.ResetDate,
		&entry.Processed,
		&entry.ProcessedAt,
		&entry.PreviousStreak,
		&entry.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue entry by id: %w", err)
	}

	return &entry, nil
}

// GetUnprocessedEntries получает необработанные записи
func (r *StreakResetQueueRepository) GetUnprocessedEntries(ctx context.Context) ([]*domain.StreakResetQueue, error) {
	query := `
		SELECT id, habit_id, user_id, reset_date, processed, processed_at, previous_streak, created_at
		FROM streak_reset_queue
		WHERE processed = false
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed entries: %w", err)
	}
	defer rows.Close()

	var entries []*domain.StreakResetQueue
	for rows.Next() {
		var entry domain.StreakResetQueue
		err := rows.Scan(
			&entry.ID,
			&entry.HabitID,
			&entry.UserID,
			&entry.ResetDate,
			&entry.Processed,
			&entry.ProcessedAt,
			&entry.PreviousStreak,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan queue entry: %w", err)
		}
		entries = append(entries, &entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entries: %w", err)
	}

	return entries, nil
}

// GetUnprocessedEntriesByDate получает необработанные записи на дату
func (r *StreakResetQueueRepository) GetUnprocessedEntriesByDate(ctx context.Context, date time.Time) ([]*domain.StreakResetQueue, error) {
	query := `
		SELECT id, habit_id, user_id, reset_date, processed, processed_at, previous_streak, created_at
		FROM streak_reset_queue
		WHERE processed = false AND reset_date = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed entries by date: %w", err)
	}
	defer rows.Close()

	var entries []*domain.StreakResetQueue
	for rows.Next() {
		var entry domain.StreakResetQueue
		err := rows.Scan(
			&entry.ID,
			&entry.HabitID,
			&entry.UserID,
			&entry.ResetDate,
			&entry.Processed,
			&entry.ProcessedAt,
			&entry.PreviousStreak,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan queue entry: %w", err)
		}
		entries = append(entries, &entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entries: %w", err)
	}

	return entries, nil
}

// UpdateQueueEntry обновляет запись в очередь
func (r *StreakResetQueueRepository) UpdateQueueEntry(ctx context.Context, entry *domain.StreakResetQueue) (*domain.StreakResetQueue, error) {
	query := `
		UPDATE streak_reset_queue
		SET processed = $1, processed_at = $2, previous_streak = $3
		WHERE id = $4
		RETURNING id, habit_id, user_id, reset_date, processed, processed_at, previous_streak, created_at
	`

	row := r.pool.QueryRow(ctx, query,
		entry.Processed,
		entry.ProcessedAt,
		entry.PreviousStreak,
		entry.ID,
	)

	var result domain.StreakResetQueue
	err := row.Scan(
		&result.ID,
		&result.HabitID,
		&result.UserID,
		&result.ResetDate,
		&result.Processed,
		&result.ProcessedAt,
		&result.PreviousStreak,
		&result.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update queue entry: %w", err)
	}

	return &result, nil
}

// DeleteQueueEntry удаляет запись из очереди
func (r *StreakResetQueueRepository) DeleteQueueEntry(ctx context.Context, id int) error {
	query := "DELETE FROM streak_reset_queue WHERE id = $1"
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete queue entry: %w", err)
	}
	return nil
}

// GetQueueEntryByHabitIDAndDate получает запись по привычке и дате
func (r *StreakResetQueueRepository) GetQueueEntryByHabitIDAndDate(ctx context.Context, habitID int, date time.Time) (*domain.StreakResetQueue, error) {
	query := `
		SELECT id, habit_id, user_id, reset_date, processed, processed_at, previous_streak, created_at
		FROM streak_reset_queue
		WHERE habit_id = $1 AND reset_date = $2
	`

	row := r.pool.QueryRow(ctx, query, habitID, date)

	var entry domain.StreakResetQueue
	err := row.Scan(
		&entry.ID,
		&entry.HabitID,
		&entry.UserID,
		&entry.ResetDate,
		&entry.Processed,
		&entry.ProcessedAt,
		&entry.PreviousStreak,
		&entry.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue entry by habit_id and date: %w", err)
	}

	return &entry, nil
}
