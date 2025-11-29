package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"HobitsService/internal/domain"
	"HobitsService/internal/repository"
)

// ReminderService сервис для управления напоминаниями
type ReminderService struct {
	reminderRepo repository.HabitReminderRepository
	habitRepo    repository.HabitRepository
	habitService *HabitService
}

// NewReminderService создает новый ReminderService
func NewReminderService(
	reminderRepo repository.HabitReminderRepository,
	habitRepo repository.HabitRepository,
	habitService *HabitService,
) *ReminderService {
	return &ReminderService{
		reminderRepo: reminderRepo,
		habitRepo:    habitRepo,
		habitService: habitService,
	}
}

// CreateReminder создает напоминание
func (s *ReminderService) CreateReminder(ctx context.Context, habitID, userID int, reminderDate time.Time) (*domain.HabitReminder, error) {
	reminder := domain.NewHabitReminder(habitID, userID, reminderDate)
	return s.reminderRepo.CreateReminder(ctx, reminder)
}

// GenerateRemindersForToday генерирует напоминания на сегодня для пользователя
func (s *ReminderService) GenerateRemindersForToday(ctx context.Context, userID int) ([]*domain.HabitReminder, error) {
	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	// Получаем существующие напоминания на сегодня
	existingReminders, err := s.reminderRepo.GetRemindersByUserIDAndDate(ctx, userID, todayDate)
	if err != nil {
		existingReminders = []*domain.HabitReminder{}
	}

	// Создаем map существующих напоминаний для быстрой проверки
	existingHabitIDs := make(map[int]bool)
	for _, reminder := range existingReminders {
		existingHabitIDs[reminder.HabitID] = true
	}

	// Получаем все активные привычки пользователя
	habits, err := s.habitRepo.GetActiveHabitsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get habits: %w", err)
	}

	var allReminders []*domain.HabitReminder
	allReminders = append(allReminders, existingReminders...)

	for _, habit := range habits {
		// Пропускаем если напоминание уже существует
		if existingHabitIDs[habit.ID] {
			continue
		}

		// Проверяем, нужно ли подтверждение сегодня
		if s.habitService.isHabitScheduledForDate(habit, todayDate) {
			reminder := domain.NewHabitReminder(habit.ID, userID, todayDate)
			created, err := s.reminderRepo.CreateReminder(ctx, reminder)
			if err != nil {
				fmt.Printf("failed to create reminder for habit %d: %v\n", habit.ID, err)
				continue
			}
			allReminders = append(allReminders, created)
		}
	}

	return allReminders, nil
}

// GetRemindersByDate получает напоминания на дату
func (s *ReminderService) GetRemindersByDate(ctx context.Context, date time.Time) ([]*domain.HabitReminder, error) {
	return s.reminderRepo.GetRemindersByDate(ctx, date)
}

// GetRemindersByUserAndDate получает напоминания пользователя на дату
func (s *ReminderService) GetRemindersByUserAndDate(ctx context.Context, userID int, date time.Time) ([]*domain.HabitReminder, error) {
	return s.reminderRepo.GetRemindersByUserIDAndDate(ctx, userID, date)
}

// MarkReminderAsCompleted отмечает напоминание как выполненное
func (s *ReminderService) MarkReminderAsCompleted(ctx context.Context, reminderID int) (*domain.HabitReminder, error) {
	reminder, err := s.reminderRepo.GetReminderByID(ctx, reminderID)
	if err != nil {
		return nil, err
	}

	reminder.MarkAsCompleted()
	return s.reminderRepo.UpdateReminder(ctx, reminder)
}

// MarkReminderAsIncomplete отмечает напоминание как невыполненное
func (s *ReminderService) MarkReminderAsIncomplete(ctx context.Context, reminderID int) (*domain.HabitReminder, error) {
	reminder, err := s.reminderRepo.GetReminderByID(ctx, reminderID)
	if err != nil {
		return nil, err
	}

	reminder.MarkAsIncomplete()
	return s.reminderRepo.UpdateReminder(ctx, reminder)
}

