package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"HobitsService/internal/domain"
)

// Mock repositories for testing
type mockHabitRepository struct {
	habits map[int]*domain.Habit
}

func (m *mockHabitRepository) CreateHabit(ctx context.Context, habit *domain.Habit) (*domain.Habit, error) {
	habit.ID = len(m.habits) + 1
	m.habits[habit.ID] = habit
	return habit, nil
}

func (m *mockHabitRepository) GetHabitByID(ctx context.Context, id int) (*domain.Habit, error) {
	if h, ok := m.habits[id]; ok {
		return h, nil
	}
	return nil, nil
}

func (m *mockHabitRepository) GetHabitsByUserID(ctx context.Context, userID int) ([]*domain.Habit, error) {
	var habits []*domain.Habit
	for _, h := range m.habits {
		if h.UserID == userID {
			habits = append(habits, h)
		}
	}
	return habits, nil
}

func (m *mockHabitRepository) GetActiveHabitsByUserID(ctx context.Context, userID int) ([]*domain.Habit, error) {
	var habits []*domain.Habit
	for _, h := range m.habits {
		if h.UserID == userID && h.IsActive {
			habits = append(habits, h)
		}
	}
	return habits, nil
}

func (m *mockHabitRepository) GetAllActiveHabits(ctx context.Context) ([]*domain.Habit, error) {
	var habits []*domain.Habit
	for _, h := range m.habits {
		if h.IsActive {
			habits = append(habits, h)
		}
	}
	return habits, nil
}

func (m *mockHabitRepository) UpdateHabit(ctx context.Context, habit *domain.Habit) (*domain.Habit, error) {
	m.habits[habit.ID] = habit
	return habit, nil
}

func (m *mockHabitRepository) DeleteHabit(ctx context.Context, id int) error {
	delete(m.habits, id)
	return nil
}

func (m *mockHabitRepository) GetHabitByUserIDAndName(ctx context.Context, userID int, name string) (*domain.Habit, error) {
	for _, h := range m.habits {
		if h.UserID == userID && h.Name == name {
			return h, nil
		}
	}
	return nil, nil
}

type mockHabitLogRepository struct {
	logs map[int]*domain.HabitLog
}

func (m *mockHabitLogRepository) CreateLog(ctx context.Context, log *domain.HabitLog) (*domain.HabitLog, error) {
	log.ID = len(m.logs) + 1
	m.logs[log.ID] = log
	return log, nil
}

func (m *mockHabitLogRepository) GetLogByID(ctx context.Context, id int) (*domain.HabitLog, error) {
	if l, ok := m.logs[id]; ok {
		return l, nil
	}
	return nil, nil
}

func (m *mockHabitLogRepository) GetLogsByHabitID(ctx context.Context, habitID int) ([]*domain.HabitLog, error) {
	var logs []*domain.HabitLog
	for _, l := range m.logs {
		if l.HabitID == habitID {
			logs = append(logs, l)
		}
	}
	return logs, nil
}

func (m *mockHabitLogRepository) GetLogsByHabitIDAndDate(ctx context.Context, habitID int, from, to time.Time) ([]*domain.HabitLog, error) {
	var logs []*domain.HabitLog
	for _, l := range m.logs {
		if l.HabitID == habitID && l.LoggedDate.After(from) && l.LoggedDate.Before(to) {
			logs = append(logs, l)
		}
	}
	return logs, nil
}

func (m *mockHabitLogRepository) GetLogByHabitIDAndDate(ctx context.Context, habitID int, date time.Time) (*domain.HabitLog, error) {
	for _, l := range m.logs {
		if l.HabitID == habitID && l.LoggedDate.Equal(date) {
			return l, nil
		}
	}
	return nil, nil
}

func (m *mockHabitLogRepository) DeleteLog(ctx context.Context, id int) error {
	delete(m.logs, id)
	return nil
}

func (m *mockHabitLogRepository) CountLogsByHabitIDAndDate(ctx context.Context, habitID int, from, to time.Time) (int, error) {
	count := 0
	for _, l := range m.logs {
		if l.HabitID == habitID && l.LoggedDate.After(from) && l.LoggedDate.Before(to) {
			count++
		}
	}
	return count, nil
}

type mockHabitReminderRepository struct {
	reminders map[int]*domain.HabitReminder
}

