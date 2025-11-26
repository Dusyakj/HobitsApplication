package grpc

import (
	"context"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "HobitsService/gen/go/HobitsService/gen/go/hobbits/api/v1"
	"HobitsService/internal/domain"
	"HobitsService/internal/logger"
	"HobitsService/internal/metrics"
	"HobitsService/internal/service"
)

// HabitServiceServer реализация HabitService
type HabitServiceServer struct {
	api.UnimplementedHabitServiceServer
	habitService *service.HabitService
}

// NewHabitServiceServer создает новый HabitServiceServer
func NewHabitServiceServer(habitService *service.HabitService) *HabitServiceServer {
	return &HabitServiceServer{
		habitService: habitService,
	}
}

// CreateHabit создает новую привычку
func (s *HabitServiceServer) CreateHabit(ctx context.Context, req *api.CreateHabitRequest) (*api.CreateHabitResponse, error) {
	logger.Debug("CreateHabit called", zap.Int32("user_id", req.UserId), zap.String("name", req.Name))

	habit, err := s.habitService.CreateHabit(ctx, int(req.UserId), req.Name, domain.HabitFrequency(req.Frequency))
	if err != nil {
		logger.Error("failed to create habit", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create habit: %v", err)
	}

	// Устанавливаем дни если они указаны
	if req.WeeklyDays != "" && req.Frequency == "weekly" {
		days := parseIntDays(req.WeeklyDays)
		habit, _ = s.habitService.SetWeeklyDays(ctx, habit.ID, days)
	}

	if req.MonthlyDays != "" && req.Frequency == "monthly" {
		days := parseIntDays(req.MonthlyDays)
		habit, _ = s.habitService.SetMonthlyDays(ctx, habit.ID, days)
	}

	// Устанавливаем описание и цель
	if req.Description != "" {
		habit.SetDescription(req.Description)
	}
	if req.Goal != "" {
		habit.SetGoal(req.Goal)
	}

	// Обновляем в БД
	habit, _ = s.habitService.UpdateHabit(ctx, habit)

	// Метрики
	metrics.HabitsCreated.WithLabelValues(strconv.Itoa(int(req.UserId))).Inc()

	return &api.CreateHabitResponse{
		Habit: habitToProto(habit),
	}, nil
}

// GetHabit получает привычку по ID
func (s *HabitServiceServer) GetHabit(ctx context.Context, req *api.GetHabitRequest) (*api.GetHabitResponse, error) {
	logger.Debug("GetHabit called", zap.Int32("id", req.Id))

	habit, err := s.habitService.GetHabit(ctx, int(req.Id))
	if err != nil {
		logger.Error("failed to get habit", zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "habit not found: %v", err)
	}

	return &api.GetHabitResponse{
		Habit: habitToProto(habit),
	}, nil
}

// GetUserHabits получает все привычки пользователя
func (s *HabitServiceServer) GetUserHabits(ctx context.Context, req *api.GetUserHabitsRequest) (*api.GetUserHabitsResponse, error) {
	logger.Debug("GetUserHabits called", zap.Int32("user_id", req.UserId))

	habits, err := s.habitService.GetUserHabits(ctx, int(req.UserId))
	if err != nil {
		logger.Error("failed to get user habits", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get user habits: %v", err)
	}

	protoHabits := make([]*api.Habit, len(habits))
	for i, h := range habits {
		protoHabits[i] = habitToProto(h)
	}

	return &api.GetUserHabitsResponse{
		Habits: protoHabits,
	}, nil
}

// GetActiveHabits получает активные привычки пользователя
func (s *HabitServiceServer) GetActiveHabits(ctx context.Context, req *api.GetActiveHabitsRequest) (*api.GetActiveHabitsResponse, error) {
	logger.Debug("GetActiveHabits called", zap.Int32("user_id", req.UserId))

	habits, err := s.habitService.GetActiveUserHabits(ctx, int(req.UserId))
	if err != nil {
		logger.Error("failed to get active habits", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get active habits: %v", err)
	}

	protoHabits := make([]*api.Habit, len(habits))
	for i, h := range habits {
		protoHabits[i] = habitToProto(h)
	}

	return &api.GetActiveHabitsResponse{
		Habits: protoHabits,
	}, nil
}

// UpdateHabit обновляет привычку
func (s *HabitServiceServer) UpdateHabit(ctx context.Context, req *api.UpdateHabitRequest) (*api.UpdateHabitResponse, error) {
	logger.Debug("UpdateHabit called", zap.Int32("id", req.Id))

	habit, err := s.habitService.GetHabit(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "habit not found")
	}

	habit.Name = req.Name
	if req.Description != "" {
		habit.SetDescription(req.Description)
	}
	if req.Goal != "" {
		habit.SetGoal(req.Goal)
	}

	habit, err = s.habitService.UpdateHabit(ctx, habit)
	if err != nil {
		logger.Error("failed to update habit", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update habit: %v", err)
	}

	return &api.UpdateHabitResponse{
		Habit: habitToProto(habit),
	}, nil
}

// DeleteHabit удаляет (деактивирует) привычку
func (s *HabitServiceServer) DeleteHabit(ctx context.Context, req *api.DeleteHabitRequest) (*api.DeleteHabitResponse, error) {
	logger.Debug("DeleteHabit called", zap.Int32("id", req.Id))

	_, err := s.habitService.DeactivateHabit(ctx, int(req.Id))
	if err != nil {
		logger.Error("failed to delete habit", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete habit: %v", err)
	}

	return &api.DeleteHabitResponse{
		Success: true,
	}, nil
}

// SetWeeklyDays устанавливает дни недели
func (s *HabitServiceServer) SetWeeklyDays(ctx context.Context, req *api.SetWeeklyDaysRequest) (*api.SetWeeklyDaysResponse, error) {
	logger.Debug("SetWeeklyDays called", zap.Int32("habit_id", req.HabitId))

	days := make([]int, len(req.Days))
	for i, d := range req.Days {
		days[i] = int(d)
	}

	habit, err := s.habitService.SetWeeklyDays(ctx, int(req.HabitId), days)
	if err != nil {
		logger.Error("failed to set weekly days", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to set weekly days: %v", err)
	}

	return &api.SetWeeklyDaysResponse{
		Habit: habitToProto(habit),
	}, nil
}

// SetMonthlyDays устанавливает дни месяца
func (s *HabitServiceServer) SetMonthlyDays(ctx context.Context, req *api.SetMonthlyDaysRequest) (*api.SetMonthlyDaysResponse, error) {
	logger.Debug("SetMonthlyDays called", zap.Int32("habit_id", req.HabitId))

	days := make([]int, len(req.Days))
	for i, d := range req.Days {
		days[i] = int(d)
	}

	habit, err := s.habitService.SetMonthlyDays(ctx, int(req.HabitId), days)
	if err != nil {
		logger.Error("failed to set monthly days", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to set monthly days: %v", err)
	}

	return &api.SetMonthlyDaysResponse{
		Habit: habitToProto(habit),
	}, nil
}

// IsScheduledToday проверяет, нужно ли подтверждение сегодня
func (s *HabitServiceServer) IsScheduledToday(ctx context.Context, req *api.IsScheduledTodayRequest) (*api.IsScheduledTodayResponse, error) {
	logger.Debug("IsScheduledToday called", zap.Int32("habit_id", req.HabitId))

	scheduled, err := s.habitService.GetScheduledDaysForToday(ctx, int(req.HabitId))
	if err != nil {
		logger.Error("failed to check scheduled today", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to check scheduled: %v", err)
	}

	return &api.IsScheduledTodayResponse{
		Scheduled: scheduled,
	}, nil
}

// parseIntDays парсит строку "1,3,5" в []int
func parseIntDays(daysStr string) []int {
	parts := strings.Split(daysStr, ",")
	days := make([]int, 0, len(parts))
	for _, part := range parts {
		if day, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			days = append(days, day)
		}
	}
	return days
}
