package service

import (
	"context"
	"errors"
	"testing"

	"HobitsService/internal/domain"
)

type mockUserRepository struct {
	users        map[int]*domain.User
	byTelegramID map[int64]*domain.User
}

func (m *mockUserRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	user.ID = len(m.users) + 1
	m.users[user.ID] = user
	m.byTelegramID[user.TelegramID] = user
	return user, nil
}

func (m *mockUserRepository) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUserRepository) GetUserByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	if u, ok := m.byTelegramID[telegramID]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	var users []*domain.User
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, nil
}

func (m *mockUserRepository) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	m.users[user.ID] = user
	return user, nil
}

func (m *mockUserRepository) DeleteUser(ctx context.Context, id int) error {
	if u, ok := m.users[id]; ok {
		delete(m.byTelegramID, u.TelegramID)
		delete(m.users, id)
	}
	return nil
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:        make(map[int]*domain.User),
		byTelegramID: make(map[int64]*domain.User),
	}
}

func TestGetOrCreateUser_CreateNew(t *testing.T) {
	userRepo := newMockUserRepository()
	service := NewUserService(userRepo)
	ctx := context.Background()

	user, err := service.GetOrCreateUser(ctx, 123456, "John", "Doe", "johndoe", "en")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if user == nil {
		t.Errorf("expected user to be returned")
	}
	if user.TelegramID != 123456 {
		t.Errorf("expected telegram_id 123456, got %d", user.TelegramID)
	}
	if user.FirstName != "John" {
		t.Errorf("expected first_name John, got %s", user.FirstName)
	}
}

func TestGetOrCreateUser_GetExisting(t *testing.T) {
	userRepo := newMockUserRepository()
	service := NewUserService(userRepo)
	ctx := context.Background()

	user1, _ := service.GetOrCreateUser(ctx, 123456, "John", "Doe", "johndoe", "en")
	user2, _ := service.GetOrCreateUser(ctx, 123456, "Jane", "Smith", "janesmith", "ru")

	if user1.ID != user2.ID {
		t.Errorf("expected same user for same telegram_id")
	}
	if user2.FirstName != "John" {
		t.Errorf("expected first_name to remain John, got %s", user2.FirstName)
	}
}

func TestGetUser(t *testing.T) {
	userRepo := newMockUserRepository()
	service := NewUserService(userRepo)
	ctx := context.Background()

	created, _ := service.GetOrCreateUser(ctx, 123456, "John", "Doe", "johndoe", "en")
	retrieved, err := service.GetUser(ctx, created.ID)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if retrieved.ID != created.ID {
		t.Errorf("expected user id %d, got %d", created.ID, retrieved.ID)
	}
}

func TestGetUserByTelegramID(t *testing.T) {
	userRepo := newMockUserRepository()
	service := NewUserService(userRepo)
	ctx := context.Background()

	telegramID := int64(123456)
	created, _ := service.GetOrCreateUser(ctx, telegramID, "John", "Doe", "johndoe", "en")
	retrieved, err := service.GetUserByTelegramID(ctx, telegramID)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if retrieved.ID != created.ID {
		t.Errorf("expected same user")
	}
}

func TestUpdateUser(t *testing.T) {
	userRepo := newMockUserRepository()
	service := NewUserService(userRepo)
	ctx := context.Background()

	created, _ := service.GetOrCreateUser(ctx, 123456, "John", "Doe", "johndoe", "en")

	updated, err := service.UpdateUser(ctx, created.ID, "Jane", "Smith", "janesmith", "ru")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if updated.FirstName != "Jane" {
		t.Errorf("expected first_name Jane, got %s", updated.FirstName)
	}
	if updated.LanguageCode != "ru" {
		t.Errorf("expected language_code ru, got %s", updated.LanguageCode)
	}
	// telegram_id should not change
	if updated.TelegramID != 123456 {
		t.Errorf("expected telegram_id to remain unchanged")
	}
}

func TestDeleteUser(t *testing.T) {
	userRepo := newMockUserRepository()
	service := NewUserService(userRepo)
	ctx := context.Background()

	created, _ := service.GetOrCreateUser(ctx, 123456, "John", "Doe", "johndoe", "en")

	err := service.DeleteUser(ctx, created.ID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	retrieved, _ := service.GetUser(ctx, created.ID)
	if retrieved != nil {
		t.Errorf("expected user to be deleted")
	}
}

func TestGetAllUsers(t *testing.T) {
	userRepo := newMockUserRepository()
	service := NewUserService(userRepo)
	ctx := context.Background()

	service.GetOrCreateUser(ctx, 111111, "John", "Doe", "johndoe", "en")
	service.GetOrCreateUser(ctx, 222222, "Jane", "Smith", "janesmith", "ru")
	service.GetOrCreateUser(ctx, 333333, "Bob", "Johnson", "bobjohnson", "es")

	users, err := service.GetAllUsers(ctx)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}
}

func TestMultipleUsersIndependence(t *testing.T) {
	userRepo := newMockUserRepository()
	service := NewUserService(userRepo)
	ctx := context.Background()

	user1, _ := service.GetOrCreateUser(ctx, 111111, "John", "Doe", "johndoe", "en")
	user2, _ := service.GetOrCreateUser(ctx, 222222, "Jane", "Smith", "janesmith", "ru")

	service.UpdateUser(ctx, user1.ID, "Jonathan", "Doe", "johndoe", "en")

	updated2, _ := service.GetUser(ctx, user2.ID)

	if updated2.FirstName != "Jane" {
		t.Errorf("expected user2 to be unaffected by user1 update")
	}
}

func TestUserLanguageCodeVariations(t *testing.T) {
	userRepo := newMockUserRepository()
	service := NewUserService(userRepo)
	ctx := context.Background()

	languageCodes := []string{"en", "ru", "es", "fr", "de", "it", "ja", "zh"}

	for i, code := range languageCodes {
		user, _ := service.GetOrCreateUser(ctx, int64(100000+i), "User", "Test", "usertest", code)
		if user.LanguageCode != code {
			t.Errorf("expected language_code %s, got %s", code, user.LanguageCode)
		}
	}
}

func TestUserIDGeneration(t *testing.T) {
	userRepo := newMockUserRepository()
	service := NewUserService(userRepo)
	ctx := context.Background()

	user1, _ := service.GetOrCreateUser(ctx, 111111, "John", "Doe", "johndoe", "en")
	user2, _ := service.GetOrCreateUser(ctx, 222222, "Jane", "Smith", "janesmith", "ru")

	if user1.ID == user2.ID {
		t.Errorf("expected different user IDs")
	}
	if user1.ID == 0 || user2.ID == 0 {
		t.Errorf("expected non-zero user IDs")
	}
}

func TestUserNonExistent(t *testing.T) {
	userRepo := newMockUserRepository()
	service := NewUserService(userRepo)
	ctx := context.Background()

	user, err := service.GetUser(ctx, 99999)

	if err == nil && user != nil {
		t.Errorf("expected no user to be returned")
	}
}
