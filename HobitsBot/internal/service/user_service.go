package service

import (
	"context"
	"fmt"
	"time"

	pb "HobitsBot/gen/go/hobbits/api/v1"

	"google.golang.org/grpc"
)

type UserService struct {
	client pb.UserServiceClient
	logger Logger
}

type UserResponse struct {
	ID           int32
	TelegramID   int64
	FirstName    string
	LastName     string
	Username     string
	LanguageCode string
	CreatedAt    time.Time
}

func NewUserService(conn *grpc.ClientConn, logger Logger) *UserService {
	return &UserService{
		client: pb.NewUserServiceClient(conn),
		logger: logger,
	}
}

// GetOrCreateUser получает или создает пользователя
func (s *UserService) GetOrCreateUser(
	ctx context.Context,
	telegramID int64,
	firstName, lastName, username, languageCode string,
) (*UserResponse, error) {
	req := &pb.GetOrCreateUserRequest{
		TelegramId:   telegramID,
		FirstName:    firstName,
		LastName:     lastName,
		Username:     username,
		LanguageCode: languageCode,
	}

	resp, err := s.client.GetOrCreateUser(ctx, req)
	if err != nil {
		s.logger.Error("failed to get or create user: %v", err)
		return nil, fmt.Errorf("не удалось получить или создать пользователя: %w", err)
	}

	return s.pbUserToResponse(resp.User), nil
}

// GetUser получает информацию о пользователе
func (s *UserService) GetUser(ctx context.Context, userID int) (*UserResponse, error) {
	req := &pb.GetUserRequest{
		Id: int32(userID),
	}

	resp, err := s.client.GetUser(ctx, req)
	if err != nil {
		s.logger.Error("failed to get user: %v", err)
		return nil, fmt.Errorf("пользователь не найден: %w", err)
	}

	return s.pbUserToResponse(resp.User), nil
}

// UpdateUser обновляет информацию о пользователе
func (s *UserService) UpdateUser(
	ctx context.Context,
	userID int,
	firstName, lastName, username, languageCode string,
) (*UserResponse, error) {
	req := &pb.UpdateUserRequest{
		Id:           int32(userID),
		FirstName:    firstName,
		LastName:     lastName,
		Username:     username,
		LanguageCode: languageCode,
	}

	resp, err := s.client.UpdateUser(ctx, req)
	if err != nil {
		s.logger.Error("failed to update user: %v", err)
		return nil, fmt.Errorf("не удалось обновить пользователя: %w", err)
	}

	return s.pbUserToResponse(resp.User), nil
}

// pbUserToResponse преобразует protobuf пользователя в ответ
func (s *UserService) pbUserToResponse(pb *pb.User) *UserResponse {
	createdAt := time.Now()
	if pb.CreatedAt != nil {
		createdAt = pb.CreatedAt.AsTime()
	}

	return &UserResponse{
		ID:           pb.Id,
		TelegramID:   pb.TelegramId,
		FirstName:    pb.FirstName,
		LastName:     pb.LastName,
		Username:     pb.Username,
		LanguageCode: pb.LanguageCode,
		CreatedAt:    createdAt,
	}
}
