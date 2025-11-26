package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"HobitsService/internal/domain"
	"HobitsService/internal/repository"
)

// HabitService сервис для управления привычками
type HabitService struct {
	habitRepo    repository.HabitRepository
	logRepo      repository.HabitLogRepository
	reminderRepo repository.HabitReminderRepository
}

// NewHabitService создает новый HabitService
func NewHabitService(
	habitRepo repository.HabitRepository,
	logRepo repository.HabitLogRepository,
	reminderRepo repository.HabitReminderRepository,
) *HabitService {
	return &HabitService{
		habitRepo:    habitRepo,
		logRepo:      logRepo,
		reminderRepo: reminderRepo,
	}
}

// CreateHabit создает новую привычку для пользователя
func (s *HabitService) CreateHabit(ctx context.Context, userID int, name string, frequency domain.HabitFrequency) (*domain.Habit, error) {
	habit := domain.NewHabit(userID, name, frequency)
	return s.habitRepo.CreateHabit(ctx, habit)
}

// GetHabit получает привычку по ID
func (s *HabitService) GetHabit(ctx context.Context, habitID int) (*domain.Habit, error) {
	return s.habitRepo.GetHabitByID(ctx, habitID)
}

// GetUserHabits получает все привычки пользователя
func (s *HabitService) GetUserHabits(ctx context.Context, userID int) ([]*domain.Habit, error) {
	return s.habitRepo.GetHabitsByUserID(ctx, userID)
}

// GetActiveUserHabits получает активные привычки пользователя
func (s *HabitService) GetActiveUserHabits(ctx context.Context, userID int) ([]*domain.Habit, error) {
	return s.habitRepo.GetActiveHabitsByUserID(ctx, userID)
}

// GetAllActiveHabits получает все активные привычки
func (s *HabitService) GetAllActiveHabits(ctx context.Context) ([]*domain.Habit, error) {
	return s.habitRepo.GetAllActiveHabits(ctx)
}

// UpdateHabit обновляет привычку
func (s *HabitService) UpdateHabit(ctx context.Context, habit *domain.Habit) (*domain.Habit, error) {
	return s.habitRepo.UpdateHabit(ctx, habit)
}

// DeactivateHabit деактивирует привычку
func (s *HabitService) DeactivateHabit(ctx context.Context, habitID int) (*domain.Habit, error) {
	habit, err := s.habitRepo.GetHabitByID(ctx, habitID)
	if err != nil {
		return nil, err
	}
	habit.Deactivate()
	return s.habitRepo.UpdateHabit(ctx, habit)
}

// ActivateHabit активирует привычку
func (s *HabitService) ActivateHabit(ctx context.Context, habitID int) (*domain.Habit, error) {
	habit, err := s.habitRepo.GetHabitByID(ctx, habitID)
	if err != nil {
		return nil, err
	}
	habit.Activate()
	return s.habitRepo.UpdateHabit(ctx, habit)
}

// SetWeeklyDays устанавливает дни недели для еженедельной привычки
func (s *HabitService) SetWeeklyDays(ctx context.Context, habitID int, days []int) (*domain.Habit, error) {
	habit, err := s.habitRepo.GetHabitByID(ctx, habitID)
	if err != nil {
		return nil, err
	}

	if habit.Frequency != domain.FrequencyWeekly {
		return nil, fmt.Errorf("habit is not weekly")
	}

	// Преобразуем массив дней в строку "1,3,5"
	daysStr := s.daysToString(days)
	habit.SetWeeklyDays(daysStr)

	return s.habitRepo.UpdateHabit(ctx, habit)
}

// SetMonthlyDays устанавливает дни месяца для ежемесячной привычки
func (s *HabitService) SetMonthlyDays(ctx context.Context, habitID int, days []int) (*domain.Habit, error) {
	habit, err := s.habitRepo.GetHabitByID(ctx, habitID)
	if err != nil {
		return nil, err
	}

	if habit.Frequency != domain.FrequencyMonthly {
		return nil, fmt.Errorf("habit is not monthly")
	}

	// Преобразуем массив дней в строку "1,15,28"
	daysStr := s.daysToString(days)
	habit.SetMonthlyDays(daysStr)

	return s.habitRepo.UpdateHabit(ctx, habit)
}

// GetScheduledDaysForToday возвращает, нужно ли подтверждение сегодня
func (s *HabitService) GetScheduledDaysForToday(ctx context.Context, habitID int) (bool, error) {
	habit, err := s.habitRepo.GetHabitByID(ctx, habitID)
	if err != nil {
		return false, err
	}

	return s.isHabitScheduledForDate(habit, time.Now()), nil
}

// isHabitScheduledForDate проверяет, запланирована ли привычка на дату
func (s *HabitService) isHabitScheduledForDate(habit *domain.Habit, date time.Time) bool {
	switch habit.Frequency {
	case domain.FrequencyDaily:
		return true

	case domain.FrequencyWeekly:
		if !habit.WeeklyDays.Valid {
			return false
		}
		weekday := s.goWeekdayToInt(date.Weekday())
		return s.containsDay(habit.WeeklyDays.String, weekday)

	case domain.FrequencyMonthly:
		if !habit.MonthlyDays.Valid {
			return false
		}
		day := date.Day()
		return s.containsDay(habit.MonthlyDays.String, day)

	default:
		return false
	}
}

// GetScheduledDaysBetween возвращает все запланированные дни между двумя датами
func (s *HabitService) GetScheduledDaysBetween(ctx context.Context, habitID int, from, to time.Time) ([]time.Time, error) {
	habit, err := s.habitRepo.GetHabitByID(ctx, habitID)
	if err != nil {
		return nil, err
	}

	var scheduledDays []time.Time
	current := from

	for current.Before(to) || current.Equal(to) {
		if s.isHabitScheduledForDate(habit, current) {
			scheduledDays = append(scheduledDays, current)
		}
		current = current.AddDate(0, 0, 1)
	}

	return scheduledDays, nil
}

// daysToString преобразует массив дней в строку "1,3,5"
func (s *HabitService) daysToString(days []int) string {
	var strs []string
	for _, day := range days {
		strs = append(strs, strconv.Itoa(day))
	}
	return strings.Join(strs, ",")
}

// goWeekdayToInt преобразует Go weekday (0=Sunday) в интервал (1=Monday, 7=Sunday)
func (s *HabitService) goWeekdayToInt(wd time.Weekday) int {
	// Go: Sunday=0, Monday=1, ..., Saturday=6
	// Нам нужно: Monday=1, ..., Sunday=7
	if wd == 0 { // Sunday
		return 7
	}
	return int(wd)
}

// containsDay проверяет, содержит ли строка дней определенный день
func (s *HabitService) containsDay(daysStr string, day int) bool {
	parts := strings.Split(daysStr, ",")
	for _, part := range parts {
		if strings.TrimSpace(part) == strconv.Itoa(day) {
			return true
		}
	}
	return false
}
