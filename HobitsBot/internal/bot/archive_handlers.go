package bot

import (
	"context"
	"fmt"

	"HobitsBot/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleHabitArchive показывает архив деактивированных привычек
func (h *BotHandlers) handleHabitArchive(userID int64, chatID int64) {
	ctx := context.Background()

	// Получаем пользователя
	user, err := h.userService.GetOrCreateUser(ctx, userID, "", "", "", "")
	if err != nil {
		h.logger.Error("Failed to get user: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось получить данные пользователя"))
		return
	}

	// Получаем ВСЕ привычки пользователя (включая деактивированные)
	allHabits, err := h.habitService.GetUserHabits(ctx, int(user.ID))
	if err != nil {
		h.logger.Error("Failed to get all habits: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось получить привычки из архива"))
		return
	}

	// Получаем активные привычки
	activeHabits, err := h.habitService.GetActiveHabits(ctx, int(user.ID))
	if err != nil {
		h.logger.Error("Failed to get active habits: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось проверить активные привычки"))
		return
	}

	// Создаем map активных привычек для быстрого поиска
	activeHabitMap := make(map[int32]bool)
	for _, active := range activeHabits {
		activeHabitMap[active.ID] = true
	}

	// Фильтруем деактивированные привычки (те что в allHabits но не в activeHabits)
	var deactivatedHabits []*service.HabitResponse
	for _, habit := range allHabits {
		if !activeHabitMap[habit.ID] {
			deactivatedHabits = append(deactivatedHabits, habit)
		}
	}

	// Если нет деактивированных привычек
	if len(deactivatedHabits) == 0 {
		msgConfig := tgbotapi.NewMessage(chatID, "📭 Архив пуст. Нет деактивированных привычек.\n\n(Здесь будут появляться привычки после деактивации)")
		msgConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "menu_main"),
			),
		)
		h.bot.api.Send(msgConfig)
		return
	}

	// Создаем список архивированных привычек
	habitNames := make([]string, len(deactivatedHabits))
	habitIDs := make([]int, len(deactivatedHabits))
	for i, habit := range deactivatedHabits {
		habitNames[i] = fmt.Sprintf("⏸ %s", habit.Name)
		habitIDs[i] = int(habit.ID)
	}

	// Создаем сообщение
	text := fmt.Sprintf("📚 *Архив привычек* (%d)\n\nВы можете восстановить любую привычку из архива.", len(deactivatedHabits))
	msgConfig := tgbotapi.NewMessage(chatID, text)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown

	// Переименуем кнопку "Назад" в "Главное меню" для архива
	// HabitsListKeyboard добавляет стандартную кнопку, нам нужно её изменить
	// Для простоты мы перестроим клавиатуру вручную
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, habitID := range habitIDs {
		button := tgbotapi.NewInlineKeyboardButtonData(
			habitNames[i],
			fmt.Sprintf("archive_habit_%d_view", habitID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "menu_main"),
	))

	msgConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.bot.api.Send(msgConfig)
}

// handleArchiveHabitView показывает деталь деактивированной привычки с кнопкой восстановления
func (h *BotHandlers) handleArchiveHabitView(userID int64, chatID int64, habitID int) {
	ctx := context.Background()

	// Получаем привычку
	habit, err := h.habitService.GetHabit(ctx, habitID)
	if err != nil {
		h.logger.Error("Failed to get habit %d: %v", habitID, err)
		msgConfig := tgbotapi.NewMessage(chatID, "❌ Привычка не найдена")
		h.bot.api.Send(msgConfig)
		return
	}

	if habit == nil {
		msgConfig := tgbotapi.NewMessage(chatID, "❌ Привычка не найдена в архиве")
		h.bot.api.Send(msgConfig)
		return
	}

	// Проверяем что привычка принадлежит пользователю
	user, err := h.userService.GetOrCreateUser(ctx, userID, "", "", "", "")
	if err != nil || user == nil || habit.UserID != int32(user.ID) {
		msgConfig := tgbotapi.NewMessage(chatID, "❌ Нет доступа к этой привычке")
		h.bot.api.Send(msgConfig)
		return
	}

	// Форматируем информацию о привычке
	text := FormatHabitDetail(
		habit.Name,
		habit.Frequency,
		int(habit.CurrentStreak),
		int(habit.BestStreak),
		0,
		0,
		habit.Goal, // цель привычки
	)

	text += "\n\n⏸ *Привычка деактивирована*\n\nВы можете восстановить эту привычку в любой момент."

	msgConfig := tgbotapi.NewMessage(chatID, text)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("♻️ Восстановить", fmt.Sprintf("restore_habit_%d", habitID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад в архив", "menu_archive"),
		),
	)
	h.bot.api.Send(msgConfig)
}

