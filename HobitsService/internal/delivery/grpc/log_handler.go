package grpc

import (
	"context"
	"strconv"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "HobitsService/gen/go/HobitsService/gen/go/hobbits/api/v1"
	"HobitsService/internal/logger"
	"HobitsService/internal/metrics"
	"HobitsService/internal/service"
)

// LogServiceServer реализация LogService
type LogServiceServer struct {
	api.UnimplementedLogServiceServer
	logService *service.LogService
}

// NewLogServiceServer создает новый LogServiceServer
func NewLogServiceServer(logService *service.LogService) *LogServiceServer {
	return &LogServiceServer{
		logService: logService,
	}
}

// LogCompletion логирует выполнение привычки
func (s *LogServiceServer) LogCompletion(ctx context.Context, req *api.LogCompletionRequest) (*api.LogCompletionResponse, error) {
	logger.Debug("LogCompletion called", zap.Int32("habit_id", req.HabitId), zap.Int32("user_id", req.UserId))

	log, err := s.logService.LogCompletion(ctx, int(req.HabitId), int(req.UserId), req.Comment)
	if err != nil {
		logger.Error("failed to log completion", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to log completion: %v", err)
	}

	// Метрики
	metrics.LoggingsCreated.WithLabelValues(
		strconv.Itoa(int(req.HabitId)),
		strconv.Itoa(int(req.UserId)),
	).Inc()

	return &api.LogCompletionResponse{
		Log:               habitLogToProto(log),
		IsFirstCompletion: log.ID > 0,
	}, nil
}

// GetHabitLogs получает логи привычки
func (s *LogServiceServer) GetHabitLogs(ctx context.Context, req *api.GetHabitLogsRequest) (*api.GetHabitLogsResponse, error) {
	logger.Debug("GetHabitLogs called", zap.Int32("habit_id", req.HabitId))

	logs, err := s.logService.GetHabitLogs(ctx, int(req.HabitId))
	if err != nil {
		logger.Error("failed to get habit logs", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get logs: %v", err)
	}

	protoLogs := make([]*api.HabitLog, len(logs))
	for i, l := range logs {
		protoLogs[i] = habitLogToProto(l)
	}

	return &api.GetHabitLogsResponse{
		Logs: protoLogs,
	}, nil
}

// GetHabitLogsByDateRange получает логи за период
func (s *LogServiceServer) GetHabitLogsByDateRange(ctx context.Context, req *api.GetHabitLogsByDateRangeRequest) (*api.GetHabitLogsByDateRangeResponse, error) {
	logger.Debug("GetHabitLogsByDateRange called", zap.Int32("habit_id", req.HabitId))

	fromDate := req.FromDate.AsTime()
	toDate := req.ToDate.AsTime()

	logs, err := s.logService.GetHabitLogsByDateRange(ctx, int(req.HabitId), fromDate, toDate)
	if err != nil {
		logger.Error("failed to get habit logs by date range", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get logs: %v", err)
	}

	protoLogs := make([]*api.HabitLog, len(logs))
	for i, l := range logs {
		protoLogs[i] = habitLogToProto(l)
	}

	return &api.GetHabitLogsByDateRangeResponse{
		Logs: protoLogs,
	}, nil
}

// GetCompletionRate получает процент выполнения за период
func (s *LogServiceServer) GetCompletionRate(ctx context.Context, req *api.GetCompletionRateRequest) (*api.GetCompletionRateResponse, error) {
	logger.Debug("GetCompletionRate called", zap.Int32("habit_id", req.HabitId))

	fromDate := req.FromDate.AsTime()
	toDate := req.ToDate.AsTime()

	rate, err := s.logService.GetCompletionRate(ctx, int(req.HabitId), fromDate, toDate)
	if err != nil {
		logger.Error("failed to get completion rate", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get completion rate: %v", err)
	}

	return &api.GetCompletionRateResponse{
		Rate: float32(rate),
	}, nil
}
