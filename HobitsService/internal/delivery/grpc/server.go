package grpc

import (
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	api "HobitsService/gen/go/HobitsService/gen/go/hobbits/api/v1"
	"HobitsService/internal/logger"
	"HobitsService/internal/service"
)

// Server gRPC сервер
type Server struct {
	server *grpc.Server
	port   int

	// Services
	userService     *service.UserService
	habitService    *service.HabitService
	logService      *service.LogService
	reminderService *service.ReminderService
}

// NewServer создает новый gRPC сервер
func NewServer(
	port int,
	userService *service.UserService,
	habitService *service.HabitService,
	logService *service.LogService,
	reminderService *service.ReminderService,
) *Server {
	return &Server{
		port:            port,
		userService:     userService,
		habitService:    habitService,
		logService:      logService,
		reminderService: reminderService,
	}
}

// Start запускает gRPC сервер
func (s *Server) Start() error {
	// Создаем gRPC сервер
	s.server = grpc.NewServer()

	// Регистрируем сервисы
	api.RegisterUserServiceServer(s.server, NewUserServiceServer(s.userService))
	api.RegisterHabitServiceServer(s.server, NewHabitServiceServer(s.habitService))
	api.RegisterLogServiceServer(s.server, NewLogServiceServer(s.logService))
	api.RegisterReminderServiceServer(s.server, NewReminderServiceServer(s.reminderService))

	// Слушаем на порту
	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Запускаем в горутине
	go func() {
		logger.Info("gRPC server started", zap.Int("port", s.port))
		if err := s.server.Serve(listener); err != nil {
			logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	return nil
}

// Stop останавливает gRPC сервер
func (s *Server) Stop() error {
	if s.server != nil {
		s.server.GracefulStop()
		logger.Info("gRPC server stopped")
	}
	return nil
}