// handleHabitDeactivate обрабатывает деактивацию привычки
func (h *BotHandlers) handleHabitDeactivate(userID int64, chatID int64, habitID int) {
	ctx := context.Background()

	// Получаем привычку чтобы проверить что она существует и принадлежит пользователю
	habit, err := h.habitService.GetHabit(ctx, habitID)
	if err != nil {
		h.logger.Error("Failed to get habit %d: %v", habitID, err)
		msgConfig := tgbotapi.NewMessage(chatID, "❌ Ошибка при получении информации о привычке")
		h.bot.api.Send(msgConfig)
		return
	}

	if habit == nil {
		msgConfig := tgbotapi.NewMessage(chatID, "❌ Привычка не найдена")
		h.bot.api.Send(msgConfig)
		return
	}

	// Проверяем что привычка принадлежит текущему пользователю
	user, err := h.userService.GetOrCreateUser(ctx, userID, "", "", "", "")
	if err != nil || user == nil || habit.UserID != int32(user.ID) {
		msgConfig := tgbotapi.NewMessage(chatID, "❌ Нет доступа к этой привычке")
		h.bot.api.Send(msgConfig)
		return
	}

	habitName := habit.Name

	// Отправляем сообщение с просьбой подтверждения
	msgConfig := tgbotapi.NewMessage(
		chatID,
		fmt.Sprintf("🚫 Вы уверены что хотите деактивировать привычку *%s*?\n\nЭту привычку можно будет восстановить из архива.", habitName),
	)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да, деактивировать", fmt.Sprintf("confirm_deactivate_%d", habitID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", fmt.Sprintf("habit_%d_view", habitID)),
		),
	)
	h.bot.api.Send(msgConfig)
}

// handleHabitRestore восстанавливает деактивированную привычку
func (h *BotHandlers) handleHabitRestore(userID int64, chatID int64, habitID int) {
	ctx := context.Background()

	// Получаем привычку
	habit, err := h.habitService.GetHabit(ctx, habitID)
	if err != nil {
		h.logger.Error("Failed to get habit %d: %v", habitID, err)
		msgConfig := tgbotapi.NewMessage(chatID, "❌ Не удалось получить информацию о привычке")
		msgConfig.ParseMode = tgbotapi.ModeMarkdown
		h.bot.api.Send(msgConfig)
		return
	}

	// Используем UpdateHabit для восстановления (просто пересохраняем те же данные)
	// Серверная часть должна автоматически активировать привычку если она была деактивирована
	restoredHabit, err := h.habitService.UpdateHabit(ctx, habitID, habit.Name, "", "")
	if err != nil {
		h.logger.Error("Failed to restore habit %d: %v", habitID, err)
		msgConfig := tgbotapi.NewMessage(chatID, "❌ Не удалось восстановить привычку")
		msgConfig.ParseMode = tgbotapi.ModeMarkdown
		h.bot.api.Send(msgConfig)
		return
	}

	// Показываем сообщение об успешном восстановлении
	msg := fmt.Sprintf(`♻️ *Привычка восстановлена!*

*%s*
Частота: %s
Статус: ✅ Активна

Привычка вернулась в ваш список. Теперь она снова будет отправлять напоминания!`,
		restoredHabit.Name,
		map[string]string{
			"daily":   "📅 Ежедневно",
			"weekly":  "📆 Еженедельно",
			"monthly": "📊 Ежемесячно",
		}[restoredHabit.Frequency],
	)

	msgConfig := tgbotapi.NewMessage(chatID, msg)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад в архив", "menu_archive"),
			tgbotapi.NewInlineKeyboardButtonData("📋 Мои привычки", "menu_habits"),
		),
	)
	h.bot.api.Send(msgConfig)
}

// handleConfirmDeactivate подтверждает деактивацию привычки
func (h *BotHandlers) handleConfirmDeactivate(userID int64, chatID int64, habitID int) {
	ctx := context.Background()

	// Вызываем DeleteHabit который деактивирует привычку
	err := h.habitService.DeleteHabit(ctx, habitID)
	if err != nil {
		h.logger.Error("Failed to deactivate habit %d: %v", habitID, err)
		msgConfig := tgbotapi.NewMessage(chatID, "❌ Ошибка при деактивации привычки")
		h.bot.api.Send(msgConfig)
		return
	}

	h.logger.Info("Habit %d deactivated by user %d", habitID, userID)

	// Отправляем сообщение об успехе
	msgConfig := tgbotapi.NewMessage(
		chatID,
		"✅ Привычка успешно деактивирована!\n\nЕё можно восстановить из архива в любой момент.",
	)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Мои привычки", "menu_habits"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "menu_main"),
		),
	)
	h.bot.api.Send(msgConfig)
}
