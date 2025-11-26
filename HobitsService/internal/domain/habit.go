package domain

import (
	"database/sql"
	"time"
)

// HabitFrequency тип частоты привычки
type HabitFrequency string

const (
	FrequencyDaily   HabitFrequency = "daily"
	FrequencyWeekly  HabitFrequency = "weekly"
	FrequencyMonthly HabitFrequency = "monthly"
)

// Habit представляет привычку пользователя
type Habit struct {
	ID                int                `db:"id"`
	UserID            int                `db:"user_id"`
	Name              string             `db:"name"`
	Description       sql.NullString     `db:"description"`
	Goal              sql.NullString     `db:"goal"`
	Frequency         HabitFrequency     `db:"frequency"`
	WeeklyDays        sql.NullString     `db:"weekly_days"`
	MonthlyDays       sql.NullString     `db:"monthly_days"`
	CurrentStreak     int                `db:"current_streak"`
	BestStreak        int                `db:"best_streak"`
	LastCompletedDate sql.NullTime       `db:"last_completed_date"`
	LastCheckedDate   sql.NullTime       `db:"last_checked_date"`
	IsActive          bool               `db:"is_active"`
	IsCompleted       bool               `db:"is_completed"`
	CreatedAt         time.Time          `db:"created_at"`
	UpdatedAt         time.Time          `db:"updated_at"`
	CompletedAt       sql.NullTime       `db:"completed_at"`
}

// NewHabit создает новую привычку
func NewHabit(userID int, name string, frequency HabitFrequency) *Habit {
	now := time.Now()
	return &Habit{
		UserID:        userID,
		Name:          name,
		Frequency:     frequency,
		IsActive:      true,
		IsCompleted:   false,
		CurrentStreak: 0,
		BestStreak:    0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// SetDescription устанавливает описание привычки
func (h *Habit) SetDescription(description string) {
	h.Description = sql.NullString{String: description, Valid: description != ""}
	h.UpdatedAt = time.Now()
}

// SetGoal устанавливает цель привычки
func (h *Habit) SetGoal(goal string) {
	h.Goal = sql.NullString{String: goal, Valid: goal != ""}
	h.UpdatedAt = time.Now()
}

// SetWeeklyDays устанавливает дни недели (для еженедельных привычек)
// days формат: "1,3,5" где 1=пн, 7=вс
func (h *Habit) SetWeeklyDays(days string) {
	h.WeeklyDays = sql.NullString{String: days, Valid: days != ""}
	h.UpdatedAt = time.Now()
}

// SetMonthlyDays устанавливает дни месяца (для ежемесячных привычек)
// days формат: "1,15,28"
func (h *Habit) SetMonthlyDays(days string) {
	h.MonthlyDays = sql.NullString{String: days, Valid: days != ""}
	h.UpdatedAt = time.Now()
}

// Deactivate деактивирует привычку
func (h *Habit) Deactivate() {
	h.IsActive = false
	h.UpdatedAt = time.Now()
}

// Activate активирует привычку
func (h *Habit) Activate() {
	h.IsActive = true
	h.UpdatedAt = time.Now()
}

// IncreaseStreak увеличивает текущий стрик на 1
func (h *Habit) IncreaseStreak() {
	h.CurrentStreak++
	if h.CurrentStreak > h.BestStreak {
		h.BestStreak = h.CurrentStreak
	}
	h.UpdatedAt = time.Now()
}

// ResetStreak сбрасывает стрик на 0
func (h *Habit) ResetStreak() {
	h.CurrentStreak = 0
	h.UpdatedAt = time.Now()
}

// MarkAsCompleted отмечает привычку как выработанную
func (h *Habit) MarkAsCompleted() {
	h.IsCompleted = true
	now := time.Now()
	h.CompletedAt = sql.NullTime{Time: now, Valid: true}
	h.UpdatedAt = now
}

// UpdateLastCompletedDate обновляет дату последнего выполнения
func (h *Habit) UpdateLastCompletedDate(date time.Time) {
	h.LastCompletedDate = sql.NullTime{Time: date, Valid: true}
	h.UpdatedAt = time.Now()
}

// UpdateLastCheckedDate обновляет дату последней проверки
func (h *Habit) UpdateLastCheckedDate(date time.Time) {
	h.LastCheckedDate = sql.NullTime{Time: date, Valid: true}
	h.UpdatedAt = time.Now()
}
