package grpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	api "HobitsService/gen/go/HobitsService/gen/go/hobbits/api/v1"
	"HobitsService/internal/domain"
)

// Converters from domain models to proto messages

func habitToProto(h *domain.Habit) *api.Habit {
	habit := &api.Habit{
		Id:            int32(h.ID),
		UserId:        int32(h.UserID),
		Name:          h.Name,
		Frequency:     string(h.Frequency),
		CurrentStreak: int32(h.CurrentStreak),
		BestStreak:    int32(h.BestStreak),
		IsActive:      h.IsActive,
		IsCompleted:   h.IsCompleted,
		CreatedAt:     timestamppb.New(h.CreatedAt),
		UpdatedAt:     timestamppb.New(h.UpdatedAt),
	}

	if h.Description.Valid {
		habit.Description = h.Description.String
	}
	if h.Goal.Valid {
		habit.Goal = h.Goal.String
	}
	if h.WeeklyDays.Valid {
		habit.WeeklyDays = h.WeeklyDays.String
	}
	if h.MonthlyDays.Valid {
		habit.MonthlyDays = h.MonthlyDays.String
	}
	if h.LastCompletedDate.Valid {
		habit.LastCompletedDate = timestamppb.New(h.LastCompletedDate.Time)
	}
	if h.LastCheckedDate.Valid {
		habit.LastCheckedDate = timestamppb.New(h.LastCheckedDate.Time)
	}
	if h.CompletedAt.Valid {
		habit.CompletedAt = timestamppb.New(h.CompletedAt.Time)
	}

	return habit
}

func habitLogToProto(log *domain.HabitLog) *api.HabitLog {
	return &api.HabitLog{
		Id:         int32(log.ID),
		HabitId:    int32(log.HabitID),
		UserId:     int32(log.UserID),
		LoggedAt:   timestamppb.New(log.LoggedAt),
		LoggedDate: log.LoggedDate.Format("2006-01-02"),
	}
}

func habitReminderToProto(r *domain.HabitReminder) *api.HabitReminder {
	reminder := &api.HabitReminder{
		Id:          int32(r.ID),
		HabitId:     int32(r.HabitID),
		UserId:      int32(r.UserID),
		IsCompleted: r.IsCompleted,
		SentAt:      timestamppb.New(r.SentAt),
	}

	if r.ReminderDate.Valid {
		reminder.ReminderDate = timestamppb.New(r.ReminderDate.Time)
	}

	return reminder
}
