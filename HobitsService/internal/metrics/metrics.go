package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Метрики приложения
var (
	// Количество созданных привычек
	HabitsCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hobbits_habits_created_total",
			Help: "Total number of habits created",
		},
		[]string{"user_id"},
	)

	// Количество логирований выполнений
	LoggingsCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hobbits_loggings_created_total",
			Help: "Total number of habit completions logged",
		},
		[]string{"habit_id", "user_id"},
	)

	// Количество сбросов стриков
	StreaksReset = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hobbits_streaks_reset_total",
			Help: "Total number of streaks reset",
		},
		[]string{"habit_id", "user_id"},
	)

	// Текущая длина стрика по привычкам
	CurrentStreak = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hobbits_current_streak",
			Help: "Current streak of a habit",
		},
		[]string{"habit_id", "user_id", "habit_name"},
	)

	// Лучший стрик за все время
	BestStreak = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hobbits_best_streak",
			Help: "Best streak of a habit",
		},
		[]string{"habit_id", "user_id", "habit_name"},
	)

	// Количество активных пользователей
	ActiveUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "hobbits_active_users",
			Help: "Number of active users with at least one habit",
		},
	)

	// Количество активных привычек
	ActiveHabits = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "hobbits_active_habits",
			Help: "Total number of active habits",
		},
	)

	// Время выполнения операций БД
	DBOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hobbits_db_operation_duration_seconds",
			Help:    "Duration of database operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Количество ошибок при операциях с БД
	DBErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hobbits_db_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation", "table", "error_type"},
	)

	// Количество напоминаний
	RemindersCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hobbits_reminders_created_total",
			Help: "Total number of reminders created",
		},
		[]string{"user_id"},
	)

	// Количество выполненных напоминаний
	RemindersCompleted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hobbits_reminders_completed_total",
			Help: "Total number of completed reminders",
		},
		[]string{"user_id"},
	)

	// Процент выполнения привычек
	CompletionRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hobbits_completion_rate_percent",
			Help: "Completion rate of habits in percent",
		},
		[]string{"habit_id", "user_id", "habit_name", "period"},
	)
)

// Init регистрирует все метрики в Prometheus registry
func Init() error {
	prometheus.MustRegister(
		HabitsCreated,
		LoggingsCreated,
		StreaksReset,
		CurrentStreak,
		BestStreak,
		ActiveUsers,
		ActiveHabits,
		DBOperationDuration,
		DBErrors,
		RemindersCreated,
		RemindersCompleted,
		CompletionRate,
	)
	return nil
}
