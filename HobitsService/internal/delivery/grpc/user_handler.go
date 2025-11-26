package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	api "HobitsService/gen/go/HobitsService/gen/go/hobbits/api/v1"
	"HobitsService/internal/domain"
	"HobitsService/internal/logger"
	"HobitsService/internal/service"
)

// UserServiceServer реализация UserService
type UserServiceServer struct {
	api.UnimplementedUserServiceServer
	userService *service.UserService
}

// NewUserServiceServer создает новый UserServiceServer
func NewUserServiceServer(userService *service.UserService) *UserServiceServer {
	return &UserServiceServer{
		userService: userService,
	}
}

// GetOrCreateUser получает или создает пользователя
func (s *UserServiceServer) GetOrCreateUser(ctx context.Context, req *api.GetOrCreateUserRequest) (*api.GetOrCreateUserResponse, error) {
	logger.Debug("GetOrCreateUser called", zap.Int64("telegram_id", req.TelegramId))

	user, err := s.userService.GetOrCreateUser(
		ctx,
		req.TelegramId,
		req.FirstName,
		req.LastName,
		req.Username,
		req.LanguageCode,
	)
	if err != nil {
		logger.Error("failed to get or create user", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get or create user: %v", err)
	}

	return &api.GetOrCreateUserResponse{
		User:    domainUserToProto(user),
		Created: user.ID > 0, // Если ID заполнен, значит существующий пользователь
	}, nil
}

// GetUser получает пользователя по ID
func (s *UserServiceServer) GetUser(ctx context.Context, req *api.GetUserRequest) (*api.GetUserResponse, error) {
	logger.Debug("GetUser called", zap.Int32("id", req.Id))

	user, err := s.userService.GetUser(ctx, int(req.Id))
	if err != nil {
		logger.Error("failed to get user", zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &api.GetUserResponse{
		User: domainUserToProto(user),
	}, nil
}

// UpdateUser обновляет пользователя
func (s *UserServiceServer) UpdateUser(ctx context.Context, req *api.UpdateUserRequest) (*api.UpdateUserResponse, error) {
	logger.Debug("UpdateUser called", zap.Int32("id", req.Id))

	user, err := s.userService.UpdateUser(
		ctx,
		int(req.Id),
		req.FirstName,
		req.LastName,
		req.Username,
		req.LanguageCode,
	)
	if err != nil {
		logger.Error("failed to update user", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &api.UpdateUserResponse{
		User: domainUserToProto(user),
	}, nil
}

// domainUserToProto преобразует domain модель в proto сообщение
func domainUserToProto(user *domain.User) *api.User {
	return &api.User{
		Id:           int32(user.ID),
		TelegramId:   user.TelegramID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Username:     user.Username,
		LanguageCode: user.LanguageCode,
		CreatedAt:    timestamppb.New(user.CreatedAt),
		UpdatedAt:    timestamppb.New(user.UpdatedAt),
	}
}
