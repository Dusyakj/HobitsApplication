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

// ReminderServiceServer реализация ReminderService
type ReminderServiceServer struct {
	api.UnimplementedReminderServiceServer
	reminderService *service.ReminderService
}

// NewReminderServiceServer создает новый ReminderServiceServer
func NewReminderServiceServer(reminderService *service.ReminderService) *ReminderServiceServer {
	return &ReminderServiceServer{
		reminderService: reminderService,
	}
}

// GenerateRemindersForToday генерирует напоминания на сегодня
func (s *ReminderServiceServer) GenerateRemindersForToday(ctx context.Context, req *api.GenerateRemindersForTodayRequest) (*api.GenerateRemindersForTodayResponse, error) {
	logger.Debug("GenerateRemindersForToday called", zap.Int32("user_id", req.UserId))

	reminders, err := s.reminderService.GenerateRemindersForToday(ctx, int(req.UserId))
	if err != nil {
		logger.Error("failed to generate reminders", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to generate reminders: %v", err)
	}

	protoReminders := make([]*api.HabitReminder, len(reminders))
	for i, r := range reminders {
		protoReminders[i] = habitReminderToProto(r)
	}

	// Метрики
	metrics.RemindersCreated.WithLabelValues(strconv.Itoa(int(req.UserId))).Add(float64(len(reminders)))

	return &api.GenerateRemindersForTodayResponse{
		Reminders: protoReminders,
		Count:     int32(len(reminders)),
	}, nil
}

// GetRemindersForDate получает напоминания на дату
func (s *ReminderServiceServer) GetRemindersForDate(ctx context.Context, req *api.GetRemindersForDateRequest) (*api.GetRemindersForDateResponse, error) {
	logger.Debug("GetRemindersForDate called")

	date := req.Date.AsTime()

	reminders, err := s.reminderService.GetRemindersByDate(ctx, date)
	if err != nil {
		logger.Error("failed to get reminders for date", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get reminders: %v", err)
	}

	protoReminders := make([]*api.HabitReminder, len(reminders))
	for i, r := range reminders {
		protoReminders[i] = habitReminderToProto(r)
	}

	return &api.GetRemindersForDateResponse{
		Reminders: protoReminders,
	}, nil
}

// GetUserRemindersForDate получает напоминания пользователя на дату
func (s *ReminderServiceServer) GetUserRemindersForDate(ctx context.Context, req *api.GetUserRemindersForDateRequest) (*api.GetUserRemindersForDateResponse, error) {
	logger.Debug("GetUserRemindersForDate called", zap.Int32("user_id", req.UserId))

	date := req.Date.AsTime()

	reminders, err := s.reminderService.GetRemindersByUserAndDate(ctx, int(req.UserId), date)
	if err != nil {
		logger.Error("failed to get user reminders for date", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get reminders: %v", err)
	}

	protoReminders := make([]*api.HabitReminder, len(reminders))
	completed := int32(0)
	for i, r := range reminders {
		protoReminders[i] = habitReminderToProto(r)
		if r.IsCompleted {
			completed++
		}
	}

	return &api.GetUserRemindersForDateResponse{
		Reminders:      protoReminders,
		CompletedCount: completed,
		TotalCount:     int32(len(reminders)),
	}, nil
}

// MarkReminderAsCompleted отмечает напоминание как выполненное
func (s *ReminderServiceServer) MarkReminderAsCompleted(ctx context.Context, req *api.MarkReminderAsCompletedRequest) (*api.MarkReminderAsCompletedResponse, error) {
	logger.Debug("MarkReminderAsCompleted called", zap.Int32("reminder_id", req.ReminderId))

	reminder, err := s.reminderService.MarkReminderAsCompleted(ctx, int(req.ReminderId))
	if err != nil {
		logger.Error("failed to mark reminder as completed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to mark reminder: %v", err)
	}

	// Метрики
	metrics.RemindersCompleted.WithLabelValues(strconv.Itoa(int(reminder.UserID))).Inc()

	return &api.MarkReminderAsCompletedResponse{
		Reminder: habitReminderToProto(reminder),
	}, nil
}

// MarkReminderAsIncomplete отмечает напоминание как невыполненное
func (s *ReminderServiceServer) MarkReminderAsIncomplete(ctx context.Context, req *api.MarkReminderAsIncompleteRequest) (*api.MarkReminderAsIncompleteResponse, error) {
	logger.Debug("MarkReminderAsIncomplete called", zap.Int32("reminder_id", req.ReminderId))

	reminder, err := s.reminderService.MarkReminderAsIncomplete(ctx, int(req.ReminderId))
	if err != nil {
		logger.Error("failed to mark reminder as incomplete", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to mark reminder: %v", err)
	}

	return &api.MarkReminderAsIncompleteResponse{
		Reminder: habitReminderToProto(reminder),
	}, nil
}
