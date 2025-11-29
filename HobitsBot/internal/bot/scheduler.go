package bot

import (
	"context"
	"fmt"
	"time"

	"HobitsBot/internal/service"
)

// Scheduler управляет расписанием задач
type Scheduler struct {
	bot              *Bot
	reminderService  *service.ReminderService
	userService      *service.UserService
	logger           Logger
	stopChan         chan struct{}
	isRunning        bool
	notificationTime time.Time
}

// NewScheduler создаёт новый планировщик
func NewScheduler(
	bot *Bot,
	reminderService *service.ReminderService,
	userService *service.UserService,
	logger Logger,
) *Scheduler {
	return &Scheduler{
		bot:              bot,
		reminderService:  reminderService,
		userService:      userService,
		logger:           logger,
		stopChan:         make(chan struct{}),
		notificationTime: parseNotificationTime("21:00"), // 9:00 PM (21:00)
	}
}

// parseNotificationTime преобразует строку в time.Time для текущего дня
func parseNotificationTime(timeStr string) time.Time {
	now := time.Now()
	parts := len(timeStr) == 5 && timeStr[2] == ':'
	if !parts {
		// Fallback to 9:00 AM if parsing fails
		return time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
	}

	hour := int(timeStr[0]-'0')*10 + int(timeStr[1]-'0')
	minute := int(timeStr[3]-'0')*10 + int(timeStr[4]-'0')

	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
}

// Start запускает планировщик
func (s *Scheduler) Start() {
	if s.isRunning {
		s.logger.Warn("Scheduler is already running")
		return
	}

	s.isRunning = true
	s.logger.Info("Scheduler started")

	go s.runSchedule()
}

// Stop останавливает планировщик
func (s *Scheduler) Stop() {
	if !s.isRunning {
		s.logger.Warn("Scheduler is not running")
		return
	}

	s.isRunning = false
	close(s.stopChan)
	s.logger.Info("Scheduler stopped")
}

// runSchedule основной цикл планировщика
func (s *Scheduler) runSchedule() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			s.logger.Info("Scheduler received stop signal")
			return
		case <-ticker.C:
			s.checkAndSendNotifications()
		}
	}
}

// checkAndSendNotifications проверяет время и отправляет уведомления
func (s *Scheduler) checkAndSendNotifications() {
	now := time.Now()
	targetHour := 21 // 9 PM
	targetMinute := 0

	// Проверяем, наступило ли время отправки уведомлений
	// Даём окно в 1 минуту (21:00-21:01)
	if now.Hour() == targetHour && now.Minute() == targetMinute {
		s.logger.Info("Time to send 9:00 PM (21:00) notifications")
		s.sendUnconfirmedHabitsNotification()
	}
}

// sendUnconfirmedHabitsNotification отправляет уведомления о неподтверждённых привычках
func (s *Scheduler) sendUnconfirmedHabitsNotification() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Примечание: В идеале нужна была бы GetAllUsers, но так как её нет,
	// мы проверяем все напоминания за сегодня и отправляем уведомления
	// для каждого уникального пользователя

	remindersToday, err := s.reminderService.GetRemindersByDate(ctx, time.Now())
	if err != nil {
		s.logger.Error("Failed to get reminders for today: %v", err)
		return
	}

	s.logger.Info("Got %d reminders for today", len(remindersToday))

	// Группируем напоминания по пользователям
	userReminders := make(map[int32][]*service.ReminderResponse)
	for _, reminder := range remindersToday {
		userReminders[reminder.UserID] = append(userReminders[reminder.UserID], reminder)
	}

	// Отправляем уведомления каждому пользователю
	for userID, reminders := range userReminders {
		// Фильтруем только незавершённые напоминания
		var incompleteReminders []*service.ReminderResponse
		for _, reminder := range reminders {
			if !reminder.IsCompleted {
				incompleteReminders = append(incompleteReminders, reminder)
			}
		}

		// Если есть незавершённые напоминания, отправляем уведомление
		if len(incompleteReminders) > 0 {
			s.sendNotificationToUser(int64(userID), len(incompleteReminders))
		}
	}
}

// sendNotificationToUser отправляет уведомление пользователю
func (s *Scheduler) sendNotificationToUser(userID int64, incompleteCount int) {
	msg := fmt.Sprintf(`🔔 *Напоминание о незавершённых привычках*

У вас есть *%d* неподтверждённ%s привычк%s на сегодня!

Нажмите /today чтобы посмотреть подробный список и выполнить их. 💪`,
		incompleteCount,
		getPluralForm(incompleteCount, "а", "ы", ""),
		getPluralForm(incompleteCount, "а", "и", ""),
	)

	err := s.bot.SendMessage(userID, msg)
	if err != nil {
		s.logger.Error("Failed to send notification to user %d: %v", userID, err)
	} else {
		s.logger.Info("Sent 9:00 PM notification to user %d (%d incomplete reminders)", userID, incompleteCount)
	}
}

// getPluralForm возвращает правильную форму слова для русского языка
func getPluralForm(count int, form1, form5, form21 string) string {
	if count%10 == 1 && count%100 != 11 {
		return form1
	}
	if count%10 >= 2 && count%10 <= 4 && (count%100 < 10 || count%100 >= 20) {
		return form5
	}
	return form21
}
