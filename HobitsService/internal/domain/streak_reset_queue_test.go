package domain

import (
	"testing"
	"time"
)

func TestNewStreakResetQueue(t *testing.T) {
	habitID := 1
	userID := 1
	resetDate := time.Now()

	entry := NewStreakResetQueue(habitID, userID, resetDate)

	if entry.HabitID != habitID {
		t.Errorf("expected habit_id %d, got %d", habitID, entry.HabitID)
	}
	if entry.UserID != userID {
		t.Errorf("expected user_id %d, got %d", userID, entry.UserID)
	}
	if !entry.ResetDate.Time.Equal(resetDate) {
		t.Errorf("expected reset_date %v, got %v", resetDate, entry.ResetDate.Time)
	}
	if entry.Processed {
		t.Errorf("expected processed to be false initially")
	}
	if entry.ProcessedAt.Valid {
		t.Errorf("expected processed_at to be null initially")
	}
}

func TestMarkQueueEntryAsProcessed(t *testing.T) {
	entry := NewStreakResetQueue(1, 1, time.Now())
	previousStreak := 5

	if entry.Processed {
		t.Errorf("expected processed to be false initially")
	}

	entry.MarkAsProcessed(previousStreak)

	if !entry.Processed {
		t.Errorf("expected processed to be true")
	}
	if !entry.ProcessedAt.Valid {
		t.Errorf("expected processed_at to be set")
	}
	if !entry.PreviousStreak.Valid || entry.PreviousStreak.Int64 != int64(previousStreak) {
		t.Errorf("expected previous_streak %d, got %d", previousStreak, entry.PreviousStreak.Int64)
	}
}

func TestQueueEntryGetResetDate(t *testing.T) {
	resetDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	entry := NewStreakResetQueue(1, 1, resetDate)

	retrieved := entry.GetResetDate()

	if !retrieved.Equal(resetDate) {
		t.Errorf("expected reset_date %v, got %v", resetDate, retrieved)
	}
}

func TestQueueEntryProcessedAtTimestamp(t *testing.T) {
	entry := NewStreakResetQueue(1, 1, time.Now())
	entry.MarkAsProcessed(5)

	if !entry.ProcessedAt.Valid {
		t.Errorf("expected processed_at to be set")
	}

	// Check if processed_at is close to now
	if time.Since(entry.ProcessedAt.Time) > time.Second {
		t.Errorf("expected processed_at to be recent")
	}
}

func TestMultipleQueueEntries(t *testing.T) {
	date1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	entry1 := NewStreakResetQueue(1, 1, date1)
	entry2 := NewStreakResetQueue(2, 1, date2)

	if entry1.HabitID == entry2.HabitID {
		t.Errorf("expected different habit_ids")
	}
	if entry1.ResetDate.Time.Equal(entry2.ResetDate.Time) {
		t.Errorf("expected different reset_dates")
	}
}

func TestQueueEntryPreviousStreakVariations(t *testing.T) {
	tests := []int{0, 1, 5, 10, 100, 365}

	for _, previousStreak := range tests {
		entry := NewStreakResetQueue(1, 1, time.Now())
		entry.MarkAsProcessed(previousStreak)

		if !entry.PreviousStreak.Valid || entry.PreviousStreak.Int64 != int64(previousStreak) {
			t.Errorf("expected previous_streak %d, got %d", previousStreak, entry.PreviousStreak.Int64)
		}
	}
}

func TestQueueEntryNullableProcessedAt(t *testing.T) {
	entry := NewStreakResetQueue(1, 1, time.Now())

	// Initially should be null
	if entry.ProcessedAt.Valid {
		t.Errorf("expected processed_at to be null initially")
	}

	// After marking as processed should be set
	entry.MarkAsProcessed(5)
	if !entry.ProcessedAt.Valid {
		t.Errorf("expected processed_at to be set after marking as processed")
	}
}

func TestQueueEntryForDifferentUsers(t *testing.T) {
	resetDate := time.Now()

	entry1 := NewStreakResetQueue(1, 1, resetDate)
	entry2 := NewStreakResetQueue(1, 2, resetDate)

	if entry1.UserID == entry2.UserID {
		t.Errorf("expected different user_ids")
	}
	if entry1.HabitID != entry2.HabitID {
		t.Errorf("expected same habit_id")
	}
}

func TestQueueEntryResetDatePreservation(t *testing.T) {
	specificDate := time.Date(2024, 12, 25, 10, 30, 0, 0, time.UTC)
	entry := NewStreakResetQueue(1, 1, specificDate)

	retrieved := entry.GetResetDate()

	if !retrieved.Equal(specificDate) {
		t.Errorf("expected reset_date to be preserved")
	}
}
