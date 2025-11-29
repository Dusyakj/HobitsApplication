package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"HobitsBot/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleAddHabit обрабатывает добавление новой привычки
func (h *BotHandlers) HandleAddHabit(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	// Очищаем предыдущий контекст
	h.contextManager.ClearContext(userID)

	// Устанавливаем состояние "ожидание названия привычки"
	h.contextManager.SetState(userID, "waiting_habit_name")

	msgConfig := tgbotapi.NewMessage(
		chatID,
		"✏️ Введите название новой привычки:\n\nНапример: Зарядка, Чтение, Медитация",
	)
	msgConfig.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}

	h.bot.api.Send(msgConfig)
}

// HandleHabitName обрабатывает введение названия привычки
func (h *BotHandlers) HandleHabitName(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID
	habitName := strings.TrimSpace(message.Text)

	if habitName == "" {
		h.bot.sendMessage(chatID, "❌ Название не может быть пустым. Попробуйте снова.")
		return
	}

	if len(habitName) > 100 {
		h.bot.sendMessage(chatID, "❌ Название слишком длинное (макс. 100 символов)")
		return
	}

	// Сохраняем название привычки в контексте
	h.contextManager.SetData(userID, "habit_name", habitName)

	// Устанавливаем состояние "ожидание выбора частоты"
	h.contextManager.SetState(userID, "waiting_frequency")

	// Отправляем сообщение с выбором частоты
	msgConfig := tgbotapi.NewMessage(
		chatID,
		fmt.Sprintf("✨ Отлично! Привычка: *%s*\n\nВыберите частоту:", habitName),
	)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = FrequencyKeyboard()

	h.bot.api.Send(msgConfig)
}

// HandleGetHabits получает список привычек пользователя
func (h *BotHandlers) HandleGetHabits(message *tgbotapi.Message) {
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

	// Получаем активные привычки
	habits, err := h.habitService.GetActiveHabits(ctx, int(user.ID))
	if err != nil {
		h.logger.Error("Failed to get habits: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось получить привычки"))
		return
	}

	if len(habits) == 0 {
		msgConfig := tgbotapi.NewMessage(chatID, FormatHabitList([]string{}))
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

	// Создаем список привычек и клавиатуру
	habitNames := make([]string, len(habits))
	habitIDs := make([]int, len(habits))

	for i, habit := range habits {
		habitNames[i] = habit.Name
		habitIDs[i] = int(habit.ID)
	}

	text := FormatHabitList(habitNames)
	msgConfig := tgbotapi.NewMessage(chatID, text)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = HabitsListKeyboard(habitIDs, habitNames)

	h.bot.api.Send(msgConfig)
}

// HandleHabitAction обрабатывает действия с привычкой
func (h *BotHandlers) HandleHabitAction(message *tgbotapi.Message, habitID int, action string) {
	ctx := context.Background()
	userID := message.From.ID
	chatID := message.Chat.ID

	// Получаем информацию о привычке
	habit, err := h.habitService.GetHabit(ctx, habitID)
	if err != nil {
		h.logger.Error("Failed to get habit: %v", err)
		h.bot.sendMessage(chatID, FormatError("привычка не найдена"))
		return
	}

	switch action {
	case "complete":
		h.handleHabitCompleteWithComment(userID, chatID, habitID)
	case "skip":
		h.handleHabitSkip(ctx, chatID, userID, habit)
	case "view":
		h.handleHabitView(ctx, chatID, userID, habit)
	}
}

func (h *BotHandlers) handleHabitComplete(
	ctx context.Context,
	chatID int64,
	userID int64,
	habit *service.HabitResponse,
) {
	// Логируем выполнение привычки
	_, err := h.logService.LogCompletion(ctx, int(habit.ID), int(userID), "")
	if err != nil {
		h.logger.Error("Failed to log completion: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось сохранить выполнение"))
		return
	}

	// Получаем обновленную информацию о привычке
	updated, _ := h.habitService.GetHabit(ctx, int(habit.ID))
	if updated == nil {
		updated = habit
	}

	text := FormatHabitCompleted(habit.Name, updated.CurrentStreak)
	h.bot.sendMessage(chatID, text)

	h.logger.Info("User %d completed habit %d", userID, habit.ID)
}

