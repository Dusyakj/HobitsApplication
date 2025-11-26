package domain

import (
	"database/sql"
	"time"
)

// HabitReminder представляет отправленное напоминание о привычке
type HabitReminder struct {
	ID           int            `db:"id"`
	HabitID      int            `db:"habit_id"`
	UserID       int            `db:"user_id"`
	ReminderDate sql.NullTime   `db:"reminder_date"`
	IsCompleted  bool           `db:"is_completed"`
	SentAt       time.Time      `db:"sent_at"`
}

// NewHabitReminder создает новое напоминание
func NewHabitReminder(habitID, userID int, reminderDate time.Time) *HabitReminder {
	return &HabitReminder{
		HabitID:      habitID,
		UserID:       userID,
		ReminderDate: sql.NullTime{Time: reminderDate, Valid: true},
		IsCompleted:  false,
		SentAt:       time.Now(),
	}
}

// MarkAsCompleted отмечает напоминание как выполненное
func (hr *HabitReminder) MarkAsCompleted() {
	hr.IsCompleted = true
}

// MarkAsIncomplete отмечает напоминание как невыполненное
func (hr *HabitReminder) MarkAsIncomplete() {
	hr.IsCompleted = false
}
