package service

import (
	"context"
	"fmt"
	"time"

	pb "HobitsBot/gen/go/hobbits/api/v1"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LogService struct {
	client pb.LogServiceClient
	logger Logger
}

type LogResponse struct {
	ID          int32
	HabitID     int32
	UserID      int32
	Comment     string
	LoggedAt    time.Time
	IsCompleted bool
}

func NewLogService(conn *grpc.ClientConn, logger Logger) *LogService {
	return &LogService{
		client: pb.NewLogServiceClient(conn),
		logger: logger,
	}
}

// LogCompletion регистрирует выполнение привычки
func (s *LogService) LogCompletion(
	ctx context.Context,
	habitID, userID int,
	comment string,
) (*LogResponse, error) {
	req := &pb.LogCompletionRequest{
		HabitId: int32(habitID),
		UserId:  int32(userID),
		Comment: comment,
	}

	resp, err := s.client.LogCompletion(ctx, req)
	if err != nil {
		s.logger.Error("failed to log completion: %v", err)
		return nil, fmt.Errorf("не удалось зарегистрировать выполнение: %w", err)
	}

	return s.pbLogToResponse(resp.Log), nil
}

// GetHabitLogs получает логи привычки
func (s *LogService) GetHabitLogs(ctx context.Context, habitID int) ([]*LogResponse, error) {
	req := &pb.GetHabitLogsRequest{
		HabitId: int32(habitID),
	}

	resp, err := s.client.GetHabitLogs(ctx, req)
	if err != nil {
		s.logger.Error("failed to get habit logs: %v", err)
		return nil, fmt.Errorf("не удалось получить логи: %w", err)
	}

	logs := make([]*LogResponse, len(resp.Logs))
	for i, log := range resp.Logs {
		logs[i] = s.pbLogToResponse(log)
	}

	return logs, nil
}

// GetHabitLogsByDateRange получает логи за период
func (s *LogService) GetHabitLogsByDateRange(
	ctx context.Context,
	habitID int,
	from, to time.Time,
) ([]*LogResponse, error) {
	req := &pb.GetHabitLogsByDateRangeRequest{
		HabitId:  int32(habitID),
		FromDate: timestamppb.New(from),
		ToDate:   timestamppb.New(to),
	}

	resp, err := s.client.GetHabitLogsByDateRange(ctx, req)
	if err != nil {
		s.logger.Error("failed to get logs by date range: %v", err)
		return nil, fmt.Errorf("не удалось получить логи за период: %w", err)
	}

	logs := make([]*LogResponse, len(resp.Logs))
	for i, log := range resp.Logs {
		logs[i] = s.pbLogToResponse(log)
	}

	return logs, nil
}

// GetCompletionRate получает процент выполнения за период
func (s *LogService) GetCompletionRate(
	ctx context.Context,
	habitID int,
	from, to time.Time,
) (float32, error) {
	req := &pb.GetCompletionRateRequest{
		HabitId:  int32(habitID),
		FromDate: timestamppb.New(from),
		ToDate:   timestamppb.New(to),
	}

	resp, err := s.client.GetCompletionRate(ctx, req)
	if err != nil {
		s.logger.Error("failed to get completion rate: %v", err)
		return 0, fmt.Errorf("не удалось получить процент выполнения: %w", err)
	}

	return resp.Rate, nil
}

// pbLogToResponse преобразует protobuf лог в ответ
func (s *LogService) pbLogToResponse(pb *pb.HabitLog) *LogResponse {
	loggedAt := time.Now()
	if pb.LoggedAt != nil {
		loggedAt = pb.LoggedAt.AsTime()
	}

	return &LogResponse{
		ID:          pb.Id,
		HabitID:     pb.HabitId,
		UserID:      pb.UserId,
		Comment:     pb.Comment,
		LoggedAt:    loggedAt,
		IsCompleted: true, // По умолчанию логи создаются для выполненных привычек
	}
}