func (h *BotHandlers) handleHabitSkip(
	ctx context.Context,
	chatID int64,
	userID int64,
	habit *service.HabitResponse,
) {
	text := FormatHabitSkipped(habit.Name)
	h.bot.sendMessage(chatID, text)

	h.logger.Info("User %d skipped habit %d", userID, habit.ID)
}

func (h *BotHandlers) handleHabitView(
	ctx context.Context,
	chatID int64,
	userID int64,
	habit *service.HabitResponse,
) {
	// Получаем пользователя для получения его ID
	user, err := h.userService.GetOrCreateUser(ctx, userID, "", "", "", "")
	if err != nil {
		h.logger.Error("Failed to get user: %v", err)
		h.bot.sendMessage(chatID, FormatError("ошибка получения данных"))
		return
	}

	// Проверяем напоминания на сегодня
	today := time.Now()
	reminders, err := h.reminderService.GetRemindersByUserAndDate(ctx, int(user.ID), today)
	if err != nil {
		h.logger.Error("Failed to get today reminders: %v", err)
		reminders = []*service.ReminderResponse{}
	}

	// Проверяем есть ли напоминание для этой привычки на сегодня
	var habitReminder *service.ReminderResponse
	for _, reminder := range reminders {
		if reminder.HabitID == habit.ID {
			habitReminder = reminder
			break
		}
	}

	// Форматируем текст с информацией о привычке
	text := FormatHabitDetail(
		habit.Name,
		habit.Frequency,
		int(habit.CurrentStreak),
		int(habit.BestStreak),
		0, // completed
		0, // total
		habit.Goal, // цель привычки
	)

	// Добавляем информацию о статусе на сегодня
	if habitReminder != nil {
		if habitReminder.IsCompleted {
			text += "\n\n✅ *Выполнена сегодня!*"
		} else {
			text += "\n\n❌ *Не выполнена сегодня*"
		}
	} else {
		text += "\n\n⏸ *Не запланирована на сегодня*"
	}

	msgConfig := tgbotapi.NewMessage(chatID, text)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown

	// Показываем кнопки только если привычка активна на сегодня И ещё не выполнена
	if habitReminder != nil && !habitReminder.IsCompleted {
		// Кнопки выполнить/пропустить только если не выполнена
		msgConfig.ReplyMarkup = HabitActionsKeyboardWithDescription(int(habit.ID), habit.Description)
	} else {
		// Если привычка не на сегодня или уже выполнена, показываем кнопку деактивации и назад
		msgConfig.ReplyMarkup = HabitDetailKeyboardWithDescription(int(habit.ID), habit.Description)
	}

	h.bot.api.Send(msgConfig)
}

// handleHabitCompleteWithComment обрабатывает выполнение привычки и спрашивает комментарий
func (h *BotHandlers) handleHabitCompleteWithComment(userID int64, chatID int64, habitID int) {
	// Удаляем предыдущее сообщение с кнопками
	h.deleteOldMessage(userID, chatID)

	ctx := context.Background()

	// Получаем или создаем пользователя
	user, err := h.userService.GetOrCreateUser(ctx, userID, "", "", "", "")
	if err != nil {
		h.logger.Error("Failed to get user: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось получить данные пользователя"))
		return
	}

	// Получаем привычку
	habit, err := h.habitService.GetHabit(ctx, habitID)
	if err != nil {
		h.logger.Error("Failed to get habit: %v", err)
		h.bot.sendMessage(chatID, FormatError("привычка не найдена"))
		return
	}

	// Логируем выполнение без комментария сначала
	_, err = h.logService.LogCompletion(ctx, habitID, int(user.ID), "")
	if err != nil {
		h.logger.Error("Failed to log completion: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось зафиксировать выполнение"))
		return
	}

	// Обновляем привычку
	updated, _ := h.habitService.GetHabit(ctx, habitID)

	// Показываем сообщение об успехе с счётчиком серии
	successText := fmt.Sprintf("✅ *%s* выполнена!\n\n🔥 Текущая серия: *%d дней*\n🏆 Лучшая серия: *%d дней*",
		habit.Name,
		updated.CurrentStreak,
		updated.BestStreak,
	)

	msgConfig := tgbotapi.NewMessage(chatID, successText)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown

	// Спрашиваем про комментарий
	msgConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Добавить комментарий", fmt.Sprintf("habit_%d_add_comment", habitID)),
			tgbotapi.NewInlineKeyboardButtonData("✅ Готово", "menu_main"),
		),
	)

	h.bot.api.Send(msgConfig)
	h.logger.Info("User %d completed habit: %s (ID: %d)", userID, habit.Name, habitID)
}

