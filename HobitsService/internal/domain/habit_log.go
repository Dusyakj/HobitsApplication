package domain

import (
	"database/sql"
	"time"
)

// HabitLog представляет логирование выполнения привычки
type HabitLog struct {
	ID        int            `db:"id"`
	HabitID   int            `db:"habit_id"`
	UserID    int            `db:"user_id"`
	Comment   sql.NullString `db:"comment"`
	LoggedDate time.Time     `db:"logged_date"`
	LoggedAt  time.Time      `db:"logged_at"`
}

// NewHabitLog создает новый лог выполнения привычки
func NewHabitLog(habitID, userID int, loggedDate time.Time, comment string) *HabitLog {
	return &HabitLog{
		HabitID:   habitID,
		UserID:    userID,
		LoggedDate: loggedDate,
		Comment:   sql.NullString{String: comment, Valid: comment != ""},
		LoggedAt:  time.Now(),
	}
}

// GetComment возвращает комментарий или пустую строку
func (hl *HabitLog) GetComment() string {
	if hl.Comment.Valid {
		return hl.Comment.String
	}
	return ""
}
