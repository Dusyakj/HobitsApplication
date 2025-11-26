package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"HobitsService/internal/domain"
)

// HabitRepository реализация интерфейса HabitRepository для PostgreSQL
type HabitRepository struct {
	pool *pgxpool.Pool
}

// NewHabitRepository создает новый HabitRepository
func NewHabitRepository(pool *pgxpool.Pool) *HabitRepository {
	return &HabitRepository{pool: pool}
}

// CreateHabit создает новую привычку
func (r *HabitRepository) CreateHabit(ctx context.Context, habit *domain.Habit) (*domain.Habit, error) {
	query := `
		INSERT INTO habits (
			user_id, name, description, goal, frequency, weekly_days, monthly_days,
			current_streak, best_streak, is_active, is_completed, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, user_id, name, description, goal, frequency, weekly_days, monthly_days,
			current_streak, best_streak, last_completed_date, last_checked_date,
			is_active, is_completed, created_at, updated_at, completed_at
	`

	row := r.pool.QueryRow(ctx, query,
		habit.UserID,
		habit.Name,
		habit.Description,
		habit.Goal,
		habit.Frequency,
		habit.WeeklyDays,
		habit.MonthlyDays,
		habit.CurrentStreak,
		habit.BestStreak,
		habit.IsActive,
		habit.IsCompleted,
		habit.CreatedAt,
		habit.UpdatedAt,
	)

	var result domain.Habit
	err := row.Scan(
		&result.ID,
		&result.UserID,
		&result.Name,
		&result.Description,
		&result.Goal,
		&result.Frequency,
		&result.WeeklyDays,
		&result.MonthlyDays,
		&result.CurrentStreak,
		&result.BestStreak,
		&result.LastCompletedDate,
		&result.LastCheckedDate,
		&result.IsActive,
		&result.IsCompleted,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create habit: %w", err)
	}

	return &result, nil
}

// GetHabitByID получает привычку по ID
func (r *HabitRepository) GetHabitByID(ctx context.Context, id int) (*domain.Habit, error) {
	query := `
		SELECT id, user_id, name, description, goal, frequency, weekly_days, monthly_days,
			current_streak, best_streak, last_completed_date, last_checked_date,
			is_active, is_completed, created_at, updated_at, completed_at
		FROM habits
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)

	var habit domain.Habit
	err := row.Scan(
		&habit.ID,
		&habit.UserID,
		&habit.Name,
		&habit.Description,
		&habit.Goal,
		&habit.Frequency,
		&habit.WeeklyDays,
		&habit.MonthlyDays,
		&habit.CurrentStreak,
		&habit.BestStreak,
		&habit.LastCompletedDate,
		&habit.LastCheckedDate,
		&habit.IsActive,
		&habit.IsCompleted,
		&habit.CreatedAt,
		&habit.UpdatedAt,
		&habit.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get habit by id: %w", err)
	}

	return &habit, nil
}

// GetHabitsByUserID получает все привычки пользователя
func (r *HabitRepository) GetHabitsByUserID(ctx context.Context, userID int) ([]*domain.Habit, error) {
	query := `
		SELECT id, user_id, name, description, goal, frequency, weekly_days, monthly_days,
			current_streak, best_streak, last_completed_date, last_checked_date,
			is_active, is_completed, created_at, updated_at, completed_at
		FROM habits
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get habits by user_id: %w", err)
	}
	defer rows.Close()

	var habits []*domain.Habit
	for rows.Next() {
		var habit domain.Habit
		err := rows.Scan(
			&habit.ID,
			&habit.UserID,
			&habit.Name,
			&habit.Description,
			&habit.Goal,
			&habit.Frequency,
			&habit.WeeklyDays,
			&habit.MonthlyDays,
			&habit.CurrentStreak,
			&habit.BestStreak,
			&habit.LastCompletedDate,
			&habit.LastCheckedDate,
			&habit.IsActive,
			&habit.IsCompleted,
			&habit.CreatedAt,
			&habit.UpdatedAt,
			&habit.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan habit: %w", err)
		}
		habits = append(habits, &habit)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating habits: %w", err)
	}

	return habits, nil
}

