package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	pb "HobitsBot/gen/go/hobbits/api/v1"

	"google.golang.org/grpc"
)

type HabitService struct {
	client pb.HabitServiceClient
	logger Logger
}

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

func NewHabitService(conn *grpc.ClientConn, logger Logger) *HabitService {
	return &HabitService{
		client: pb.NewHabitServiceClient(conn),
		logger: logger,
	}
}

type CreateHabitRequest struct {
	UserID      int
	Name        string
	Frequency   string // daily, weekly, monthly
	Description string
	Goal        string
	WeeklyDays  string // "1,3,5" for weekly habits
	MonthlyDays string // "1,15,28" for monthly habits
}

type HabitResponse struct {
	ID            int32
	UserID        int32
	Name          string
	Frequency     string
	Description   string
	Goal          string
	CurrentStreak int32
	BestStreak    int32
	IsActive      bool
	CreatedAt     time.Time
}

// CreateHabit создает новую привычку
func (s *HabitService) CreateHabit(ctx context.Context, req CreateHabitRequest) (*HabitResponse, error) {
	pbReq := &pb.CreateHabitRequest{
		UserId:      int32(req.UserID),
		Name:        req.Name,
		Frequency:   req.Frequency,
		Description: req.Description,
		Goal:        req.Goal,
		WeeklyDays:  req.WeeklyDays,
		MonthlyDays: req.MonthlyDays,
	}

	resp, err := s.client.CreateHabit(ctx, pbReq)
	if err != nil {
		s.logger.Error("failed to create habit: %v", err)
		return nil, fmt.Errorf("не удалось создать привычку: %w", err)
	}

	return s.pbHabitToResponse(resp.Habit), nil
}

// GetHabit получает информацию о привычке
func (s *HabitService) GetHabit(ctx context.Context, habitID int) (*HabitResponse, error) {
	req := &pb.GetHabitRequest{
		Id: int32(habitID),
	}

	resp, err := s.client.GetHabit(ctx, req)
	if err != nil {
		s.logger.Error("failed to get habit: %v", err)
		return nil, fmt.Errorf("привычка не найдена: %w", err)
	}

	return s.pbHabitToResponse(resp.Habit), nil
}

// GetUserHabits получает все привычки пользователя
func (s *HabitService) GetUserHabits(ctx context.Context, userID int) ([]*HabitResponse, error) {
	req := &pb.GetUserHabitsRequest{
		UserId: int32(userID),
	}

	resp, err := s.client.GetUserHabits(ctx, req)
	if err != nil {
		s.logger.Error("failed to get user habits: %v", err)
		return nil, fmt.Errorf("не удалось получить привычки: %w", err)
	}

	habits := make([]*HabitResponse, len(resp.Habits))
	for i, h := range resp.Habits {
		habits[i] = s.pbHabitToResponse(h)
	}

	return habits, nil
}

// GetActiveHabits получает активные привычки пользователя
func (s *HabitService) GetActiveHabits(ctx context.Context, userID int) ([]*HabitResponse, error) {
	req := &pb.GetActiveHabitsRequest{
		UserId: int32(userID),
	}

	resp, err := s.client.GetActiveHabits(ctx, req)
	if err != nil {
		s.logger.Error("failed to get active habits: %v", err)
		return nil, fmt.Errorf("не удалось получить активные привычки: %w", err)
	}

	habits := make([]*HabitResponse, len(resp.Habits))
	for i, h := range resp.Habits {
		habits[i] = s.pbHabitToResponse(h)
	}

	return habits, nil
}

// UpdateHabit обновляет информацию о привычке
func (s *HabitService) UpdateHabit(ctx context.Context, habitID int, name, description, goal string) (*HabitResponse, error) {
	req := &pb.UpdateHabitRequest{
		Id:          int32(habitID),
		Name:        name,
		Description: description,
		Goal:        goal,
	}

	resp, err := s.client.UpdateHabit(ctx, req)
	if err != nil {
		s.logger.Error("failed to update habit: %v", err)
		return nil, fmt.Errorf("не удалось обновить привычку: %w", err)
	}

	return s.pbHabitToResponse(resp.Habit), nil
}

// DeleteHabit деактивирует привычку
func (s *HabitService) DeleteHabit(ctx context.Context, habitID int) error {
	req := &pb.DeactivateHabitRequest{
		Id: int32(habitID),
	}

	_, err := s.client.DeactivateHabit(ctx, req)
	if err != nil {
		s.logger.Error("failed to deactivate habit: %v", err)
		return fmt.Errorf("не удалось деактивировать привычку: %w", err)
	}

	return nil
}

// ActivateHabit активирует привычку
func (s *HabitService) ActivateHabit(ctx context.Context, habitID int) error {
	req := &pb.ActivateHabitRequest{
		Id: int32(habitID),
	}

	_, err := s.client.ActivateHabit(ctx, req)
	if err != nil {
		s.logger.Error("failed to activate habit: %v", err)
		return fmt.Errorf("не удалось активировать привычку: %w", err)
	}

	return nil
}

// SetWeeklyDays устанавливает дни недели для еженедельной привычки
func (s *HabitService) SetWeeklyDays(ctx context.Context, habitID int, days string) error {
	// Parse days string (e.g., "1,3,5") to int32 array
	dayInts := []int32{}
	if days != "" {
		for _, d := range strings.Split(days, ",") {
			if v, err := strconv.Atoi(strings.TrimSpace(d)); err == nil {
				dayInts = append(dayInts, int32(v))
			}
		}
	}

	req := &pb.SetWeeklyDaysRequest{
		HabitId: int32(habitID),
		Days:    dayInts,
	}

	_, err := s.client.SetWeeklyDays(ctx, req)
	if err != nil {
		s.logger.Error("failed to set weekly days: %v", err)
		return fmt.Errorf("не удалось установить дни: %w", err)
	}

	return nil
}

// SetMonthlyDays устанавливает дни месяца для ежемесячной привычки
func (s *HabitService) SetMonthlyDays(ctx context.Context, habitID int, days string) error {
	// Parse days string (e.g., "1,15,28") to int32 array
	dayInts := []int32{}
	if days != "" {
		for _, d := range strings.Split(days, ",") {
			if v, err := strconv.Atoi(strings.TrimSpace(d)); err == nil {
				dayInts = append(dayInts, int32(v))
			}
		}
	}

	req := &pb.SetMonthlyDaysRequest{
		HabitId: int32(habitID),
		Days:    dayInts,
	}

	_, err := s.client.SetMonthlyDays(ctx, req)
	if err != nil {
		s.logger.Error("failed to set monthly days: %v", err)
		return fmt.Errorf("не удалось установить дни месяца: %w", err)
	}

	return nil
}

// pbHabitToResponse преобразует protobuf привычку в ответ
func (s *HabitService) pbHabitToResponse(pb *pb.Habit) *HabitResponse {
	createdAt := time.Now()
	if pb.CreatedAt != nil {
		createdAt = pb.CreatedAt.AsTime()
	}

	return &HabitResponse{
		ID:            pb.Id,
		UserID:        pb.UserId,
		Name:          pb.Name,
		Frequency:     pb.Frequency,
		Description:   pb.Description,
		Goal:          pb.Goal,
		CurrentStreak: pb.CurrentStreak,
		BestStreak:    pb.BestStreak,
		IsActive:      pb.IsActive,
		CreatedAt:     createdAt,
	}
}
