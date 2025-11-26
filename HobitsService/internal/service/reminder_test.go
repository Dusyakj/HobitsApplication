package service

import (
	"context"
	"testing"
	"time"

	"HobitsService/internal/domain"
)

func TestGenerateRemindersForToday(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	reminderService := NewReminderService(reminderRepo, habitRepo, habitService)
	ctx := context.Background()

	_, _ = habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)

	reminders, err := reminderService.GenerateRemindersForToday(ctx, 1)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if reminders == nil {
		t.Errorf("expected reminders to be returned")
	}
	if len(reminders) == 0 {
		t.Errorf("expected at least one reminder for daily habit")
	}
}

func TestGenerateRemindersForToday_NoActiveHabits(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	reminderService := NewReminderService(reminderRepo, habitRepo, habitService)
	ctx := context.Background()

	reminders, err := reminderService.GenerateRemindersForToday(ctx, 999)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(reminders) != 0 {
		t.Errorf("expected no reminders for user with no habits")
	}
}

func TestGenerateRemindersForToday_WeeklyHabit(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	reminderService := NewReminderService(reminderRepo, habitRepo, habitService)
	ctx := context.Background()

	habit, _ := habitService.CreateHabit(ctx, 1, "Reading", domain.FrequencyWeekly)
	_ = habit

	// Set to include today's weekday
	today := time.Now()
	weekday := int(today.Weekday())
	if weekday == 0 {
		weekday = 7 // Convert Sunday from 0 to 7
	}

	habitService.SetWeeklyDays(ctx, habit.ID, []int{weekday})

	reminders, err := reminderService.GenerateRemindersForToday(ctx, 1)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(reminders) == 0 {
		t.Errorf("expected reminder for weekly habit on scheduled day")
	}
}

func TestGetRemindersByUserAndDate(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	reminderService := NewReminderService(reminderRepo, habitRepo, habitService)
	ctx := context.Background()

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	_, _ = habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)
	reminderService.GenerateRemindersForToday(ctx, 1)

	reminders, err := reminderService.GetRemindersByUserAndDate(ctx, 1, todayDate)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if reminders == nil {
		t.Errorf("expected reminders slice")
	}
}

func TestMarkReminderAsCompleted(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	reminderService := NewReminderService(reminderRepo, habitRepo, habitService)
	ctx := context.Background()

	_, _ = habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)
	reminders, _ := reminderService.GenerateRemindersForToday(ctx, 1)

	if len(reminders) == 0 {
		t.Errorf("expected at least one reminder")
		return
	}

	reminder := reminders[0]
	updated, err := reminderService.MarkReminderAsCompleted(ctx, reminder.ID)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !updated.IsCompleted {
		t.Errorf("expected reminder to be marked as completed")
	}
}

func TestMarkReminderAsIncomplete(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	reminderService := NewReminderService(reminderRepo, habitRepo, habitService)
	ctx := context.Background()

	_, _ = habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)
	reminders, _ := reminderService.GenerateRemindersForToday(ctx, 1)

	if len(reminders) == 0 {
		t.Errorf("expected at least one reminder")
		return
	}

	reminder := reminders[0]
	reminderService.MarkReminderAsCompleted(ctx, reminder.ID)

	updated, err := reminderService.MarkReminderAsIncomplete(ctx, reminder.ID)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if updated.IsCompleted {
		t.Errorf("expected reminder to be marked as incomplete")
	}
}

func TestGenerateRemindersForToday_MultipleHabits(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	reminderService := NewReminderService(reminderRepo, habitRepo, habitService)
	ctx := context.Background()

	habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)
	habitService.CreateHabit(ctx, 1, "Reading", domain.FrequencyDaily)
	habitService.CreateHabit(ctx, 1, "Meditation", domain.FrequencyDaily)

	reminders, err := reminderService.GenerateRemindersForToday(ctx, 1)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(reminders) != 3 {
		t.Errorf("expected 3 reminders, got %d", len(reminders))
	}
}

func TestGenerateRemindersForToday_DeactivatedHabits(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	reminderService := NewReminderService(reminderRepo, habitRepo, habitService)
	ctx := context.Background()

	_, _ = habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)
	h2, _ := habitService.CreateHabit(ctx, 1, "Reading", domain.FrequencyDaily)

	habitService.DeactivateHabit(ctx, h2.ID)

	reminders, err := reminderService.GenerateRemindersForToday(ctx, 1)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(reminders) != 1 {
		t.Errorf("expected 1 reminder for active habit, got %d", len(reminders))
	}
}

func TestReminderToggleCompletion(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	reminderService := NewReminderService(reminderRepo, habitRepo, habitService)
	ctx := context.Background()

	_, _ = habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)
	reminders, _ := reminderService.GenerateRemindersForToday(ctx, 1)

	if len(reminders) == 0 {
		t.Errorf("expected at least one reminder")
		return
	}

	reminder := reminders[0]

	// Mark as completed
	completed, _ := reminderService.MarkReminderAsCompleted(ctx, reminder.ID)
	if !completed.IsCompleted {
		t.Errorf("expected reminder to be completed")
	}

	// Mark as incomplete
	incomplete, _ := reminderService.MarkReminderAsIncomplete(ctx, reminder.ID)
	if incomplete.IsCompleted {
		t.Errorf("expected reminder to be incomplete")
	}

	// Mark as completed again
	completedAgain, _ := reminderService.MarkReminderAsCompleted(ctx, reminder.ID)
	if !completedAgain.IsCompleted {
		t.Errorf("expected reminder to be completed again")
	}
}

func TestGetRemindersByDate(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	habitService := NewHabitService(habitRepo, logRepo, reminderRepo)
	reminderService := NewReminderService(reminderRepo, habitRepo, habitService)
	ctx := context.Background()

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	habitService.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)
	reminderService.GenerateRemindersForToday(ctx, 1)

	reminders, err := reminderService.GetRemindersByDate(ctx, todayDate)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(reminders) == 0 {
		t.Errorf("expected reminders for today")
	}
}
