package service

import (
	"context"
	"fmt"
	"time"

	"HobitsService/internal/domain"
	"HobitsService/internal/repository"
)

// StreakResetService сервис для управления сбросом стриков
type StreakResetService struct {
	queueRepo    repository.StreakResetQueueRepository
	habitRepo    repository.HabitRepository
	logRepo      repository.HabitLogRepository
	reminderRepo repository.HabitReminderRepository
	habitService *HabitService
}

// NewStreakResetService создает новый StreakResetService
func NewStreakResetService(
	queueRepo repository.StreakResetQueueRepository,
	habitRepo repository.HabitRepository,
	logRepo repository.HabitLogRepository,
	reminderRepo repository.HabitReminderRepository,
	habitService *HabitService,
) *StreakResetService {
	return &StreakResetService{
		queueRepo:    queueRepo,
		habitRepo:    habitRepo,
		logRepo:      logRepo,
		reminderRepo: reminderRepo,
		habitService: habitService,
	}
}

// CheckAndQueueStreakResets проверяет все привычки и добавляет в очередь на сброс
// Должна вызваться каждый день вечером (например 23:59)
func (s *StreakResetService) CheckAndQueueStreakResets(ctx context.Context) error {
	//today := time.Now()
	//todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	// Получаем все активные привычки
	// Это неидеально, но без кэширования придется так
	// TODO: добавить оптимизацию с получением привычек порциями

	// Для каждой привычки проверяем, не пропустил ли пользователь
	// В идеале это должен быть конкретный запрос в БД, но здесь логика сложная
	fmt.Println("[StreakResetService] Starting streak check")

	// В реальной реализации нужно было бы получить все активные привычки с последней датой проверки
	// и для каждой проверить, есть ли пропуски

	return nil
}

// QueueStreakReset добавляет привычку в очередь на сброс стрика
func (s *StreakResetService) QueueStreakReset(ctx context.Context, habitID, userID int, resetDate time.Time) error {
	queueEntry := domain.NewStreakResetQueue(habitID, userID, resetDate)
	_, err := s.queueRepo.CreateQueueEntry(ctx, queueEntry)
	if err != nil {
		return fmt.Errorf("failed to create queue entry: %w", err)
	}
	return nil
}

// CheckHabitStreak проверяет, нужно ли сбросить стрик для привычки
func (s *StreakResetService) CheckHabitStreak(ctx context.Context, habitID int) error {
	habit, err := s.habitRepo.GetHabitByID(ctx, habitID)
	if err != nil {
		return fmt.Errorf("failed to get habit: %w", err)
	}

	if !habit.IsActive {
		return nil // Пропускаем неактивные привычки
	}

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	// Если привычка еще не выполнялась - пропускаем
	if !habit.LastCompletedDate.Valid {
		return nil
	}

	// Если уже проверили сегодня - пропускаем
	if habit.LastCheckedDate.Valid {
		lastCheckedDate := habit.LastCheckedDate.Time
		lastCheckedDateOnly := time.Date(lastCheckedDate.Year(), lastCheckedDate.Month(), lastCheckedDate.Day(), 0, 0, 0, 0, lastCheckedDate.Location())
		if lastCheckedDateOnly.Equal(todayDate) {
			return nil
		}
	}

	// Получаем все запланированные дни с момента последней проверки до сегодня
	lastCheckedDate := habit.LastCompletedDate.Time
	if habit.LastCheckedDate.Valid {
		lastCheckedDate = habit.LastCheckedDate.Time
	}

	lastCheckedDateOnly := time.Date(lastCheckedDate.Year(), lastCheckedDate.Month(), lastCheckedDate.Day(), 0, 0, 0, 0, lastCheckedDate.Location())

	scheduledDays, err := s.habitService.GetScheduledDaysBetween(ctx, habitID, lastCheckedDateOnly.AddDate(0, 0, 1), todayDate)
	if err != nil {
		return fmt.Errorf("failed to get scheduled days: %w", err)
	}

	// Проверяем каждый запланированный день
	for _, scheduledDay := range scheduledDays {
		// Пропускаем сегодня - если не выполнено, проверим завтра
		if scheduledDay.Equal(todayDate) {
			continue
		}

		// Проверяем, выполнена ли привычка в этот день
		log, err := s.logRepo.GetLogByHabitIDAndDate(ctx, habitID, scheduledDay)
		if err != nil || log == nil {
			// День пропущен - добавляем в очередь на сброс
			queueEntry := domain.NewStreakResetQueue(habitID, habit.UserID, scheduledDay)
			_, err := s.queueRepo.CreateQueueEntry(ctx, queueEntry)
			if err != nil {
				fmt.Printf("failed to create queue entry for habit %d, date %s: %v\n", habitID, scheduledDay.Format("2006-01-02"), err)
			}
		}
	}

	// Обновляем last_checked_date
	habit.UpdateLastCheckedDate(todayDate)
	_, err = s.habitRepo.UpdateHabit(ctx, habit)
	if err != nil {
		return fmt.Errorf("failed to update last_checked_date: %w", err)
	}

	return nil
}