// GetUnconfirmedRemindersByDate получает неподтвержденные напоминания на дату
func (s *ReminderService) GetUnconfirmedRemindersByDate(ctx context.Context, date time.Time) ([]*domain.HabitReminder, error) {
	allReminders, err := s.reminderRepo.GetRemindersByDate(ctx, date)
	if err != nil {
		return nil, err
	}

	var unconfirmed []*domain.HabitReminder
	for _, reminder := range allReminders {
		if !reminder.IsCompleted {
			unconfirmed = append(unconfirmed, reminder)
		}
	}

	return unconfirmed, nil
}

// CountUnconfirmedRemindersByUserAndDate подсчитывает неподтвержденные напоминания пользователя на дату
func (s *ReminderService) CountUnconfirmedRemindersByUserAndDate(ctx context.Context, userID int, date time.Time) (int, error) {
	reminders, err := s.reminderRepo.GetRemindersByUserIDAndDate(ctx, userID, date)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, reminder := range reminders {
		if !reminder.IsCompleted {
			count++
		}
	}

	return count, nil
}

// CalculateNextReminderDate вычисляет следующую дату напоминания для привычки
func (s *ReminderService) CalculateNextReminderDate(ctx context.Context, habit *domain.Habit, currentDate time.Time) (time.Time, error) {
	switch habit.Frequency {
	case domain.FrequencyDaily:
		// Для ежедневных - следующий день
		return currentDate.AddDate(0, 0, 1), nil

	case domain.FrequencyWeekly:
		// Для еженедельных - следующий запланированный день недели
		return s.calculateNextWeeklyDate(habit, currentDate)

	case domain.FrequencyMonthly:
		// Для ежемесячных - следующий запланированный день месяца
		return s.calculateNextMonthlyDate(habit, currentDate)

	default:
		return time.Time{}, fmt.Errorf("unknown frequency: %s", habit.Frequency)
	}
}

// calculateNextWeeklyDate вычисляет следующую дату для weekly привычки
func (s *ReminderService) calculateNextWeeklyDate(habit *domain.Habit, currentDate time.Time) (time.Time, error) {
	if !habit.WeeklyDays.Valid || habit.WeeklyDays.String == "" {
		return time.Time{}, fmt.Errorf("weekly days not set for habit %d", habit.ID)
	}

	// Парсим дни недели (формат: "1,3,5")
	daysStr := strings.Split(habit.WeeklyDays.String, ",")
	var days []int
	for _, dayStr := range daysStr {
		day, err := strconv.Atoi(strings.TrimSpace(dayStr))
		if err != nil {
			continue
		}
		days = append(days, day)
	}

	if len(days) == 0 {
		return time.Time{}, fmt.Errorf("no valid weekly days for habit %d", habit.ID)
	}

	// Конвертируем текущий день недели в ISO формат (1 = Понедельник, 7 = Воскресенье)
	currentWeekday := int(currentDate.Weekday())
	if currentWeekday == 0 {
		currentWeekday = 7 // Воскресенье
	}

	// Ищем следующий запланированный день
	for i := 1; i <= 7; i++ {
		nextDate := currentDate.AddDate(0, 0, i)
		nextWeekday := int(nextDate.Weekday())
		if nextWeekday == 0 {
			nextWeekday = 7
		}

		for _, day := range days {
			if day == nextWeekday {
				return time.Date(nextDate.Year(), nextDate.Month(), nextDate.Day(), 0, 0, 0, 0, nextDate.Location()), nil
			}
		}
	}

	// Если не нашли (не должно произойти), возвращаем завтра
	return currentDate.AddDate(0, 0, 1), nil
}

