package scheduler

import (
	"context"
	"time"

	"go.uber.org/zap"

	"HobitsService/internal/logger"
	"HobitsService/internal/service"
)

// Scheduler запускает периодические задачи
type Scheduler struct {
	habitService       *service.HabitService
	logService         *service.LogService
	reminderService    *service.ReminderService
	streakResetService *service.StreakResetService
	userService        *service.UserService

	stopChan chan struct{}
}

// NewScheduler создает новый scheduler
func NewScheduler(
	habitService *service.HabitService,
	logService *service.LogService,
	reminderService *service.ReminderService,
	streakResetService *service.StreakResetService,
	userService *service.UserService,
) *Scheduler {
	return &Scheduler{
		habitService:       habitService,
		logService:         logService,
		reminderService:    reminderService,
		streakResetService: streakResetService,
		userService:        userService,
		stopChan:           make(chan struct{}),
	}
}

// Start запускает scheduler с периодическими задачами
func (s *Scheduler) Start() {
	logger.Info("Scheduler started")

	ctx := context.Background()

	// Задача 1: Генерация напоминаний каждый день в 08:00
	go s.scheduleReminders(ctx)

	// Задача 2: Проверка и сброс стриков каждый день в 23:59
	go s.scheduleStreakCheck(ctx)

	// Задача 3: Обработка очереди сброса стриков каждый день в 00:30
	go s.processStreakResetQueue(ctx)
}

// Stop останавливает scheduler
func (s *Scheduler) Stop() {
	logger.Info("Scheduler stopping")
	close(s.stopChan)
}

// scheduleReminders генерирует напоминания в 08:00 каждый день
// Это нужно запустить для каждого пользователя
func (s *Scheduler) scheduleReminders(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Проверяем каждый час
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			logger.Info("Reminders scheduler stopped")
			return
		case <-ticker.C:
			now := time.Now()

			// Проверяем, 08:00 ли сейчас
			if now.Hour() == 8 && now.Minute() < 60 {
				logger.Info("Generating reminders for all users")

				// Получаем всех пользователей
				users, err := s.userService.GetAllUsers(ctx)
				if err != nil {
					logger.Error("Failed to get all users for reminder generation", zap.Error(err))
					continue
				}

				// Для каждого пользователя генерируем напоминания
				for _, user := range users {
					_, err := s.reminderService.GenerateRemindersForToday(ctx, user.ID)
					if err != nil {
						logger.Error("Failed to generate reminders for user", zap.Error(err), zap.Int("user_id", user.ID))
					}
				}

				logger.Info("Reminders generated for all users", zap.Int("count", len(users)))
			}
		}
	}
}

// scheduleStreakCheck проверяет и добавляет стрики в очередь на сброс в 23:59
func (s *Scheduler) scheduleStreakCheck(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Проверяем каждый час
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			logger.Info("Streak check scheduler stopped")
			return
		case <-ticker.C:
			now := time.Now()

			// Проверяем, 23:59 ли сейчас
			if now.Hour() == 23 && now.Minute() >= 55 {
				logger.Info("Checking streaks and queuing for reset")

				// Получаем все активные привычки
				habits, err := s.habitService.GetAllActiveHabits(ctx)
				if err != nil {
					logger.Error("Failed to get all active habits for streak check", zap.Error(err))
					continue
				}

				// Проверяем каждую привычку
				for _, habit := range habits {
					if err := s.streakResetService.CheckHabitStreak(ctx, habit.ID); err != nil {
						logger.Error("Failed to check streak for habit", zap.Error(err), zap.Int("habit_id", habit.ID))
					}
				}

				logger.Info("Streak check completed", zap.Int("habits_checked", len(habits)))
			}
		}
	}
}

// processStreakResetQueue обрабатывает очередь на сброс в 00:30
func (s *Scheduler) processStreakResetQueue(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Проверяем каждый час
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			logger.Info("Streak reset queue processor stopped")
			return
		case <-ticker.C:
			now := time.Now()

			// Проверяем, 00:30 ли сейчас (+ небольшой диапазон)
			if now.Hour() == 0 && now.Minute() >= 25 && now.Minute() <= 35 {
				logger.Info("Processing streak reset queue")

				if err := s.streakResetService.ProcessQueueEntries(ctx); err != nil {
					logger.Error("failed to process streak reset queue", zap.Error(err))
				}
			}
		}
	}
}

// ScheduleRemindersForUser генерирует напоминания для конкретного пользователя
func (s *Scheduler) ScheduleRemindersForUser(ctx context.Context, userID int) error {
	_, err := s.reminderService.GenerateRemindersForToday(ctx, userID)
	return err
}

// CheckHabitStreak проверяет стрик для конкретной привычки
func (s *Scheduler) CheckHabitStreak(ctx context.Context, habitID int) error {
	return s.streakResetService.CheckHabitStreak(ctx, habitID)
}
