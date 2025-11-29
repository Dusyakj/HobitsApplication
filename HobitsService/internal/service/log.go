package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"HobitsService/internal/domain"
	"HobitsService/internal/repository"
)

// LogService сервис для логирования выполнений привычек
type LogService struct {
	logRepo      repository.HabitLogRepository
	habitRepo    repository.HabitRepository
	reminderRepo repository.HabitReminderRepository
	queueRepo    repository.StreakResetQueueRepository
	habitService *HabitService
}

// NewLogService создает новый LogService
func NewLogService(
	logRepo repository.HabitLogRepository,
	habitRepo repository.HabitRepository,
	reminderRepo repository.HabitReminderRepository,
	queueRepo repository.StreakResetQueueRepository,
	habitService *HabitService,
) *LogService {
	return &LogService{
		logRepo:      logRepo,
		habitRepo:    habitRepo,
		reminderRepo: reminderRepo,
		queueRepo:    queueRepo,
		habitService: habitService,
	}
}

// LogCompletion логирует выполнение привычки и обновляет стрик
func (s *LogService) LogCompletion(ctx context.Context, habitID, userID int, comment string) (*domain.HabitLog, error) {
	// Получаем привычку
	habit, err := s.habitRepo.GetHabitByID(ctx, habitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get habit: %w", err)
	}

	if habit.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	// Проверяем, уже ли выполнена сегодня
	existingLog, err := s.logRepo.GetLogByHabitIDAndDate(ctx, habitID, todayDate)
	if err == nil && existingLog != nil {
		// Уже выполнена, возвращаем существующий логи
		return existingLog, nil
	}

	// Создаем новый лог
	log := domain.NewHabitLog(habitID, userID, todayDate, comment)
	createdLog, err := s.logRepo.CreateLog(ctx, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create log: %w", err)
	}

	// Обновляем статус напоминания
	reminder, err := s.reminderRepo.GetReminderByHabitIDAndDate(ctx, habitID, todayDate)
	if err == nil && reminder != nil {
		reminder.MarkAsCompleted()
		_, _ = s.reminderRepo.UpdateReminder(ctx, reminder)
	}

	// Обновляем стрик привычки
	if err := s.updateStreak(ctx, habit); err != nil {
		// Логируем ошибку но не прерываем основной процесс
		fmt.Printf("failed to update streak: %v\n", err)
	}

	// Удаляем из очереди сброса если была добавлена
	queueEntry, _ := s.queueRepo.GetQueueEntryByHabitIDAndDate(ctx, habitID, todayDate)
	if queueEntry != nil {
		_ = s.queueRepo.DeleteQueueEntry(ctx, queueEntry.ID)
	}

	return createdLog, nil
}

// updateStreak обновляет стрик привычки
func (s *LogService) updateStreak(ctx context.Context, habit *domain.Habit) error {
	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	// Если это первое выполнение
	if !habit.LastCompletedDate.Valid {
		habit.IncreaseStreak()
		habit.UpdateLastCompletedDate(todayDate)
		_, err := s.habitRepo.UpdateHabit(ctx, habit)

		return err
	}

	lastCompleted := habit.LastCompletedDate.Time
	lastCompletedDate := time.Date(lastCompleted.Year(), lastCompleted.Month(), lastCompleted.Day(), 0, 0, 0, 0, lastCompleted.Location())

	// Проверяем, нарушен ли стрик
	if s.isStreakBroken(habit, lastCompletedDate, todayDate) {
		// Стрик нарушен - сбрасываем
		if habit.CurrentStreak > habit.BestStreak {
			habit.BestStreak = habit.CurrentStreak
		}
		habit.CurrentStreak = 1
	} else {
		// Стрик продолжается
		habit.IncreaseStreak()
	}

	habit.UpdateLastCompletedDate(todayDate)
	_, err := s.habitRepo.UpdateHabit(ctx, habit)

	return err
}

// isStreakBroken проверяет, нарушен ли стрик между двумя датами
func (s *LogService) isStreakBroken(habit *domain.Habit, lastDate, today time.Time) bool {
	scheduledDays, _ := s.habitService.GetScheduledDaysBetween(context.Background(), habit.ID, lastDate.AddDate(0, 0, 1), today)

	// Если между последним выполнением и сегодня есть запланированные дни, которые не выполнены - стрик нарушен
	if len(scheduledDays) > 1 {
		return true
	}

	return false
}

// GetHabitLogs получает логи привычки
func (s *LogService) GetHabitLogs(ctx context.Context, habitID int) ([]*domain.HabitLog, error) {
	return s.logRepo.GetLogsByHabitID(ctx, habitID)
}

// GetHabitLogsByDateRange получает логи привычки за период
func (s *LogService) GetHabitLogsByDateRange(ctx context.Context, habitID int, from, to time.Time) ([]*domain.HabitLog, error) {
	return s.logRepo.GetLogsByHabitIDAndDate(ctx, habitID, from, to)
}

// GetCompletionRate получает процент выполнения за период
func (s *LogService) GetCompletionRate(ctx context.Context, habitID int, from, to time.Time) (float64, error) {
	scheduledDays, err := s.habitService.GetScheduledDaysBetween(ctx, habitID, from, to)
	if err != nil {
		return 0, err
	}

	if len(scheduledDays) == 0 {
		return 0, nil
	}

	count, err := s.logRepo.CountLogsByHabitIDAndDate(ctx, habitID, from, to)
	if err != nil {
		return 0, err
	}

	return float64(count) / float64(len(scheduledDays)) * 100, nil
}