// calculateNextMonthlyDate вычисляет следующую дату для monthly привычки
func (s *ReminderService) calculateNextMonthlyDate(habit *domain.Habit, currentDate time.Time) (time.Time, error) {
	if !habit.MonthlyDays.Valid || habit.MonthlyDays.String == "" {
		return time.Time{}, fmt.Errorf("monthly days not set for habit %d", habit.ID)
	}

	// Парсим дни месяца (формат: "1,15,28")
	daysStr := strings.Split(habit.MonthlyDays.String, ",")
	var days []int
	for _, dayStr := range daysStr {
		day, err := strconv.Atoi(strings.TrimSpace(dayStr))
		if err != nil {
			continue
		}
		days = append(days, day)
	}

	if len(days) == 0 {
		return time.Time{}, fmt.Errorf("no valid monthly days for habit %d", habit.ID)
	}

	currentDay := currentDate.Day()

	// Ищем следующий день в текущем месяце
	for _, day := range days {
		if day > currentDay {
			nextDate := time.Date(currentDate.Year(), currentDate.Month(), day, 0, 0, 0, 0, currentDate.Location())
			// Проверяем, что дата валидна (например, 31 число может не существовать в некоторых месяцах)
			if nextDate.Month() == currentDate.Month() {
				return nextDate, nil
			}
		}
	}

	// Если не нашли в текущем месяце, ищем в следующем
	nextMonth := currentDate.AddDate(0, 1, 0)
	nextMonth = time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, nextMonth.Location())

	for _, day := range days {
		nextDate := time.Date(nextMonth.Year(), nextMonth.Month(), day, 0, 0, 0, 0, nextMonth.Location())
		if nextDate.Month() == nextMonth.Month() {
			return nextDate, nil
		}
	}

	// В крайнем случае возвращаем первый день следующего месяца
	return nextMonth, nil
}

// CalculateInitialReminderDate вычисляет первую дату напоминания при создании привычки
func (s *ReminderService) CalculateInitialReminderDate(habit *domain.Habit) (time.Time, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	switch habit.Frequency {
	case domain.FrequencyDaily:
		// Для ежедневных - сегодня
		return today, nil

	case domain.FrequencyWeekly:
		// Проверяем, есть ли сегодня в запланированных днях
		if s.isWeeklyScheduledForToday(habit, today) {
			return today, nil
		}
		// Иначе - следующий запланированный день
		return s.calculateNextWeeklyDate(habit, today.AddDate(0, 0, -1))

	case domain.FrequencyMonthly:
		// Проверяем, есть ли сегодня в запланированных днях
		if s.isMonthlyScheduledForToday(habit, today) {
			return today, nil
		}
		// Иначе - следующий запланированный день
		return s.calculateNextMonthlyDate(habit, today.AddDate(0, 0, -1))

	default:
		return time.Time{}, fmt.Errorf("unknown frequency: %s", habit.Frequency)
	}
}

// isWeeklyScheduledForToday проверяет, запланирована ли weekly привычка на сегодня
func (s *ReminderService) isWeeklyScheduledForToday(habit *domain.Habit, date time.Time) bool {
	if !habit.WeeklyDays.Valid || habit.WeeklyDays.String == "" {
		return false
	}

	currentWeekday := int(date.Weekday())
	if currentWeekday == 0 {
		currentWeekday = 7
	}

	daysStr := strings.Split(habit.WeeklyDays.String, ",")
	for _, dayStr := range daysStr {
		day, err := strconv.Atoi(strings.TrimSpace(dayStr))
		if err != nil {
			continue
		}
		if day == currentWeekday {
			return true
		}
	}

	return false
}

// isMonthlyScheduledForToday проверяет, запланирована ли monthly привычка на сегодня
func (s *ReminderService) isMonthlyScheduledForToday(habit *domain.Habit, date time.Time) bool {
	if !habit.MonthlyDays.Valid || habit.MonthlyDays.String == "" {
		return false
	}

	currentDay := date.Day()

	daysStr := strings.Split(habit.MonthlyDays.String, ",")
	for _, dayStr := range daysStr {
		day, err := strconv.Atoi(strings.TrimSpace(dayStr))
		if err != nil {
			continue
		}
		if day == currentDay {
			return true
		}
	}

	return false
}

// CreateNextReminder создает следующее напоминание для привычки
func (s *ReminderService) CreateNextReminder(ctx context.Context, habit *domain.Habit, currentDate time.Time) (*domain.HabitReminder, error) {
	nextDate, err := s.CalculateNextReminderDate(ctx, habit, currentDate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate next reminder date: %w", err)
	}

	reminder := &domain.HabitReminder{
		HabitID:      habit.ID,
		UserID:       habit.UserID,
		ReminderDate: sql.NullTime{Time: nextDate, Valid: true},
		IsCompleted:  false,
		SentAt:       time.Now(),
	}

	return s.reminderRepo.CreateReminder(ctx, reminder)
}
