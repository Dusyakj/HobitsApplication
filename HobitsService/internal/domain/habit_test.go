package domain

import (
	"database/sql"
	"testing"
	"time"
)

func TestNewHabit(t *testing.T) {
	userID := 1
	name := "Exercise"
	frequency := FrequencyDaily

	habit := NewHabit(userID, name, frequency)

	if habit.UserID != userID {
		t.Errorf("expected user_id %d, got %d", userID, habit.UserID)
	}
	if habit.Name != name {
		t.Errorf("expected name %s, got %s", name, habit.Name)
	}
	if habit.Frequency != frequency {
		t.Errorf("expected frequency %s, got %s", frequency, habit.Frequency)
	}
	if habit.IsActive != true {
		t.Errorf("expected is_active to be true, got %v", habit.IsActive)
	}
	if habit.CurrentStreak != 0 {
		t.Errorf("expected current_streak 0, got %d", habit.CurrentStreak)
	}
	if habit.BestStreak != 0 {
		t.Errorf("expected best_streak 0, got %d", habit.BestStreak)
	}
}

func TestIncreaseStreak(t *testing.T) {
	habit := NewHabit(1, "Test", FrequencyDaily)

	if habit.CurrentStreak != 0 {
		t.Errorf("expected initial streak 0, got %d", habit.CurrentStreak)
	}

	habit.IncreaseStreak()
	if habit.CurrentStreak != 1 {
		t.Errorf("expected streak 1, got %d", habit.CurrentStreak)
	}

	habit.IncreaseStreak()
	if habit.CurrentStreak != 2 {
		t.Errorf("expected streak 2, got %d", habit.CurrentStreak)
	}
}

func TestResetStreak(t *testing.T) {
	habit := NewHabit(1, "Test", FrequencyDaily)
	habit.CurrentStreak = 5
	habit.BestStreak = 3

	habit.ResetStreak()

	if habit.CurrentStreak != 0 {
		t.Errorf("expected streak 0 after reset, got %d", habit.CurrentStreak)
	}
	if habit.BestStreak != 3 {
		t.Errorf("expected best_streak unchanged, got %d", habit.BestStreak)
	}
}

func TestMarkAsCompleted(t *testing.T) {
	habit := NewHabit(1, "Test", FrequencyDaily)

	if habit.IsCompleted {
		t.Errorf("expected is_completed to be false initially")
	}

	habit.MarkAsCompleted()

	if !habit.IsCompleted {
		t.Errorf("expected is_completed to be true")
	}
	if !habit.CompletedAt.Valid {
		t.Errorf("expected completed_at to be set")
	}
}

func TestUpdateLastCompletedDate(t *testing.T) {
	habit := NewHabit(1, "Test", FrequencyDaily)
	now := time.Now()

	habit.UpdateLastCompletedDate(now)

	if !habit.LastCompletedDate.Valid {
		t.Errorf("expected last_completed_date to be set")
	}
	if !habit.LastCompletedDate.Time.Equal(now) {
		t.Errorf("expected last_completed_date %v, got %v", now, habit.LastCompletedDate.Time)
	}
}

func TestSetWeeklyDays(t *testing.T) {
	habit := NewHabit(1, "Test", FrequencyWeekly)

	habit.SetWeeklyDays("1,3,5")

	if !habit.WeeklyDays.Valid {
		t.Errorf("expected weekly_days to be set")
	}
	if habit.WeeklyDays.String != "1,3,5" {
		t.Errorf("expected weekly_days '1,3,5', got '%s'", habit.WeeklyDays.String)
	}
}

func TestSetMonthlyDays(t *testing.T) {
	habit := NewHabit(1, "Test", FrequencyMonthly)

	habit.SetMonthlyDays("1,15,28")

	if !habit.MonthlyDays.Valid {
		t.Errorf("expected monthly_days to be set")
	}
	if habit.MonthlyDays.String != "1,15,28" {
		t.Errorf("expected monthly_days '1,15,28', got '%s'", habit.MonthlyDays.String)
	}
}

func TestActivateDeactivate(t *testing.T) {
	habit := NewHabit(1, "Test", FrequencyDaily)

	habit.Deactivate()
	if habit.IsActive {
		t.Errorf("expected is_active to be false after deactivate")
	}

	habit.Activate()
	if !habit.IsActive {
		t.Errorf("expected is_active to be true after activate")
	}
}

func TestHabitFrequencyValidation(t *testing.T) {
	tests := []struct {
		frequency HabitFrequency
		valid     bool
	}{
		{FrequencyDaily, true},
		{FrequencyWeekly, true},
		{FrequencyMonthly, true},
		{HabitFrequency("invalid"), false},
	}

	for _, tt := range tests {
		habit := NewHabit(1, "Test", tt.frequency)
		if tt.valid && habit.Frequency != tt.frequency {
			t.Errorf("expected frequency %s, got %s", tt.frequency, habit.Frequency)
		}
	}
}

func TestHabitBestStreakUpdate(t *testing.T) {
	habit := NewHabit(1, "Test", FrequencyDaily)
	habit.CurrentStreak = 10
	habit.BestStreak = 5

	// Simulate streak break with best streak update
	if habit.CurrentStreak > habit.BestStreak {
		habit.BestStreak = habit.CurrentStreak
	}
	habit.ResetStreak()

	if habit.BestStreak != 10 {
		t.Errorf("expected best_streak 10, got %d", habit.BestStreak)
	}
	if habit.CurrentStreak != 0 {
		t.Errorf("expected current_streak 0, got %d", habit.CurrentStreak)
	}
}

func TestHabitNullableFields(t *testing.T) {
	habit := NewHabit(1, "Test", FrequencyDaily)

	tests := []struct {
		name  string
		field sql.NullString
	}{
		{"Description", habit.Description},
		{"Goal", habit.Goal},
		{"WeeklyDays", habit.WeeklyDays},
		{"MonthlyDays", habit.MonthlyDays},
	}

	for _, tt := range tests {
		if tt.field.Valid {
			t.Errorf("expected %s to be null initially", tt.name)
		}
	}
}
