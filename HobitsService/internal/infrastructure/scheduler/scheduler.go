package scheduler

import (
	"context"
	"time"

	"HobitsService/internal/infrastructure/queue"
	"HobitsService/internal/logger"
	"HobitsService/internal/service"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Scheduler запускает периодические задачи с использованием cron
type Scheduler struct {
	habitService       *service.HabitService
	logService         *service.LogService
	reminderService    *service.ReminderService
	streakResetService *service.StreakResetService
	userService        *service.UserService
	rabbitMQ           *queue.RabbitMQClient

	cron *cron.Cron
}

// NewScheduler создает новый scheduler
func NewScheduler(
	habitService *service.HabitService,
	logService *service.LogService,
	reminderService *service.ReminderService,
	streakResetService *service.StreakResetService,
	userService *service.UserService,
	rabbitMQ *queue.RabbitMQClient,
) *Scheduler {
	// Создаем cron с секундной точностью
	c := cron.New(cron.WithSeconds())

	return &Scheduler{
		habitService:       habitService,
		logService:         logService,
		reminderService:    reminderService,
		streakResetService: streakResetService,
		userService:        userService,
		rabbitMQ:           rabbitMQ,
		cron:               c,
	}
}

// Start запускает все периодические задачи
func (s *Scheduler) Start() error {
	ctx := context.Background()

	// Задача 1: Проверка неподтвержденных привычек в 23:59 каждый день
	// Формат cron: "секунды минуты часы день месяц день_недели"
	_, err := s.cron.AddFunc("0 59 23 * * *", func() {
		s.checkUnconfirmedHabits(ctx)
	})
	if err != nil {
		logger.Error("Failed to schedule checkUnconfirmedHabits task", zap.Error(err))
		return err
	}

	// Задача 2: Обработка очереди сброса стриков в 00:10 каждый день
	_, err = s.cron.AddFunc("0 10 0 * * *", func() {
		s.processStreakResetQueue(ctx)
	})
	if err != nil {
		logger.Error("Failed to schedule processStreakResetQueue task", zap.Error(err))
		return err
	}

	// Задача 3: Отправка уведомлений о неподтвержденных привычках в 21:00
	_, err = s.cron.AddFunc("0 0 21 * * *", func() {
		s.sendUnconfirmedNotifications(ctx)
	})
	if err != nil {
		logger.Error("Failed to schedule sendUnconfirmedNotifications task", zap.Error(err))
		return err
	}

	// Запускаем cron scheduler
	s.cron.Start()
	logger.Info("Scheduler started successfully with 3 tasks",
		zap.String("task1", "23:59 - Check unconfirmed habits"),
		zap.String("task2", "00:10 - Process streak reset queue"),
		zap.String("task3", "21:00 - Send notifications"),
	)

	return nil
}

// Stop останавливает scheduler
func (s *Scheduler) Stop() {
	logger.Info("Scheduler stopping...")
	ctx := s.cron.Stop()
	<-ctx.Done()
	logger.Info("Scheduler stopped successfully")
}

// checkUnconfirmedHabits - Задача в 23:59
// Получает неподтвержденные habit_reminders за сегодня,
// добавляет их в очередь на сброс стрика и создает следующие напоминания
func (s *Scheduler) checkUnconfirmedHabits(ctx context.Context) {
	logger.Info("Starting checkUnconfirmedHabits task")
	startTime := time.Now()

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	// Получаем все неподтвержденные напоминания на сегодня
	unconfirmedReminders, err := s.reminderService.GetUnconfirmedRemindersByDate(ctx, todayDate)
	if err != nil {
		logger.Error("Failed to get unconfirmed reminders", zap.Error(err))
		return
	}

	logger.Info("Found unconfirmed reminders", zap.Int("count", len(unconfirmedReminders)))

	successCount := 0
	errorCount := 0

	for _, reminder := range unconfirmedReminders {
		// Получаем привычку
		habit, err := s.habitService.GetHabit(ctx, reminder.HabitID)
		if err != nil {
			logger.Error("Failed to get habit", zap.Error(err), zap.Int("habit_id", reminder.HabitID))
			errorCount++
			continue
		}

		// Добавляем в очередь на сброс стрика
		err = s.streakResetService.QueueStreakReset(ctx, habit.ID, habit.UserID, todayDate)
		if err != nil {
			logger.Error("Failed to queue streak reset", zap.Error(err), zap.Int("habit_id", habit.ID))
			errorCount++
			continue
		}

		// Создаем следующее напоминание
		_, err = s.reminderService.CreateNextReminder(ctx, habit, todayDate)
		if err != nil {
			logger.Error("Failed to create next reminder", zap.Error(err), zap.Int("habit_id", habit.ID))
			errorCount++
			continue
		}

		successCount++
	}

	duration := time.Since(startTime)
	logger.Info("Completed checkUnconfirmedHabits task",
		zap.Int("total", len(unconfirmedReminders)),
		zap.Int("success", successCount),
		zap.Int("errors", errorCount),
		zap.Duration("duration", duration),
	)
}

// processStreakResetQueue - Задача в 00:10
// Получает привычки из reset_queue на сегодня,
// ресетает стрики и сохраняет best_streak
func (s *Scheduler) processStreakResetQueue(ctx context.Context) {
	logger.Info("Starting processStreakResetQueue task")
	startTime := time.Now()

	err := s.streakResetService.ProcessQueueEntries(ctx)
	if err != nil {
		logger.Error("Failed to process streak reset queue", zap.Error(err))
		return
	}

	duration := time.Since(startTime)
	logger.Info("Completed processStreakResetQueue task", zap.Duration("duration", duration))
}

// sendUnconfirmedNotifications - Задача в 21:00
// Проверяет habit_reminders и если у пользователя есть неподтвержденные привычки,
// отправляет сообщение в RabbitMQ для Telegram бота
func (s *Scheduler) sendUnconfirmedNotifications(ctx context.Context) {
	logger.Info("Starting sendUnconfirmedNotifications task")
	startTime := time.Now()

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	// Получаем всех пользователей
	users, err := s.userService.GetAllUsers(ctx)
	if err != nil {
		logger.Error("Failed to get all users", zap.Error(err))
		return
	}

	logger.Info("Checking unconfirmed habits for users", zap.Int("user_count", len(users)))

	sentCount := 0
	errorCount := 0

	for _, user := range users {
		// Подсчитываем неподтвержденные напоминания для пользователя
		unconfirmedCount, err := s.reminderService.CountUnconfirmedRemindersByUserAndDate(ctx, user.ID, todayDate)
		if err != nil {
			logger.Error("Failed to count unconfirmed reminders",
				zap.Error(err),
				zap.Int("user_id", user.ID),
			)
			errorCount++
			continue
		}

		// Если есть неподтвержденные привычки, отправляем уведомление
		if unconfirmedCount > 0 {
			message := queue.NotificationMessage{
				UserID:           user.ID,
				TelegramID:       user.TelegramID,
				Message:          "У вас есть неподтвержденные привычки",
				UnconfirmedCount: unconfirmedCount,
			}

			err = s.rabbitMQ.PublishNotification(ctx, message)
			if err != nil {
				logger.Error("Failed to publish notification",
					zap.Error(err),
					zap.Int("user_id", user.ID),
					zap.Int64("telegram_id", user.TelegramID),
				)
				errorCount++
				continue
			}

			logger.Debug("Notification sent",
				zap.Int("user_id", user.ID),
				zap.Int("unconfirmed_count", unconfirmedCount),
			)
			sentCount++
		}
	}

	duration := time.Since(startTime)
	logger.Info("Completed sendUnconfirmedNotifications task",
		zap.Int("users_checked", len(users)),
		zap.Int("notifications_sent", sentCount),
		zap.Int("errors", errorCount),
		zap.Duration("duration", duration),
	)
}

// Manual execution methods for testing or manual triggers

// ManualCheckUnconfirmedHabits позволяет вручную запустить проверку неподтвержденных привычек
func (s *Scheduler) ManualCheckUnconfirmedHabits(ctx context.Context) {
	logger.Info("Manual execution of checkUnconfirmedHabits")
	s.checkUnconfirmedHabits(ctx)
}

// ManualProcessStreakResetQueue позволяет вручную запустить обработку очереди
func (s *Scheduler) ManualProcessStreakResetQueue(ctx context.Context) {
	logger.Info("Manual execution of processStreakResetQueue")
	s.processStreakResetQueue(ctx)
}

// ManualSendUnconfirmedNotifications позволяет вручную запустить отправку уведомлений
func (s *Scheduler) ManualSendUnconfirmedNotifications(ctx context.Context) {
	logger.Info("Manual execution of sendUnconfirmedNotifications")
	s.sendUnconfirmedNotifications(ctx)
}
