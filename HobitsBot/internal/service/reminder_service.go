package service

import (
	"context"
	"fmt"
	"time"

	pb "HobitsBot/gen/go/hobbits/api/v1"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ReminderService struct {
	client pb.ReminderServiceClient
	logger Logger
}

type ReminderResponse struct {
	ID           int32
	HabitID      int32
	UserID       int32
	IsCompleted  bool
	ReminderDate time.Time
	SentAt       *time.Time
}

func NewReminderService(conn *grpc.ClientConn, logger Logger) *ReminderService {
	return &ReminderService{
		client: pb.NewReminderServiceClient(conn),
		logger: logger,
	}
}

// GenerateRemindersForToday генерирует напоминания на сегодня
func (s *ReminderService) GenerateRemindersForToday(
	ctx context.Context,
	userID int,
) ([]*ReminderResponse, error) {
	req := &pb.GenerateRemindersForTodayRequest{
		UserId: int32(userID),
	}

	resp, err := s.client.GenerateRemindersForToday(ctx, req)
	if err != nil {
		s.logger.Error("failed to generate reminders: %v", err)
		return nil, fmt.Errorf("не удалось сгенерировать напоминания: %w", err)
	}

	reminders := make([]*ReminderResponse, len(resp.Reminders))
	for i, reminder := range resp.Reminders {
		reminders[i] = s.pbReminderToResponse(reminder)
	}

	return reminders, nil
}

// GetRemindersByUserAndDate получает напоминания пользователя на дату
func (s *ReminderService) GetRemindersByUserAndDate(
	ctx context.Context,
	userID int,
	date time.Time,
) ([]*ReminderResponse, error) {
	req := &pb.GetUserRemindersForDateRequest{
		UserId: int32(userID),
		Date:   timestamppb.New(date),
	}

	resp, err := s.client.GetUserRemindersForDate(ctx, req)
	if err != nil {
		s.logger.Error("failed to get user reminders: %v", err)
		return nil, fmt.Errorf("не удалось получить напоминания: %w", err)
	}

	reminders := make([]*ReminderResponse, len(resp.Reminders))
	for i, reminder := range resp.Reminders {
		reminders[i] = s.pbReminderToResponse(reminder)
	}

	return reminders, nil
}

// GetRemindersByDate получает все напоминания на дату
func (s *ReminderService) GetRemindersByDate(
	ctx context.Context,
	date time.Time,
) ([]*ReminderResponse, error) {
	req := &pb.GetRemindersForDateRequest{
		Date: timestamppb.New(date),
	}

	resp, err := s.client.GetRemindersForDate(ctx, req)
	if err != nil {
		s.logger.Error("failed to get reminders for date: %v", err)
		return nil, fmt.Errorf("не удалось получить напоминания: %w", err)
	}

	reminders := make([]*ReminderResponse, len(resp.Reminders))
	for i, reminder := range resp.Reminders {
		reminders[i] = s.pbReminderToResponse(reminder)
	}

	return reminders, nil
}

// MarkReminderAsCompleted отмечает напоминание как выполненное
func (s *ReminderService) MarkReminderAsCompleted(
	ctx context.Context,
	reminderID int,
) (*ReminderResponse, error) {
	req := &pb.MarkReminderAsCompletedRequest{
		ReminderId: int32(reminderID),
	}

	resp, err := s.client.MarkReminderAsCompleted(ctx, req)
	if err != nil {
		s.logger.Error("failed to mark reminder as completed: %v", err)
		return nil, fmt.Errorf("не удалось отметить напоминание: %w", err)
	}

	return s.pbReminderToResponse(resp.Reminder), nil
}

// MarkReminderAsIncomplete отмечает напоминание как невыполненное
func (s *ReminderService) MarkReminderAsIncomplete(
	ctx context.Context,
	reminderID int,
) (*ReminderResponse, error) {
	req := &pb.MarkReminderAsIncompleteRequest{
		ReminderId: int32(reminderID),
	}

	resp, err := s.client.MarkReminderAsIncomplete(ctx, req)
	if err != nil {
		s.logger.Error("failed to mark reminder as incomplete: %v", err)
		return nil, fmt.Errorf("не удалось отметить напоминание: %w", err)
	}

	return s.pbReminderToResponse(resp.Reminder), nil
}

// pbReminderToResponse преобразует protobuf напоминание в ответ
func (s *ReminderService) pbReminderToResponse(pb *pb.HabitReminder) *ReminderResponse {
	reminderDate := time.Now()
	if pb.ReminderDate != nil {
		reminderDate = pb.ReminderDate.AsTime()
	}

	var sentAt *time.Time
	if pb.SentAt != nil {
		t := pb.SentAt.AsTime()
		sentAt = &t
	}

	return &ReminderResponse{
		ID:           pb.Id,
		HabitID:      pb.HabitId,
		UserID:       pb.UserId,
		IsCompleted:  pb.IsCompleted,
		ReminderDate: reminderDate,
		SentAt:       sentAt,
	}
}