func (m *mockHabitReminderRepository) CreateReminder(ctx context.Context, reminder *domain.HabitReminder) (*domain.HabitReminder, error) {
	reminder.ID = len(m.reminders) + 1
	m.reminders[reminder.ID] = reminder
	return reminder, nil
}

func (m *mockHabitReminderRepository) GetReminderByID(ctx context.Context, id int) (*domain.HabitReminder, error) {
	if r, ok := m.reminders[id]; ok {
		return r, nil
	}
	return nil, nil
}

func (m *mockHabitReminderRepository) GetRemindersByUserID(ctx context.Context, userID int) ([]*domain.HabitReminder, error) {
	var reminders []*domain.HabitReminder
	for _, r := range m.reminders {
		if r.UserID == userID {
			reminders = append(reminders, r)
		}
	}
	return reminders, nil
}

func (m *mockHabitReminderRepository) GetRemindersByDate(ctx context.Context, date time.Time) ([]*domain.HabitReminder, error) {
	var reminders []*domain.HabitReminder
	for _, r := range m.reminders {
		if r.ReminderDate.Valid && r.ReminderDate.Time.Equal(date) {
			reminders = append(reminders, r)
		}
	}
	return reminders, nil
}

func (m *mockHabitReminderRepository) GetRemindersByUserIDAndDate(ctx context.Context, userID int, date time.Time) ([]*domain.HabitReminder, error) {
	var reminders []*domain.HabitReminder
	for _, r := range m.reminders {
		if r.UserID == userID && r.ReminderDate.Valid && r.ReminderDate.Time.Equal(date) {
			reminders = append(reminders, r)
		}
	}
	return reminders, nil
}

func (m *mockHabitReminderRepository) GetReminderByHabitIDAndDate(ctx context.Context, habitID int, date time.Time) (*domain.HabitReminder, error) {
	for _, r := range m.reminders {
		if r.HabitID == habitID && r.ReminderDate.Valid && r.ReminderDate.Time.Equal(date) {
			return r, nil
		}
	}
	return nil, nil
}

func (m *mockHabitReminderRepository) UpdateReminder(ctx context.Context, reminder *domain.HabitReminder) (*domain.HabitReminder, error) {
	m.reminders[reminder.ID] = reminder
	return reminder, nil
}

func (m *mockHabitReminderRepository) DeleteReminder(ctx context.Context, id int) error {
	delete(m.reminders, id)
	return nil
}

func newMockHabitRepository() *mockHabitRepository {
	return &mockHabitRepository{
		habits: make(map[int]*domain.Habit),
	}
}

func newMockHabitLogRepository() *mockHabitLogRepository {
	return &mockHabitLogRepository{
		logs: make(map[int]*domain.HabitLog),
	}
}

func newMockHabitReminderRepository() *mockHabitReminderRepository {
	return &mockHabitReminderRepository{
		reminders: make(map[int]*domain.HabitReminder),
	}
}

func TestCreateHabit(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	service := NewHabitService(habitRepo, logRepo, reminderRepo)
	ctx := context.Background()

	habit, err := service.CreateHabit(ctx, 1, "Exercise", domain.FrequencyDaily)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if habit.Name != "Exercise" {
		t.Errorf("expected name Exercise, got %s", habit.Name)
	}
	if habit.UserID != 1 {
		t.Errorf("expected user_id 1, got %d", habit.UserID)
	}
}

func TestGetHabit(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	service := NewHabitService(habitRepo, logRepo, reminderRepo)
	ctx := context.Background()

	created, _ := service.CreateHabit(ctx, 1, "Test", domain.FrequencyDaily)
	retrieved, err := service.GetHabit(ctx, created.ID)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if retrieved.ID != created.ID {
		t.Errorf("expected id %d, got %d", created.ID, retrieved.ID)
	}
}

func TestGetUserHabits(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	service := NewHabitService(habitRepo, logRepo, reminderRepo)
	ctx := context.Background()

	service.CreateHabit(ctx, 1, "Habit1", domain.FrequencyDaily)
	service.CreateHabit(ctx, 1, "Habit2", domain.FrequencyWeekly)
	service.CreateHabit(ctx, 2, "Habit3", domain.FrequencyDaily)

	habits, err := service.GetUserHabits(ctx, 1)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(habits) != 2 {
		t.Errorf("expected 2 habits, got %d", len(habits))
	}
}