// HandleHabitComment обрабатывает ввод комментария при выполнении
func (h *BotHandlers) HandleHabitComment(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID
	comment := message.Text

	// Достаем ID привычки из контекста (сохранялся при нажатии кнопки)
	habitIDStr := h.contextManager.GetData(userID, "completing_habit_id")
	if habitIDStr == "" {
		h.bot.sendMessage(chatID, FormatError("ошибка: привычка не найдена"))
		h.contextManager.ClearContext(userID)
		return
	}

	habitID, err := strconv.Atoi(habitIDStr)
	if err != nil {
		h.bot.sendMessage(chatID, FormatError("ошибка обработки"))
		h.contextManager.ClearContext(userID)
		return
	}

	ctx := context.Background()

	// Получаем пользователя
	user, err := h.userService.GetOrCreateUser(ctx, userID, "", "", "", "")
	if err != nil {
		h.logger.Error("Failed to get user: %v", err)
		h.bot.sendMessage(chatID, FormatError("ошибка пользователя"))
		h.contextManager.ClearContext(userID)
		return
	}

	// Логируем завершение с комментарием
	_, err = h.logService.LogCompletion(ctx, habitID, int(user.ID), comment)
	if err != nil {
		h.logger.Error("Failed to log completion with comment: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось сохранить комментарий"))
		h.contextManager.ClearContext(userID)
		return
	}

	// Получаем привычку для отображения
	habit, _ := h.habitService.GetHabit(ctx, habitID)

	confirmText := fmt.Sprintf("✅ Спасибо!\n\n*%s* выполнена\n\n💬 Комментарий сохранён", habit.Name)
	msgConfig := tgbotapi.NewMessage(chatID, confirmText)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = MainMenuInlineKeyboard()

	h.bot.api.Send(msgConfig)

	// Очищаем контекст
	h.contextManager.ClearContext(userID)
	h.logger.Info("User %d saved comment for habit %d: %s", userID, habitID, comment)
}

// handleHabitDescription показывает описание привычки
func (h *BotHandlers) handleHabitDescription(userID int64, chatID int64, habitID int) {
	ctx := context.Background()

	// Получаем привычку
	habit, err := h.habitService.GetHabit(ctx, habitID)
	if err != nil {
		h.logger.Error("Failed to get habit: %v", err)
		h.bot.sendMessage(chatID, FormatError("привычка не найдена"))
		return
	}

	// Проверяем что пользователь имеет доступ к этой привычке
	user, err := h.userService.GetOrCreateUser(ctx, userID, "", "", "", "")
	if err != nil || user == nil || habit.UserID != int32(user.ID) {
		h.bot.sendMessage(chatID, FormatError("нет доступа к этой привычке"))
		return
	}

	// Форматируем описание
	descriptionText := fmt.Sprintf(`📄 *Описание привычки*

*%s*

%s`, habit.Name, habit.Description)

	msgConfig := tgbotapi.NewMessage(chatID, descriptionText)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад к привычке", fmt.Sprintf("habit_%d_view", habitID)),
		),
	)

	h.bot.api.Send(msgConfig)
}

