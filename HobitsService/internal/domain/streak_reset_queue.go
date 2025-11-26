package domain

import (
	"database/sql"
	"time"
)

// StreakResetQueue представляет запись в очереди на сброс стрика
type StreakResetQueue struct {
	ID              int            `db:"id"`
	HabitID         int            `db:"habit_id"`
	UserID          int            `db:"user_id"`
	ResetDate       sql.NullTime   `db:"reset_date"`
	Processed       bool           `db:"processed"`
	ProcessedAt     sql.NullTime   `db:"processed_at"`
	PreviousStreak  sql.NullInt64  `db:"previous_streak"`
	CreatedAt       time.Time      `db:"created_at"`
}

// NewStreakResetQueue создает новую запись в очередь на сброс
func NewStreakResetQueue(habitID, userID int, resetDate time.Time) *StreakResetQueue {
	return &StreakResetQueue{
		HabitID:    habitID,
		UserID:     userID,
		ResetDate:  sql.NullTime{Time: resetDate, Valid: true},
		Processed:  false,
		CreatedAt:  time.Now(),
	}
}

// MarkAsProcessed отмечает запись как обработанную
func (srq *StreakResetQueue) MarkAsProcessed(previousStreak int) {
	srq.Processed = true
	now := time.Now()
	srq.ProcessedAt = sql.NullTime{Time: now, Valid: true}
	srq.PreviousStreak = sql.NullInt64{Int64: int64(previousStreak), Valid: true}
}

// GetResetDate возвращает дату сброса или нулевое время
func (srq *StreakResetQueue) GetResetDate() time.Time {
	if srq.ResetDate.Valid {
		return srq.ResetDate.Time
	}
	return time.Time{}
}

// GetPreviousStreak возвращает предыдущий стрик или 0
func (srq *StreakResetQueue) GetPreviousStreak() int {
	if srq.PreviousStreak.Valid {
		return int(srq.PreviousStreak.Int64)
	}
	return 0
}
