package domain

import (
	"testing"
	"time"
)

func TestNewHabitReminder(t *testing.T) {
	habitID := 1
	userID := 1
	date := time.Now()

	reminder := NewHabitReminder(habitID, userID, date)

	if reminder.HabitID != habitID {
		t.Errorf("expected habit_id %d, got %d", habitID, reminder.HabitID)
	}
	if reminder.UserID != userID {
		t.Errorf("expected user_id %d, got %d", userID, reminder.UserID)
	}
	if !reminder.ReminderDate.Valid {
		t.Errorf("expected reminder_date to be set")
	}
	if reminder.IsCompleted {
		t.Errorf("expected is_completed to be false initially")
	}
}

func TestMarkReminderAsCompleted(t *testing.T) {
	reminder := NewHabitReminder(1, 1, time.Now())

	if reminder.IsCompleted {
		t.Errorf("expected is_completed to be false initially")
	}

	reminder.MarkAsCompleted()

	if !reminder.IsCompleted {
		t.Errorf("expected is_completed to be true")
	}
}

func TestMarkReminderAsIncomplete(t *testing.T) {
	reminder := NewHabitReminder(1, 1, time.Now())
	reminder.MarkAsCompleted()

	if !reminder.IsCompleted {
		t.Errorf("expected is_completed to be true")
	}

	reminder.MarkAsIncomplete()

	if reminder.IsCompleted {
		t.Errorf("expected is_completed to be false")
	}
}

func TestReminderToggleCompletion(t *testing.T) {
	reminder := NewHabitReminder(1, 1, time.Now())

	// Toggle multiple times
	reminder.MarkAsCompleted()
	if !reminder.IsCompleted {
		t.Errorf("expected is_completed to be true after first toggle")
	}

	reminder.MarkAsIncomplete()
	if reminder.IsCompleted {
		t.Errorf("expected is_completed to be false after second toggle")
	}

	reminder.MarkAsCompleted()
	if !reminder.IsCompleted {
		t.Errorf("expected is_completed to be true after third toggle")
	}
}

func TestReminderSentAt(t *testing.T) {
	reminder := NewHabitReminder(1, 1, time.Now())

	if reminder.SentAt.IsZero() {
		t.Errorf("expected sent_at to be set")
	}

	// Check if sent_at is close to now
	if time.Since(reminder.SentAt) > time.Second {
		t.Errorf("expected sent_at to be recent")
	}
}

func TestReminderDatePreservation(t *testing.T) {
	specificDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	reminder := NewHabitReminder(1, 1, specificDate)

	if !reminder.ReminderDate.Time.Equal(specificDate) {
		t.Errorf("expected reminder_date %v, got %v", specificDate, reminder.ReminderDate.Time)
	}
}

func TestMultipleRemindersForSameHabit(t *testing.T) {
	habitID := 1
	userID := 1

	date1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	reminder1 := NewHabitReminder(habitID, userID, date1)
	reminder2 := NewHabitReminder(habitID, userID, date2)

	if reminder1.ReminderDate.Time.Equal(reminder2.ReminderDate.Time) {
		t.Errorf("expected different reminder_dates")
	}

	reminder1.MarkAsCompleted()

	if reminder2.IsCompleted {
		t.Errorf("expected reminder2 to be independent from reminder1")
	}
}

func TestReminderCompletionPersistence(t *testing.T) {
	reminder := NewHabitReminder(1, 1, time.Now())

	// Complete and check multiple times
	reminder.MarkAsCompleted()
	if !reminder.IsCompleted {
		t.Errorf("expected is_completed to be true")
	}

	// Should still be true after multiple checks
	if !reminder.IsCompleted {
		t.Errorf("expected is_completed to remain true")
	}
}
