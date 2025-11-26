package domain

import (
	"testing"
)

func TestNewUser(t *testing.T) {
	telegramID := int64(123456)
	firstName := "John"
	lastName := "Doe"
	username := "johndoe"
	languageCode := "en"

	user := NewUser(telegramID, firstName, lastName, username, languageCode)

	if user.TelegramID != telegramID {
		t.Errorf("expected telegram_id %d, got %d", telegramID, user.TelegramID)
	}
	if user.FirstName != firstName {
		t.Errorf("expected first_name %s, got %s", firstName, user.FirstName)
	}
	if user.LastName != lastName {
		t.Errorf("expected last_name %s, got %s", lastName, user.LastName)
	}
	if user.Username != username {
		t.Errorf("expected username %s, got %s", username, user.Username)
	}
	if user.LanguageCode != languageCode {
		t.Errorf("expected language_code %s, got %s", languageCode, user.LanguageCode)
	}
}

func TestNewUserWithDifferentLanguages(t *testing.T) {
	tests := []struct {
		languageCode string
	}{
		{"en"},
		{"ru"},
		{"es"},
		{"fr"},
		{"de"},
	}

	for _, tt := range tests {
		user := NewUser(123456, "John", "Doe", "johndoe", tt.languageCode)
		if user.LanguageCode != tt.languageCode {
			t.Errorf("expected language_code %s, got %s", tt.languageCode, user.LanguageCode)
		}
	}
}

func TestNewUserWithEmptyFields(t *testing.T) {
	user := NewUser(123456, "", "", "", "en")

	if user.FirstName != "" {
		t.Errorf("expected empty first_name")
	}
	if user.LastName != "" {
		t.Errorf("expected empty last_name")
	}
	if user.Username != "" {
		t.Errorf("expected empty username")
	}
}

func TestNewUserWithLargeTelegramID(t *testing.T) {
	largeID := int64(9223372036854775807) // Max int64

	user := NewUser(largeID, "John", "Doe", "johndoe", "en")

	if user.TelegramID != largeID {
		t.Errorf("expected large telegram_id %d, got %d", largeID, user.TelegramID)
	}
}

func TestMultipleUsers(t *testing.T) {
	user1 := NewUser(111111, "John", "Doe", "johndoe", "en")
	user2 := NewUser(222222, "Jane", "Smith", "janesmith", "ru")

	if user1.TelegramID == user2.TelegramID {
		t.Errorf("expected different telegram_ids")
	}
	if user1.FirstName == user2.FirstName {
		t.Errorf("expected different first_names")
	}
	if user1.LanguageCode == user2.LanguageCode {
		t.Errorf("expected different language_codes")
	}
}

func TestUserWithSpecialCharacters(t *testing.T) {
	specialUsername := "user.name_123-test"
	specialFirstName := "Иван" // Cyrillic

	user := NewUser(123456, specialFirstName, "Петров", specialUsername, "ru")

	if user.FirstName != specialFirstName {
		t.Errorf("expected first_name %s, got %s", specialFirstName, user.FirstName)
	}
	if user.Username != specialUsername {
		t.Errorf("expected username %s, got %s", specialUsername, user.Username)
	}
}

func TestUserTimestamps(t *testing.T) {
	user := NewUser(123456, "John", "Doe", "johndoe", "en")

	if user.CreatedAt.IsZero() {
		t.Errorf("expected created_at to be set")
	}
	if user.UpdatedAt.IsZero() {
		t.Errorf("expected updated_at to be set")
	}

	// created_at and updated_at should be close (within 1 second)
	if user.CreatedAt.Sub(user.UpdatedAt) > 1 || user.UpdatedAt.Sub(user.CreatedAt) > 1 {
		t.Errorf("expected created_at and updated_at to be similar")
	}
}