// handleHabitsQuickAction обрабатывает быстрые действия со всеми привычками
func (h *BotHandlers) handleHabitsQuickAction(userID int64, chatID int64, action string) {
	ctx := context.Background()

	// Получаем или создаем пользователя (нужен внутренний ID)
	user, err := h.userService.GetOrCreateUser(ctx, userID, "", "", "", "")
	if err != nil {
		h.logger.Error("Failed to get user: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось получить данные пользователя"))
		return
	}

	// Получаем все активные привычки пользователя
	habits, err := h.habitService.GetActiveHabits(ctx, int(user.ID))
	if err != nil {
		h.logger.Error("Failed to get habits: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось получить привычки"))
		return
	}

	if len(habits) == 0 {
		h.bot.sendMessage(chatID, "📭 У вас нет привычек для быстрых действий")
		return
	}

	switch action {
	case "all_complete":
		// Отмечаем все привычки как выполненные
		for _, habit := range habits {
			_, err := h.logService.LogCompletion(ctx, int(habit.ID), int(user.ID), "")
			if err != nil {
				h.logger.Error("Failed to log completion for habit %d: %v", habit.ID, err)
			}
		}
		text := fmt.Sprintf("✅ Все %d привычек отмечены как выполненные!", len(habits))
		h.bot.sendMessage(chatID, text)

	case "all_skip":
		// Пропускаем все привычки
		text := fmt.Sprintf("⏭️ Все %d привычек пропущены", len(habits))
		h.bot.sendMessage(chatID, text)
	}

	// Возвращаемся в главное меню
	msgConfig := tgbotapi.NewMessage(chatID, "🏠 Главное меню")
	msgConfig.ReplyMarkup = MainMenuInlineKeyboard()
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	h.bot.api.Send(msgConfig)

	h.logger.Info("User %d performed quick action: %s", userID, action)
}

// HandleOptionalData обрабатывает ввод описания привычки
func (h *BotHandlers) HandleOptionalData(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID
	description := strings.TrimSpace(message.Text)

	// Сохраняем описание (может быть пусто)
	h.contextManager.SetData(userID, "habit_description", description)

	// Спрашиваем про цель
	habitName := h.contextManager.GetData(userID, "habit_name")

	h.contextManager.SetState(userID, "waiting_goal")

	text := fmt.Sprintf(
		"ℹ️ Привычка: *%s*\n\n🎯 Введите цель (опционально):\n_Например: Сделать 20 отжиманий_\n_Оставьте пусто, если не нужна_",
		habitName,
	)

	msgConfig := tgbotapi.NewMessage(chatID, text)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
	h.bot.api.Send(msgConfig)
}

// HandleGoal обрабатывает ввод цели привычки
func (h *BotHandlers) HandleGoal(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID
	goal := strings.TrimSpace(message.Text)

	// Сохраняем цель (может быть пусто)
	h.contextManager.SetData(userID, "habit_goal", goal)

	habitName := h.contextManager.GetData(userID, "habit_name")
	frequency := h.contextManager.GetData(userID, "habit_frequency")
	description := h.contextManager.GetData(userID, "habit_description")

	// Показываем подтверждение
	h.contextManager.SetState(userID, "waiting_confirmation")

	confirmText := fmt.Sprintf(
		"✨ Подтверждение:\n\n*Название:* %s\n*Частота:* %s\n*Описание:* %s\n*Цель:* %s\n\nСоздать привычку?",
		habitName,
		formatFrequencyText(frequency),
		formatOptional(description),
		formatOptional(goal),
	)

	msgConfig := tgbotapi.NewMessage(chatID, confirmText)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = ConfirmKeyboard("confirm_create_habit", "menu_main")
	sent, _ := h.bot.api.Send(msgConfig)
	// Сохраняем messageID для будущего удаления
	h.contextManager.SetMessageID(userID, chatID, sent.MessageID)
}