// ProcessQueueEntries обрабатывает очередь на сброс стриков
// Должна вызваться после CheckAndQueueStreakResets (например 00:30)
func (s *StreakResetService) ProcessQueueEntries(ctx context.Context) error {
	entries, err := s.queueRepo.GetUnprocessedEntries(ctx)
	if err != nil {
		return fmt.Errorf("failed to get unprocessed entries: %w", err)
	}

	for _, entry := range entries {
		if err := s.processQueueEntry(ctx, entry); err != nil {
			fmt.Printf("failed to process queue entry %d: %v\n", entry.ID, err)
		}
	}

	return nil
}

// processQueueEntry обрабатывает одну запись в очереди
func (s *StreakResetService) processQueueEntry(ctx context.Context, entry *domain.StreakResetQueue) error {
	habit, err := s.habitRepo.GetHabitByID(ctx, entry.HabitID)
	if err != nil {
		return fmt.Errorf("failed to get habit: %w", err)
	}

	// Сохраняем предыдущий стрик для аудита
	entry.MarkAsProcessed(habit.CurrentStreak)

	// Если текущий стрик больше лучшего - обновляем лучший
	if habit.CurrentStreak > habit.BestStreak {
		habit.BestStreak = habit.CurrentStreak
	}

	// Сбрасываем текущий стрик
	habit.CurrentStreak = 0
	habit.UpdatedAt = time.Now()

	// Обновляем привычку
	if _, err := s.habitRepo.UpdateHabit(ctx, habit); err != nil {
		return fmt.Errorf("failed to update habit: %w", err)
	}

	// Обновляем запись в очереди
	now := time.Now()
	entry.ProcessedAt.Time = now
	entry.ProcessedAt.Valid = true
	entry.Processed = true

	if _, err := s.queueRepo.UpdateQueueEntry(ctx, entry); err != nil {
		return fmt.Errorf("failed to update queue entry: %w", err)
	}

	// Отмечаем напоминание на эту дату как невыполненное
	resetDate := entry.GetResetDate()
	reminder, err := s.reminderRepo.GetReminderByHabitIDAndDate(ctx, entry.HabitID, resetDate)
	if err == nil && reminder != nil {
		reminder.MarkAsIncomplete()
		_, _ = s.reminderRepo.UpdateReminder(ctx, reminder)
	}

	return nil
}

// GetQueueEntry получает запись из очереди
func (s *StreakResetService) GetQueueEntry(ctx context.Context, entryID int) (*domain.StreakResetQueue, error) {
	return s.queueRepo.GetQueueEntryByID(ctx, entryID)
}

// GetUnprocessedQueueEntries получает необработанные записи
func (s *StreakResetService) GetUnprocessedQueueEntries(ctx context.Context) ([]*domain.StreakResetQueue, error) {
	return s.queueRepo.GetUnprocessedEntries(ctx)
}
