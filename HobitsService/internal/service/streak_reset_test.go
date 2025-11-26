package service

import (
	"context"
	"testing"
	"time"

	"HobitsService/internal/domain"
)

func TestCheckHabitStreak_FirstCompletion(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()
	queueRepo := newMockStreakResetQueueRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	streakResetService := NewStreakResetService(queueRepo, habitRepo, logRepo, reminderRepo, habitService)
	ctx := context.Background()

	habit, _ := habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)

	err := streakResetService.CheckHabitStreak(ctx, habit.ID)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCheckHabitStreak_DeactivatedHabit(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()
	queueRepo := newMockStreakResetQueueRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	streakResetService := NewStreakResetService(queueRepo, habitRepo, logRepo, reminderRepo, habitService)
	ctx := context.Background()

	habit, _ := habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)
	habitService.DeactivateHabit(ctx, habit.ID)

	err := streakResetService.CheckHabitStreak(ctx, habit.ID)

	if err != nil {
		t.Errorf("expected no error for deactivated habit")
	}
}

func TestCheckHabitStreak_NeverCompleted(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()
	queueRepo := newMockStreakResetQueueRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	streakResetService := NewStreakResetService(queueRepo, habitRepo, logRepo, reminderRepo, habitService)
	ctx := context.Background()

	habit, _ := habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)

	err := streakResetService.CheckHabitStreak(ctx, habit.ID)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// No queue entries should be created for habit never completed
	entries, _ := queueRepo.GetUnprocessedEntries(ctx)
	if len(entries) != 0 {
		t.Errorf("expected no queue entries for never completed habit")
	}
}

func TestProcessQueueEntries(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()
	queueRepo := newMockStreakResetQueueRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	streakResetService := NewStreakResetService(queueRepo, habitRepo, logRepo, reminderRepo, habitService)
	ctx := context.Background()

	habit, _ := habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)
	habit.CurrentStreak = 5
	habitRepo.UpdateHabit(ctx, habit)

	entry := domain.NewStreakResetQueue(habit.ID, 1, time.Now())
	queueRepo.CreateQueueEntry(ctx, entry)

	err := streakResetService.ProcessQueueEntries(ctx)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Check habit streak was reset
	updated, _ := habitService.GetHabit(ctx, habit.ID)
	if updated.CurrentStreak != 0 {
		t.Errorf("expected streak to be reset to 0, got %d", updated.CurrentStreak)
	}

	// Check entry is marked as processed
	processedEntry, _ := queueRepo.GetQueueEntryByID(ctx, entry.ID)
	if !processedEntry.Processed {
		t.Errorf("expected queue entry to be marked as processed")
	}
}

func TestProcessQueueEntry_BestStreakUpdate(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()
	queueRepo := newMockStreakResetQueueRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	streakResetService := NewStreakResetService(queueRepo, habitRepo, logRepo, reminderRepo, habitService)
	ctx := context.Background()

	habit, _ := habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)
	habit.CurrentStreak = 10
	habit.BestStreak = 5
	habitRepo.UpdateHabit(ctx, habit)

	entry := domain.NewStreakResetQueue(habit.ID, 1, time.Now())
	queueRepo.CreateQueueEntry(ctx, entry)

	streakResetService.ProcessQueueEntries(ctx)

	updated, _ := habitService.GetHabit(ctx, habit.ID)
	if updated.BestStreak != 10 {
		t.Errorf("expected best_streak to be updated to 10, got %d", updated.BestStreak)
	}
	if updated.CurrentStreak != 0 {
		t.Errorf("expected current_streak to be reset to 0")
	}
}

func TestGetUnprocessedQueueEntries(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()
	queueRepo := newMockStreakResetQueueRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	streakResetService := NewStreakResetService(queueRepo, habitRepo, logRepo, reminderRepo, habitService)
	ctx := context.Background()

	habit, _ := habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)

	entry1 := domain.NewStreakResetQueue(habit.ID, 1, time.Now())
	entry2 := domain.NewStreakResetQueue(habit.ID, 1, time.Now())

	queueRepo.CreateQueueEntry(ctx, entry1)
	queueRepo.CreateQueueEntry(ctx, entry2)

	entries, err := streakResetService.GetUnprocessedQueueEntries(ctx)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 unprocessed entries, got %d", len(entries))
	}
}

