package app

import (
	"HobitsService/internal/delivery/grpc"
	"HobitsService/internal/infrastructure/database"
	"HobitsService/internal/infrastructure/scheduler"
	"HobitsService/internal/repository/postgres"
	"HobitsService/internal/service"
)

// App содержит все зависимости приложения
type App struct {
	// Infrastructure
	Database *database.Database

	// Repositories
	UserRepository             *postgres.UserRepository
	HabitRepository            *postgres.HabitRepository
	HabitLogRepository         *postgres.HabitLogRepository
	HabitReminderRepository    *postgres.HabitReminderRepository
	StreakResetQueueRepository *postgres.StreakResetQueueRepository

	// Services
	UserService        *service.UserService
	HabitService       *service.HabitService
	LogService         *service.LogService
	ReminderService    *service.ReminderService
	StreakResetService *service.StreakResetService

	// Delivery
	GRPCServer *grpc.Server

	// Infrastructure
	Scheduler *scheduler.Scheduler
}

// NewApp инициализирует все зависимости и возвращает готовое приложение
func NewApp(db *database.Database) *App {
	// Инициализируем репозитории
	userRepo := postgres.NewUserRepository(db.Pool)
	habitRepo := postgres.NewHabitRepository(db.Pool)
	habitLogRepo := postgres.NewHabitLogRepository(db.Pool)
	habitReminderRepo := postgres.NewHabitReminderRepository(db.Pool)
	streakResetQueueRepo := postgres.NewStreakResetQueueRepository(db.Pool)

	// Инициализируем сервисы
	userService := service.NewUserService(userRepo)
	habitService := service.NewHabitService(habitRepo, habitLogRepo, habitReminderRepo)
	logService := service.NewLogService(habitLogRepo, habitRepo, habitReminderRepo, streakResetQueueRepo, habitService)
	reminderService := service.NewReminderService(habitReminderRepo, habitRepo, habitService)
	streakResetService := service.NewStreakResetService(streakResetQueueRepo, habitRepo, habitLogRepo, habitReminderRepo, habitService)

	// Инициализируем gRPC сервер (порт будет задан в main.go)
	grpcServer := grpc.NewServer(
		50051, // Порт по умолчанию
		userService,
		habitService,
		logService,
		reminderService,
	)

	// Инициализируем scheduler
	sched := scheduler.NewScheduler(
		habitService,
		logService,
		reminderService,
		streakResetService,
		userService,
	)

	return &App{
		Database:                   db,
		UserRepository:             userRepo,
		HabitRepository:            habitRepo,
		HabitLogRepository:         habitLogRepo,
		HabitReminderRepository:    habitReminderRepo,
		StreakResetQueueRepository: streakResetQueueRepo,
		UserService:                userService,
		HabitService:               habitService,
		LogService:                 logService,
		ReminderService:            reminderService,
		StreakResetService:         streakResetService,
		GRPCServer:                 grpcServer,
		Scheduler:                  sched,
	}
}

// Close закрывает все подключения и останавливает сервисы
func (a *App) Close() error {
	// Останавливаем scheduler
	if a.Scheduler != nil {
		a.Scheduler.Stop()
	}

	// Останавливаем gRPC сервер
	if a.GRPCServer != nil {
		_ = a.GRPCServer.Stop()
	}

	// Закрываем БД
	a.Database.Close()
	return nil
}
