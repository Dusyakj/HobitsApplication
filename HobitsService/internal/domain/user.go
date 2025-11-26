package domain

import "time"

// User представляет пользователя приложения
type User struct {
	ID           int       `db:"id"`
	TelegramID   int64     `db:"telegram_id"`
	FirstName    string    `db:"first_name"`
	LastName     string    `db:"last_name"`
	Username     string    `db:"username"`
	LanguageCode string    `db:"language_code"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// NewUser создает нового пользователя из данных Telegram
func NewUser(telegramID int64, firstName, lastName, username, languageCode string) *User {
	now := time.Now()
	return &User{
		TelegramID:   telegramID,
		FirstName:    firstName,
		LastName:     lastName,
		Username:     username,
		LanguageCode: languageCode,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
