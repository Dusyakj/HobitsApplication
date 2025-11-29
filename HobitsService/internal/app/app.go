package app

import (
	"HobitsService/internal/delivery/grpc"
	"HobitsService/internal/infrastructure/database"
	"HobitsService/internal/infrastructure/queue"
	"HobitsService/internal/infrastructure/scheduler"
	"HobitsService/internal/logger"
	"HobitsService/internal/repository/postgres"
	"HobitsService/internal/service"
	"go.uber.org/zap"
)

// App содержит все зависимости приложения
type App struct {
	// Infrastructure
	Database *database.Database
	RabbitMQ *queue.RabbitMQClient

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
func NewApp(db *database.Database, rabbitMQURL string) *App {
	// Initialize repositories
	userRepo := postgres.NewUserRepository(db.Pool)
	habitRepo := postgres.NewHabitRepository(db.Pool)
	habitLogRepo := postgres.NewHabitLogRepository(db.Pool)
	habitReminderRepo := postgres.NewHabitReminderRepository(db.Pool)
	streakResetQueueRepo := postgres.NewStreakResetQueueRepository(db.Pool)

	// Initialize services
	userService := service.NewUserService(userRepo)
	habitService := service.NewHabitService(habitRepo, habitLogRepo, habitReminderRepo)
	logService := service.NewLogService(habitLogRepo, habitRepo, habitReminderRepo, streakResetQueueRepo, habitService)
	reminderService := service.NewReminderService(habitReminderRepo, habitRepo, habitService)
	streakResetService := service.NewStreakResetService(streakResetQueueRepo, habitRepo, habitLogRepo, habitReminderRepo, habitService)

	// Initialize RabbitMQ client
	rabbitMQClient, err := queue.NewRabbitMQClient(rabbitMQURL)
	if err != nil {
		logger.Error("Failed to initialize RabbitMQ client", zap.Error(err))
		logger.Warn("Continuing without RabbitMQ - notifications will not be sent")
		rabbitMQClient = nil
	}

	// Initialize gRPC server
	grpcServer := grpc.NewServer(
		50051,
		userService,
		habitService,
		logService,
		reminderService,
	)

	// Initialize scheduler
	sched := scheduler.NewScheduler(
		habitService,
		logService,
		reminderService,
		streakResetService,
		userService,
		rabbitMQClient,
	)

	return &App{
		Database:                   db,
		RabbitMQ:                   rabbitMQClient,
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
	if a.Scheduler != nil {
		a.Scheduler.Stop()
	}

	if a.GRPCServer != nil {
		_ = a.GRPCServer.Stop()
	}

	if a.RabbitMQ != nil {
		_ = a.RabbitMQ.Close()
	}

	a.Database.Close()
	return nil
}