func TestGetQueueEntry(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()
	queueRepo := newMockStreakResetQueueRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	streakResetService := NewStreakResetService(queueRepo, habitRepo, logRepo, reminderRepo, habitService)
	ctx := context.Background()

	habit, _ := habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)
	entry := domain.NewStreakResetQueue(habit.ID, 1, time.Now())
	created, _ := queueRepo.CreateQueueEntry(ctx, entry)

	retrieved, err := streakResetService.GetQueueEntry(ctx, created.ID)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if retrieved.ID != created.ID {
		t.Errorf("expected same queue entry")
	}
}

func TestProcessQueueEntry_PreservesPreviousStreak(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()
	queueRepo := newMockStreakResetQueueRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	streakResetService := NewStreakResetService(queueRepo, habitRepo, logRepo, reminderRepo, habitService)
	ctx := context.Background()

	habit, _ := habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)
	habit.CurrentStreak = 7
	habitRepo.UpdateHabit(ctx, habit)

	entry := domain.NewStreakResetQueue(habit.ID, 1, time.Now())
	queueRepo.CreateQueueEntry(ctx, entry)

	streakResetService.ProcessQueueEntries(ctx)

	processedEntry, _ := queueRepo.GetQueueEntryByID(ctx, entry.ID)
	if !processedEntry.PreviousStreak.Valid || processedEntry.PreviousStreak.Int64 != 7 {
		t.Errorf("expected previous_streak to be 7, got %d", processedEntry.PreviousStreak.Int64)
	}
}

func TestProcessQueueEntry_ReminderMarkedIncomplete(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()
	queueRepo := newMockStreakResetQueueRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	_ = NewReminderService(reminderRepo, habitRepo, habitService)
	streakResetService := NewStreakResetService(queueRepo, habitRepo, logRepo, reminderRepo, habitService)
	ctx := context.Background()

	habit, _ := habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)
	habit.CurrentStreak = 5
	habitRepo.UpdateHabit(ctx, habit)

	resetDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	reminder := domain.NewHabitReminder(habit.ID, 1, resetDate)
	reminder.MarkAsCompleted()
	reminderRepo.CreateReminder(ctx, reminder)

	entry := domain.NewStreakResetQueue(habit.ID, 1, resetDate)
	queueRepo.CreateQueueEntry(ctx, entry)

	streakResetService.ProcessQueueEntries(ctx)

	updated, _ := reminderRepo.GetReminderByID(ctx, reminder.ID)
	if updated.IsCompleted {
		t.Errorf("expected reminder to be marked as incomplete")
	}
}

func TestMultipleQueueEntries_AllProcessed(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()
	queueRepo := newMockStreakResetQueueRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	streakResetService := NewStreakResetService(queueRepo, habitRepo, logRepo, reminderRepo, habitService)
	ctx := context.Background()

	h1, _ := habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)
	h2, _ := habitService.CreateHabit(ctx, 1, "Reading", domain.FrequencyDaily)
	h3, _ := habitService.CreateHabit(ctx, 1, "Meditation", domain.FrequencyDaily)

	h1.CurrentStreak = 5
	h2.CurrentStreak = 10
	h3.CurrentStreak = 3

	habitRepo.UpdateHabit(ctx, h1)
	habitRepo.UpdateHabit(ctx, h2)
	habitRepo.UpdateHabit(ctx, h3)

	queueRepo.CreateQueueEntry(ctx, domain.NewStreakResetQueue(h1.ID, 1, time.Now()))
	queueRepo.CreateQueueEntry(ctx, domain.NewStreakResetQueue(h2.ID, 1, time.Now()))
	queueRepo.CreateQueueEntry(ctx, domain.NewStreakResetQueue(h3.ID, 1, time.Now()))

	streakResetService.ProcessQueueEntries(ctx)

	entries, _ := queueRepo.GetUnprocessedEntries(ctx)
	if len(entries) != 0 {
		t.Errorf("expected all entries to be processed")
	}

	updated1, _ := habitService.GetHabit(ctx, h1.ID)
	updated2, _ := habitService.GetHabit(ctx, h2.ID)
	updated3, _ := habitService.GetHabit(ctx, h3.ID)

	if updated1.CurrentStreak != 0 || updated2.CurrentStreak != 0 || updated3.CurrentStreak != 0 {
		t.Errorf("expected all streaks to be reset")
	}
}