func TestGetActiveUserHabits(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	service := NewHabitService(habitRepo, logRepo, reminderRepo)
	ctx := context.Background()

	_, _ = service.CreateHabit(ctx, 1, "Habit1", domain.FrequencyDaily)
	h2, _ := service.CreateHabit(ctx, 1, "Habit2", domain.FrequencyWeekly)

	service.DeactivateHabit(ctx, h2.ID)

	habits, err := service.GetActiveUserHabits(ctx, 1)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(habits) != 1 {
		t.Errorf("expected 1 active habit, got %d", len(habits))
	}
}

func TestDeactivateAndActivateHabit(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	service := NewHabitService(habitRepo, logRepo, reminderRepo)
	ctx := context.Background()

	habit, _ := service.CreateHabit(ctx, 1, "Test", domain.FrequencyDaily)

	if !habit.IsActive {
		t.Errorf("expected new habit to be active")
	}

	deactivated, err := service.DeactivateHabit(ctx, habit.ID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if deactivated.IsActive {
		t.Errorf("expected habit to be inactive")
	}

	activated, err := service.ActivateHabit(ctx, habit.ID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !activated.IsActive {
		t.Errorf("expected habit to be active")
	}
}

func TestSetWeeklyDays(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	service := NewHabitService(habitRepo, logRepo, reminderRepo)
	ctx := context.Background()

	habit, _ := service.CreateHabit(ctx, 1, "Test", domain.FrequencyWeekly)
	updated, err := service.SetWeeklyDays(ctx, habit.ID, []int{1, 3, 5})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !updated.WeeklyDays.Valid {
		t.Errorf("expected weekly_days to be set")
	}
}

func TestIsHabitScheduledForDate(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	service := NewHabitService(habitRepo, logRepo, reminderRepo)

	// Daily habit should be scheduled every day
	daily := &domain.Habit{
		ID:        1,
		Frequency: domain.FrequencyDaily,
	}
	if !service.isHabitScheduledForDate(daily, time.Now()) {
		t.Errorf("expected daily habit to be scheduled")
	}

	// Weekly habit with specific days
	weekly := &domain.Habit{
		ID:        2,
		Frequency: domain.FrequencyWeekly,
		WeeklyDays: sql.NullString{
			Valid:  true,
			String: "1,3,5",
		},
	}

	// Monday should be scheduled (1)
	monday := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // Monday
	if !service.isHabitScheduledForDate(weekly, monday) {
		t.Errorf("expected Monday to be scheduled for weekly habit with days 1,3,5")
	}

	// Sunday should not be scheduled unless 7 is in the list
	sunday := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC) // Sunday
	if service.isHabitScheduledForDate(weekly, sunday) {
		t.Errorf("expected Sunday to not be scheduled for weekly habit without day 7")
	}
}

func TestGetScheduledDaysBetween(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	service := NewHabitService(habitRepo, logRepo, reminderRepo)
	ctx := context.Background()

	daily, _ := service.CreateHabit(ctx, 1, "Daily", domain.FrequencyDaily)
	habitRepo.UpdateHabit(ctx, daily)

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

	scheduledDays, err := service.GetScheduledDaysBetween(ctx, daily.ID, from, to)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(scheduledDays) != 5 {
		t.Errorf("expected 5 scheduled days, got %d", len(scheduledDays))
	}
}

func TestGetAllActiveHabits(t *testing.T) {
	habitRepo := newMockHabitRepository()
	logRepo := newMockHabitLogRepository()
	reminderRepo := newMockHabitReminderRepository()

	service := NewHabitService(habitRepo, logRepo, reminderRepo)
	ctx := context.Background()

	_, _ = service.CreateHabit(ctx, 1, "Habit1", domain.FrequencyDaily)
	_, _ = service.CreateHabit(ctx, 2, "Habit2", domain.FrequencyDaily)
	h3, _ := service.CreateHabit(ctx, 1, "Habit3", domain.FrequencyDaily)

	service.DeactivateHabit(ctx, h3.ID)

	habits, err := service.GetAllActiveHabits(ctx)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(habits) != 2 {
		t.Errorf("expected 2 active habits, got %d", len(habits))
	}
}
