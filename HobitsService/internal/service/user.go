package service

import (
	"context"
	"fmt"

	"HobitsService/internal/domain"
	"HobitsService/internal/repository"
)

// UserService сервис для управления пользователями
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService создает новый UserService
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetOrCreateUser получает пользователя или создает нового
func (s *UserService) GetOrCreateUser(ctx context.Context, telegramID int64, firstName, lastName, username, languageCode string) (*domain.User, error) {
	// Пытаемся найти существующего пользователя
	user, err := s.userRepo.GetUserByTelegramID(ctx, telegramID)
	if err == nil {
		return user, nil
	}

	// Создаем нового пользователя
	newUser := domain.NewUser(telegramID, firstName, lastName, username, languageCode)
	return s.userRepo.CreateUser(ctx, newUser)
}

// GetUser получает пользователя по ID
func (s *UserService) GetUser(ctx context.Context, id int) (*domain.User, error) {
	return s.userRepo.GetUserByID(ctx, id)
}

// GetUserByTelegramID получает пользователя по Telegram ID
func (s *UserService) GetUserByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	return s.userRepo.GetUserByTelegramID(ctx, telegramID)
}

// GetAllUsers получает всех пользователей
func (s *UserService) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	return s.userRepo.GetAllUsers(ctx)
}

// UpdateUser обновляет информацию пользователя
func (s *UserService) UpdateUser(ctx context.Context, id int, firstName, lastName, username, languageCode string) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Username = username
	user.LanguageCode = languageCode

	return s.userRepo.UpdateUser(ctx, user)
}

// DeleteUser удаляет пользователя
func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	return s.userRepo.DeleteUser(ctx, id)
}
