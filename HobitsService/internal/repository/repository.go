package repository

import (
	"context"
	"time"

	"HobitsService/internal/domain"
)

// UserRepository определяет интерфейс для работы с пользователями
type UserRepository interface {
	// CreateUser создает нового пользователя
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	// GetUserByID получает пользователя по ID
	GetUserByID(ctx context.Context, id int) (*domain.User, error)
	// GetUserByTelegramID получает пользователя по Telegram ID
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error)
	// GetAllUsers получает всех пользователей
	GetAllUsers(ctx context.Context) ([]*domain.User, error)
	// UpdateUser обновляет пользователя
	UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	// DeleteUser удаляет пользователя
	DeleteUser(ctx context.Context, id int) error
}

// HabitRepository определяет интерфейс для работы с привычками
type HabitRepository interface {
	// CreateHabit создает новую привычку
	CreateHabit(ctx context.Context, habit *domain.Habit) (*domain.Habit, error)
	// GetHabitByID получает привычку по ID
	GetHabitByID(ctx context.Context, id int) (*domain.Habit, error)
	// GetHabitsByUserID получает все привычки пользователя
	GetHabitsByUserID(ctx context.Context, userID int) ([]*domain.Habit, error)
	// GetActiveHabitsByUserID получает активные привычки пользователя
	GetActiveHabitsByUserID(ctx context.Context, userID int) ([]*domain.Habit, error)
	// GetAllActiveHabits получает все активные привычки
	GetAllActiveHabits(ctx context.Context) ([]*domain.Habit, error)
	// UpdateHabit обновляет привычку
	UpdateHabit(ctx context.Context, habit *domain.Habit) (*domain.Habit, error)
	// DeleteHabit удаляет привычку
	DeleteHabit(ctx context.Context, id int) error
	// GetHabitByUserIDAndName получает привычку по user ID и названию
	GetHabitByUserIDAndName(ctx context.Context, userID int, name string) (*domain.Habit, error)
}

// HabitLogRepository определяет интерфейс для работы с логами привычек
type HabitLogRepository interface {
	// CreateLog создает новый лог выполнения
	CreateLog(ctx context.Context, log *domain.HabitLog) (*domain.HabitLog, error)
	// GetLogByID получает лог по ID
	GetLogByID(ctx context.Context, id int) (*domain.HabitLog, error)
	// GetLogsByHabitID получает логи по привычке
	GetLogsByHabitID(ctx context.Context, habitID int) ([]*domain.HabitLog, error)
	// GetLogsByHabitIDAndDate получает логи за определенный период
	GetLogsByHabitIDAndDate(ctx context.Context, habitID int, from, to time.Time) ([]*domain.HabitLog, error)
	// GetLogByHabitIDAndDate получает лог за конкретный день
	GetLogByHabitIDAndDate(ctx context.Context, habitID int, date time.Time) (*domain.HabitLog, error)
	// DeleteLog удаляет лог
	DeleteLog(ctx context.Context, id int) error
	// CountLogsByHabitIDAndDate считает логи за период
	CountLogsByHabitIDAndDate(ctx context.Context, habitID int, from, to time.Time) (int, error)
}

// HabitReminderRepository определяет интерфейс для работы с напоминаниями
type HabitReminderRepository interface {
	// CreateReminder создает новое напоминание
	CreateReminder(ctx context.Context, reminder *domain.HabitReminder) (*domain.HabitReminder, error)
	// GetReminderByID получает напоминание по ID
	GetReminderByID(ctx context.Context, id int) (*domain.HabitReminder, error)
	// GetRemindersByUserID получает напоминания пользователя
	GetRemindersByUserID(ctx context.Context, userID int) ([]*domain.HabitReminder, error)
	// GetRemindersByDate получает напоминания на дату
	GetRemindersByDate(ctx context.Context, date time.Time) ([]*domain.HabitReminder, error)
	// GetRemindersByUserIDAndDate получает напоминания пользователя на дату
	GetRemindersByUserIDAndDate(ctx context.Context, userID int, date time.Time) ([]*domain.HabitReminder, error)
	// GetReminderByHabitIDAndDate получает напоминание по привычке и дате
	GetReminderByHabitIDAndDate(ctx context.Context, habitID int, date time.Time) (*domain.HabitReminder, error)
	// UpdateReminder обновляет напоминание
	UpdateReminder(ctx context.Context, reminder *domain.HabitReminder) (*domain.HabitReminder, error)
	// DeleteReminder удаляет напоминание
	DeleteReminder(ctx context.Context, id int) error
}

// StreakResetQueueRepository определяет интерфейс для работы с очередью сброса стриков
type StreakResetQueueRepository interface {
	// CreateQueueEntry создает новую запись в очередь
	CreateQueueEntry(ctx context.Context, entry *domain.StreakResetQueue) (*domain.StreakResetQueue, error)
	// GetQueueEntryByID получает запись по ID
	GetQueueEntryByID(ctx context.Context, id int) (*domain.StreakResetQueue, error)
	// GetUnprocessedEntries получает необработанные записи
	GetUnprocessedEntries(ctx context.Context) ([]*domain.StreakResetQueue, error)
	// GetUnprocessedEntriesByDate получает необработанные записи на дату
	GetUnprocessedEntriesByDate(ctx context.Context, date time.Time) ([]*domain.StreakResetQueue, error)
	// UpdateQueueEntry обновляет запись в очередь
	UpdateQueueEntry(ctx context.Context, entry *domain.StreakResetQueue) (*domain.StreakResetQueue, error)
	// DeleteQueueEntry удаляет запись из очереди
	DeleteQueueEntry(ctx context.Context, id int) error
	// GetQueueEntryByHabitIDAndDate получает запись по привычке и дате
	GetQueueEntryByHabitIDAndDate(ctx context.Context, habitID int, date time.Time) (*domain.StreakResetQueue, error)
}
