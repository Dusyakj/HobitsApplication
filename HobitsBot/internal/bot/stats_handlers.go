package bot

import (
	"context"
	"fmt"
	"time"

	"HobitsBot/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleGetToday получает напоминания на сегодня
func (h *BotHandlers) HandleGetToday(message *tgbotapi.Message) {
	// Удаляем предыдущее сообщение с кнопками
	h.deleteOldMessage(message.From.ID, message.Chat.ID)

	ctx := context.Background()
	userID := message.From.ID
	chatID := message.Chat.ID

	// Получаем или создаем пользователя
	user, err := h.userService.GetOrCreateUser(ctx, userID, "", "", "", "")
	if err != nil {
		h.logger.Error("Failed to get user: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось получить данные пользователя"))
		return
	}

	// Получаем напоминания на сегодня с counts
	today := time.Now()
	reminders, err := h.reminderService.GetRemindersByUserAndDate(ctx, int(user.ID), today)
	if err != nil {
		h.logger.Error("Failed to get today reminders: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось получить напоминания"))
		return
	}

	// Фильтруем только незавершённые напоминания
	var incompleteReminders []*service.ReminderResponse
	for _, reminder := range reminders {
		if !reminder.IsCompleted {
			incompleteReminders = append(incompleteReminders, reminder)
		}
	}

	if len(incompleteReminders) == 0 {
		msgConfig := tgbotapi.NewMessage(chatID, "🎉 Все напоминания выполнены или их нет на сегодня!")
		msgConfig.ParseMode = tgbotapi.ModeMarkdown
		msgConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "menu_main"),
			),
		)
		h.bot.api.Send(msgConfig)
		return
	}

	// Форматируем список с inline кнопками
	headerText := fmt.Sprintf("🔔 *Напоминания на сегодня* (%d/%d)\n\n", len(incompleteReminders), len(reminders))

	msgConfig := tgbotapi.NewMessage(chatID, headerText)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown

	// Создаем inline keyboard с кнопками только для незавершённых
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, reminder := range incompleteReminders {
		habit, err := h.habitService.GetHabit(ctx, int(reminder.HabitID))
		if err != nil {
			continue
		}

		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("❌ %s [Выполнить]", habit.Name),
			fmt.Sprintf("habit_%d_complete_from_reminders", reminder.HabitID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	// Добавляем кнопку "Главное меню"
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "menu_main"),
	))

	msgConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.bot.api.Send(msgConfig)
}

// HandleGetStats получает статистику
func (h *BotHandlers) HandleGetStats(message *tgbotapi.Message) {
	// Удаляем предыдущее сообщение с кнопками
	h.deleteOldMessage(message.From.ID, message.Chat.ID)

	ctx := context.Background()
	userID := message.From.ID
	chatID := message.Chat.ID

	// Получаем или создаем пользователя (нужен внутренний ID)
	user, err := h.userService.GetOrCreateUser(ctx, userID, "", "", "", "")
	if err != nil {
		h.logger.Error("Failed to get user: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось получить данные пользователя"))
		return
	}

	// Получаем привычки пользователя
	habits, err := h.habitService.GetActiveHabits(ctx, int(user.ID))
	if err != nil {
		h.logger.Error("Failed to get habits: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось получить статистику"))
		return
	}

	if len(habits) == 0 {
		msgConfig := tgbotapi.NewMessage(chatID, "📭 У вас нет привычек для отслеживания статистики")
		msgConfig.ParseMode = tgbotapi.ModeMarkdown
		msgConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ Создать привычку", "menu_add"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "menu_main"),
			),
		)
		h.bot.api.Send(msgConfig)
		return
	}

	// Показываем список всех привычек со статистикой
	now := time.Now()
	monthAgo := now.AddDate(0, -1, 0)

	var rows [][]tgbotapi.InlineKeyboardButton
	statsText := "📊 *Статистика привычек за месяц*\n\n"

	for _, habit := range habits {
		rate, err := h.logService.GetCompletionRate(ctx, int(habit.ID), monthAgo, now)
		if err != nil {
			rate = 0
		}
		statsText += fmt.Sprintf("*%s*\nВыполнение: %.1f%% | Серия: %d\n\n", habit.Name, rate, habit.CurrentStreak)

		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("📈 %s", habit.Name),
			fmt.Sprintf("habit_%d_stats", habit.ID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "menu_main"),
	))

	msgConfig := tgbotapi.NewMessage(chatID, statsText)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	h.bot.api.Send(msgConfig)
}

// handleHabitStats показывает детальную статистику по привычке
func (h *BotHandlers) handleHabitStats(userID int64, chatID int64, habitID int) {
	ctx := context.Background()

	// Получаем пользователя для проверки прав
	user, err := h.userService.GetOrCreateUser(ctx, userID, "", "", "", "")
	if err != nil {
		h.logger.Error("Failed to get user: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось получить данные"))
		return
	}

	// Получаем привычку
	habit, err := h.habitService.GetHabit(ctx, habitID)
	if err != nil {
		h.logger.Error("Failed to get habit: %v", err)
		h.bot.sendMessage(chatID, FormatError("привычка не найдена"))
		return
	}

	// Проверяем что это привычка текущего пользователя
	if habit.UserID != int32(user.ID) {
		h.bot.sendMessage(chatID, FormatError("у вас нет доступа к этой привычке"))
		return
	}

	// Считаем статистику за месяц
	now := time.Now()
	monthAgo := now.AddDate(0, -1, 0)

	rate, err := h.logService.GetCompletionRate(ctx, habitID, monthAgo, now)
	if err != nil {
		rate = 0
	}

	statsText := fmt.Sprintf(`📊 *Статистика: %s*

Частота: %s
Выполнение за месяц: *%.1f%%*
Текущая серия: *%d дней*
Лучшая серия: *%d дней*

_За последние 30 дней_`,
		habit.Name,
		formatFrequencyText(habit.Frequency),
		rate,
		habit.CurrentStreak,
		habit.BestStreak,
	)

	msgConfig := tgbotapi.NewMessage(chatID, statsText)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Другие привычки", "stats_all"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное", "menu_main"),
		),
	)

	h.bot.api.Send(msgConfig)
}
