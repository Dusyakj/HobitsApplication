package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"HobitsService/internal/domain"
)

// HabitReminderRepository реализация интерфейса HabitReminderRepository для PostgreSQL
type HabitReminderRepository struct {
	pool *pgxpool.Pool
}

// NewHabitReminderRepository создает новый HabitReminderRepository
func NewHabitReminderRepository(pool *pgxpool.Pool) *HabitReminderRepository {
	return &HabitReminderRepository{pool: pool}
}

// CreateReminder создает новое напоминание
func (r *HabitReminderRepository) CreateReminder(ctx context.Context, reminder *domain.HabitReminder) (*domain.HabitReminder, error) {
	query := `
		INSERT INTO habit_reminders (habit_id, user_id, reminder_date, is_completed, sent_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, habit_id, user_id, reminder_date, is_completed, sent_at
	`

	row := r.pool.QueryRow(ctx, query,
		reminder.HabitID,
		reminder.UserID,
		reminder.ReminderDate,
		reminder.IsCompleted,
		reminder.SentAt,
	)

	var result domain.HabitReminder
	err := row.Scan(
		&result.ID,
		&result.HabitID,
		&result.UserID,
		&result.ReminderDate,
		&result.IsCompleted,
		&result.SentAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create reminder: %w", err)
	}

	return &result, nil
}

// GetReminderByID получает напоминание по ID
func (r *HabitReminderRepository) GetReminderByID(ctx context.Context, id int) (*domain.HabitReminder, error) {
	query := `
		SELECT id, habit_id, user_id, reminder_date, is_completed, sent_at
		FROM habit_reminders
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)

	var reminder domain.HabitReminder
	err := row.Scan(
		&reminder.ID,
		&reminder.HabitID,
		&reminder.UserID,
		&reminder.ReminderDate,
		&reminder.IsCompleted,
		&reminder.SentAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get reminder by id: %w", err)
	}

	return &reminder, nil
}

// GetRemindersByUserID получает напоминания пользователя
func (r *HabitReminderRepository) GetRemindersByUserID(ctx context.Context, userID int) ([]*domain.HabitReminder, error) {
	query := `
		SELECT id, habit_id, user_id, reminder_date, is_completed, sent_at
		FROM habit_reminders
		WHERE user_id = $1
		ORDER BY reminder_date DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reminders by user_id: %w", err)
	}
	defer rows.Close()

	var reminders []*domain.HabitReminder
	for rows.Next() {
		var reminder domain.HabitReminder
		err := rows.Scan(
			&reminder.ID,
			&reminder.HabitID,
			&reminder.UserID,
			&reminder.ReminderDate,
			&reminder.IsCompleted,
			&reminder.SentAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}
		reminders = append(reminders, &reminder)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reminders: %w", err)
	}

	return reminders, nil
}

// GetRemindersByDate получает напоминания на дату
func (r *HabitReminderRepository) GetRemindersByDate(ctx context.Context, date time.Time) ([]*domain.HabitReminder, error) {
	query := `
		SELECT id, habit_id, user_id, reminder_date, is_completed, sent_at
		FROM habit_reminders
		WHERE reminder_date = $1
		ORDER BY sent_at DESC
	`

	rows, err := r.pool.Query(ctx, query, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get reminders by date: %w", err)
	}
	defer rows.Close()

	var reminders []*domain.HabitReminder
	for rows.Next() {
		var reminder domain.HabitReminder
		err := rows.Scan(
			&reminder.ID,
			&reminder.HabitID,
			&reminder.UserID,
			&reminder.ReminderDate,
			&reminder.IsCompleted,
			&reminder.SentAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}
		reminders = append(reminders, &reminder)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reminders: %w", err)
	}

	return reminders, nil
}

// GetRemindersByUserIDAndDate получает напоминания пользователя на дату
func (r *HabitReminderRepository) GetRemindersByUserIDAndDate(ctx context.Context, userID int, date time.Time) ([]*domain.HabitReminder, error) {
	query := `
		SELECT id, habit_id, user_id, reminder_date, is_completed, sent_at
		FROM habit_reminders
		WHERE user_id = $1 AND reminder_date = $2
		ORDER BY sent_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get reminders by user_id and date: %w", err)
	}
	defer rows.Close()

	var reminders []*domain.HabitReminder
	for rows.Next() {
		var reminder domain.HabitReminder
		err := rows.Scan(
			&reminder.ID,
			&reminder.HabitID,
			&reminder.UserID,
			&reminder.ReminderDate,
			&reminder.IsCompleted,
			&reminder.SentAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}
		reminders = append(reminders, &reminder)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reminders: %w", err)
	}

	return reminders, nil
}

// GetReminderByHabitIDAndDate получает напоминание по привычке и дате
func (r *HabitReminderRepository) GetReminderByHabitIDAndDate(ctx context.Context, habitID int, date time.Time) (*domain.HabitReminder, error) {
	query := `
		SELECT id, habit_id, user_id, reminder_date, is_completed, sent_at
		FROM habit_reminders
		WHERE habit_id = $1 AND reminder_date = $2
	`

	row := r.pool.QueryRow(ctx, query, habitID, date)

	var reminder domain.HabitReminder
	err := row.Scan(
		&reminder.ID,
		&reminder.HabitID,
		&reminder.UserID,
		&reminder.ReminderDate,
		&reminder.IsCompleted,
		&reminder.SentAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get reminder by habit_id and date: %w", err)
	}

	return &reminder, nil
}

// UpdateReminder обновляет напоминание
func (r *HabitReminderRepository) UpdateReminder(ctx context.Context, reminder *domain.HabitReminder) (*domain.HabitReminder, error) {
	query := `
		UPDATE habit_reminders
		SET is_completed = $1
		WHERE id = $2
		RETURNING id, habit_id, user_id, reminder_date, is_completed, sent_at
	`

	row := r.pool.QueryRow(ctx, query,
		reminder.IsCompleted,
		reminder.ID,
	)

	var result domain.HabitReminder
	err := row.Scan(
		&result.ID,
		&result.HabitID,
		&result.UserID,
		&result.ReminderDate,
		&result.IsCompleted,
		&result.SentAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update reminder: %w", err)
	}

	return &result, nil
}

// DeleteReminder удаляет напоминание
func (r *HabitReminderRepository) DeleteReminder(ctx context.Context, id int) error {
	query := "DELETE FROM habit_reminders WHERE id = $1"
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}
	return nil
}