// GetActiveHabitsByUserID получает активные привычки пользователя
func (r *HabitRepository) GetActiveHabitsByUserID(ctx context.Context, userID int) ([]*domain.Habit, error) {
	query := `
		SELECT id, user_id, name, description, goal, frequency, weekly_days, monthly_days,
			current_streak, best_streak, last_completed_date, last_checked_date,
			is_active, is_completed, created_at, updated_at, completed_at
		FROM habits
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active habits by user_id: %w", err)
	}
	defer rows.Close()

	var habits []*domain.Habit
	for rows.Next() {
		var habit domain.Habit
		err := rows.Scan(
			&habit.ID,
			&habit.UserID,
			&habit.Name,
			&habit.Description,
			&habit.Goal,
			&habit.Frequency,
			&habit.WeeklyDays,
			&habit.MonthlyDays,
			&habit.CurrentStreak,
			&habit.BestStreak,
			&habit.LastCompletedDate,
			&habit.LastCheckedDate,
			&habit.IsActive,
			&habit.IsCompleted,
			&habit.CreatedAt,
			&habit.UpdatedAt,
			&habit.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan habit: %w", err)
		}
		habits = append(habits, &habit)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating habits: %w", err)
	}

	return habits, nil
}

// UpdateHabit обновляет привычку
func (r *HabitRepository) UpdateHabit(ctx context.Context, habit *domain.Habit) (*domain.Habit, error) {
	query := `
		UPDATE habits
		SET name = $1, description = $2, goal = $3, frequency = $4, weekly_days = $5,
			monthly_days = $6, current_streak = $7, best_streak = $8,
			last_completed_date = $9, last_checked_date = $10,
			is_active = $11, is_completed = $12, updated_at = $13, completed_at = $14
		WHERE id = $15
		RETURNING id, user_id, name, description, goal, frequency, weekly_days, monthly_days,
			current_streak, best_streak, last_completed_date, last_checked_date,
			is_active, is_completed, created_at, updated_at, completed_at
	`

	row := r.pool.QueryRow(ctx, query,
		habit.Name,
		habit.Description,
		habit.Goal,
		habit.Frequency,
		habit.WeeklyDays,
		habit.MonthlyDays,
		habit.CurrentStreak,
		habit.BestStreak,
		habit.LastCompletedDate,
		habit.LastCheckedDate,
		habit.IsActive,
		habit.IsCompleted,
		habit.UpdatedAt,
		habit.CompletedAt,
		habit.ID,
	)

	var result domain.Habit
	err := row.Scan(
		&result.ID,
		&result.UserID,
		&result.Name,
		&result.Description,
		&result.Goal,
		&result.Frequency,
		&result.WeeklyDays,
		&result.MonthlyDays,
		&result.CurrentStreak,
		&result.BestStreak,
		&result.LastCompletedDate,
		&result.LastCheckedDate,
		&result.IsActive,
		&result.IsCompleted,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update habit: %w", err)
	}

	return &result, nil
}

// DeleteHabit удаляет привычку
func (r *HabitRepository) DeleteHabit(ctx context.Context, id int) error {
	query := "DELETE FROM habits WHERE id = $1"
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete habit: %w", err)
	}
	return nil
}

// GetAllActiveHabits получает все активные привычки
func (r *HabitRepository) GetAllActiveHabits(ctx context.Context) ([]*domain.Habit, error) {
	query := `
		SELECT id, user_id, name, description, goal, frequency, weekly_days, monthly_days,
			current_streak, best_streak, last_completed_date, last_checked_date,
			is_active, is_completed, created_at, updated_at, completed_at
		FROM habits
		WHERE is_active = true
		ORDER BY id ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all active habits: %w", err)
	}
	defer rows.Close()

	var habits []*domain.Habit
	for rows.Next() {
		var habit domain.Habit
		err := rows.Scan(
			&habit.ID,
			&habit.UserID,
			&habit.Name,
			&habit.Description,
			&habit.Goal,
			&habit.Frequency,
			&habit.WeeklyDays,
			&habit.MonthlyDays,
			&habit.CurrentStreak,
			&habit.BestStreak,
			&habit.LastCompletedDate,
			&habit.LastCheckedDate,
			&habit.IsActive,
			&habit.IsCompleted,
			&habit.CreatedAt,
			&habit.UpdatedAt,
			&habit.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan habit: %w", err)
		}
		habits = append(habits, &habit)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating habits: %w", err)
	}

	return habits, nil
}

// GetHabitByUserIDAndName получает привычку по user ID и названию
func (r *HabitRepository) GetHabitByUserIDAndName(ctx context.Context, userID int, name string) (*domain.Habit, error) {
	query := `
		SELECT id, user_id, name, description, goal, frequency, weekly_days, monthly_days,
			current_streak, best_streak, last_completed_date, last_checked_date,
			is_active, is_completed, created_at, updated_at, completed_at
		FROM habits
		WHERE user_id = $1 AND name = $2
	`

	row := r.pool.QueryRow(ctx, query, userID, name)

	var habit domain.Habit
	err := row.Scan(
		&habit.ID,
		&habit.UserID,
		&habit.Name,
		&habit.Description,
		&habit.Goal,
		&habit.Frequency,
		&habit.WeeklyDays,
		&habit.MonthlyDays,
		&habit.CurrentStreak,
		&habit.BestStreak,
		&habit.LastCompletedDate,
		&habit.LastCheckedDate,
		&habit.IsActive,
		&habit.IsCompleted,
		&habit.CreatedAt,
		&habit.UpdatedAt,
		&habit.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get habit by user_id and name: %w", err)
	}

	return &habit, nil
}
