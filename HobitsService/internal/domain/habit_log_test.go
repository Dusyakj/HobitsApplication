package domain

import (
	"testing"
	"time"
)

func TestNewHabitLog(t *testing.T) {
	habitID := 1
	userID := 1
	date := time.Now()
	comment := "Test comment"

	log := NewHabitLog(habitID, userID, date, comment)

	if log.HabitID != habitID {
		t.Errorf("expected habit_id %d, got %d", habitID, log.HabitID)
	}
	if log.UserID != userID {
		t.Errorf("expected user_id %d, got %d", userID, log.UserID)
	}
	if log.LoggedDate != date {
		t.Errorf("expected logged_date %v, got %v", date, log.LoggedDate)
	}
	if log.Comment.String != comment {
		t.Errorf("expected comment %s, got %s", comment, log.Comment.String)
	}
}

func TestHabitLogTimestamp(t *testing.T) {
	log := NewHabitLog(1, 1, time.Now(), "")

	if log.LoggedAt.IsZero() {
		t.Errorf("expected logged_at to be set")
	}

	// Check if logged_at is close to now (within 1 second)
	if time.Since(log.LoggedAt) > time.Second {
		t.Errorf("expected logged_at to be recent")
	}
}

func TestHabitLogWithEmptyComment(t *testing.T) {
	log := NewHabitLog(1, 1, time.Now(), "")

	// Empty comment should not be valid (based on domain/habit_log.go line 24)
	if log.Comment.Valid {
		t.Errorf("expected comment to be null for empty string")
	}
}

func TestHabitLogWithLongComment(t *testing.T) {
	longComment := "This is a very long comment that describes the habit completion in detail. " +
		"It can contain multiple sentences and provide context about the activity."

	log := NewHabitLog(1, 1, time.Now(), longComment)

	if log.Comment.String != longComment {
		t.Errorf("expected comment %s, got %s", longComment, log.Comment.String)
	}
}

func TestMultipleHabitLogs(t *testing.T) {
	habitID := 1
	userID := 1

	date1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	log1 := NewHabitLog(habitID, userID, date1, "Day 1")
	log2 := NewHabitLog(habitID, userID, date2, "Day 2")

	if log1.LoggedDate.Equal(log2.LoggedDate) {
		t.Errorf("expected different logged_dates")
	}
	if log1.Comment == log2.Comment {
		t.Errorf("expected different comments")
	}
}

func TestHabitLogDatePreservation(t *testing.T) {
	specificDate := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)

	log := NewHabitLog(1, 1, specificDate, "Test")

	if !log.LoggedDate.Equal(specificDate) {
		t.Errorf("expected logged_date %v, got %v", specificDate, log.LoggedDate)
	}
}
