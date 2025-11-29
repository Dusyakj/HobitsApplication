package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleStart обрабатывает команду /start
func (h *BotHandlers) HandleStart(message *tgbotapi.Message) {
	ctx := context.Background()
	userID := message.From.ID
	chatID := message.Chat.ID

	// Создаем или получаем пользователя
	user, err := h.userService.GetOrCreateUser(
		ctx,
		userID,
		message.From.FirstName,
		message.From.LastName,
		message.From.UserName,
		message.From.LanguageCode,
	)

	if err != nil {
		h.logger.Error("Failed to create/get user: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось инициализировать профиль"))
		return
	}

	h.logger.Info("User %d (%s) started bot", user.ID, user.FirstName)

	msg := FormatWelcomeMessage(message.From.FirstName)
	msgConfig := tgbotapi.NewMessage(chatID, msg)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = MainMenuInlineKeyboard() // Inline клавиатура главного меню

	h.bot.api.Send(msgConfig)
}

// handleMenuSelect обрабатывает выбор пунктов меню
func (h *BotHandlers) handleMenuSelect(userID int64, chatID int64, menu string) {
	// Получаем и удаляем старое сообщение
	oldChatID, oldMessageID := h.contextManager.GetMessageID(userID)
	if oldMessageID > 0 && oldChatID == chatID {
		go func() {
			err := h.bot.DeleteMessage(oldChatID, oldMessageID)
			if err != nil {
				h.logger.Debug("Could not delete old message %d: %v", oldMessageID, err)
			}
		}()
	}

	h.contextManager.ClearContext(userID)

	switch menu {
	case "main":
		mainMenuText := `🏠 *Главное меню*

Добро пожаловать в трекер привычек! 🎯

*Как это работает:*
1️⃣ Создайте привычку и выберите частоту (ежедневно, по дням недели или месяца)
2️⃣ Каждый день бот напомнит о ваших задачах
3️⃣ Отмечайте выполненные привычки и следите за серией
4️⃣ Анализируйте статистику и мотивируйте себя! 📈

_Серия дней — это количество дней подряд, когда вы выполняли привычку_`
		msgConfig := tgbotapi.NewMessage(chatID, mainMenuText)
		msgConfig.ReplyMarkup = MainMenuInlineKeyboard()
		msgConfig.ParseMode = tgbotapi.ModeMarkdown
		h.bot.api.Send(msgConfig)
	case "habits":
		h.HandleGetHabits(&tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}})
	case "archive":
		h.handleHabitArchive(userID, chatID)
	case "add":
		h.HandleAddHabit(&tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}})
	case "today":
		h.HandleGetToday(&tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}})
	case "stats":
		h.HandleGetStats(&tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}})
	case "help":
		helpText := `📖 *Доступные команды:*

/start - Начало работы
/habits - Мои привычки
/add - Добавить новую привычку
/today - Напоминания на сегодня
/stats - Статистика привычек
/help - Справка

*Примеры использования:*
• Нажмите ➕ Добавить и следуйте инструкциям
• Используйте кнопки для управления
• 🔔 Напоминания показывает все задачи на сегодня`
		msgConfig := tgbotapi.NewMessage(chatID, helpText)
		msgConfig.ParseMode = tgbotapi.ModeMarkdown
		msgConfig.ReplyMarkup = MainMenuInlineKeyboard()
		h.bot.api.Send(msgConfig)
	}
}
