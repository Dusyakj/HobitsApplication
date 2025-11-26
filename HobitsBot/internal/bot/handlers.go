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

type BotHandlers struct {
	bot             *Bot
	habitService    *service.HabitService
	userService     *service.UserService
	logService      *service.LogService
	reminderService *service.ReminderService
	contextManager  *ContextManager
	logger          Logger
}

func NewBotHandlers(
	bot *Bot,
	habitService *service.HabitService,
	userService *service.UserService,
	logService *service.LogService,
	reminderService *service.ReminderService,
	contextManager *ContextManager,
	logger Logger,
) *BotHandlers {
	return &BotHandlers{
		bot:             bot,
		habitService:    habitService,
		userService:     userService,
		logService:      logService,
		reminderService: reminderService,
		contextManager:  contextManager,
		logger:          logger,
	}
}

// GetContextManager returns the context manager
func (h *BotHandlers) GetContextManager() *ContextManager {
	return h.contextManager
}

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
	msgConfig.ReplyMarkup = MainMenuKeyboard() // Это остается reply клавиатура

	h.bot.api.Send(msgConfig)
}

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
		h.bot.sendMessage(chatID, FormatHabitList([]string{}))
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

// HandleGetToday получает напоминания на сегодня
func (h *BotHandlers) HandleGetToday(message *tgbotapi.Message) {
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
		h.bot.sendMessage(chatID, "🎉 Все напоминания выполнены или их нет на сегодня!")
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

func countCompletedReminders(reminders []*service.ReminderResponse) int {
	count := 0
	for _, r := range reminders {
		if r.IsCompleted {
			count++
		}
	}
	return count
}

// HandleGetStats получает статистику
func (h *BotHandlers) HandleGetStats(message *tgbotapi.Message) {
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
		h.bot.sendMessage(chatID, "📭 У вас нет привычек для отслеживания статистики")
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
	)

	// Добавляем информацию о статусе на сегодня
	if habitReminder != nil {
		if habitReminder.IsCompleted {
			text += "\n✅ *Выполнена сегодня!*"
		} else {
			text += "\n❌ *Не выполнена сегодня*"
		}
	} else {
		text += "\n⏸ *Не запланирована на сегодня*"
	}

	msgConfig := tgbotapi.NewMessage(chatID, text)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown

	// Показываем кнопки только если привычка активна на сегодня И ещё не выполнена
	if habitReminder != nil && !habitReminder.IsCompleted {
		// Кнопки выполнить/пропустить только если не выполнена
		msgConfig.ReplyMarkup = HabitActionsKeyboard(int(habit.ID))
	} else {
		// Если привычка не на сегодня или уже выполнена, показываем кнопку деактивации и назад
		msgConfig.ReplyMarkup = HabitDetailKeyboardDeactivateOnly(int(habit.ID))
	}

	h.bot.api.Send(msgConfig)
}

// HandleCallbackQuery обрабатывает callback запросы
func (h *BotHandlers) HandleCallbackQuery(query *tgbotapi.CallbackQuery) {
	// Проверяем что query имеет необходимые поля
	if query.From == nil || query.Message == nil || query.Message.Chat == nil {
		h.logger.Error("Invalid callback query: missing required fields")
		return
	}

	userID := query.From.ID
	chatID := query.Message.Chat.ID
	data := query.Data
	messageID := query.Message.MessageID

	// Сохраняем текущий message ID для потенциального удаления
	h.contextManager.SetMessageID(userID, chatID, messageID)

	// Закрываем уведомление загрузки
	h.bot.answerCallbackQuery(query.ID, "")

	// Парсим callback данные
	parts := strings.Split(data, "_")
	if len(parts) < 2 {
		h.logger.Warn("Invalid callback data: %s", data)
		return
	}

	action := parts[0]

	switch action {
	case "habit":
		if len(parts) >= 3 {
			habitID, _ := strconv.Atoi(parts[1])
			habitAction := parts[2]

			if habitAction == "add_comment" {
				h.contextManager.SetState(userID, "waiting_habit_comment")
				h.contextManager.SetData(userID, "completing_habit_id", parts[1])
				msgConfig := tgbotapi.NewMessage(chatID, "📝 Введите комментарий к выполнению:")
				msgConfig.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
				h.bot.api.Send(msgConfig)
			} else if habitAction == "stats" {
				h.handleHabitStats(userID, chatID, habitID)
			} else if habitAction == "complete_from_reminders" {
				h.handleHabitCompleteWithComment(userID, chatID, habitID)
			} else if habitAction == "deactivate" {
				h.handleHabitDeactivate(userID, chatID, habitID)
			} else {
				h.HandleHabitAction(&tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: chatID},
					From: &tgbotapi.User{ID: userID},
				}, habitID, habitAction)
			}
		}
	case "archive_habit":
		if len(parts) >= 3 {
			habitID, _ := strconv.Atoi(parts[1])
			habitAction := parts[2]
			if habitAction == "view" {
				h.handleArchiveHabitView(userID, chatID, habitID)
			}
		}
	case "restore_habit":
		if len(parts) >= 2 {
			habitID, _ := strconv.Atoi(parts[1])
			h.handleHabitRestore(userID, chatID, habitID)
		}
	case "freq":
		h.handleFrequencySelect(userID, chatID, parts[1])
	case "weekday":
		h.handleWeekdaySelect(userID, chatID, parts[1], query.Message.MessageID)
	case "weekdays":
		h.handleWeekdaysDone(userID, chatID, query.Message.MessageID)
	case "monthday":
		h.handleMonthdaySelect(userID, chatID, parts[1], query.Message.MessageID)
	case "monthdays":
		h.handleMonthdaysDone(userID, chatID, query.Message.MessageID)
	case "time":
		h.handleTimeSelect(userID, chatID, strings.Join(parts[1:], "_"))
	case "menu":
		h.handleMenuSelect(userID, chatID, parts[1])
	case "confirm":
		if parts[1] == "create" && len(parts) >= 3 && parts[2] == "habit" {
			h.handleCreateHabitConfirm(userID, chatID)
		} else if parts[1] == "deactivate" && len(parts) >= 3 {
			habitID, _ := strconv.Atoi(parts[2])
			h.handleConfirmDeactivate(userID, chatID, habitID)
		}
	case "habits":
		// Быстрые действия со всеми привычками
		if len(parts) >= 2 {
			h.handleHabitsQuickAction(userID, chatID, parts[1])
		}
	case "stats":
		if parts[1] == "all" {
			h.HandleGetStats(&tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}})
		}
	}
}

// handleFrequencySelect обрабатывает выбор частоты привычки
func (h *BotHandlers) handleFrequencySelect(userID int64, chatID int64, frequency string) {
	h.contextManager.SetData(userID, "habit_frequency", frequency)

	habitName := h.contextManager.GetData(userID, "habit_name")

	switch frequency {
	case "daily":
		// Для ежедневной привычки - сразу спрашиваем про опциональные данные
		h.askForOptionalData(userID, chatID, habitName, frequency)

	case "weekly":
		// Для еженедельной - спрашиваем дни недели
		h.contextManager.SetData(userID, "selected_weekdays", "")
		msgConfig := tgbotapi.NewMessage(
			chatID,
			"📅 Выберите дни недели для привычки:\n\n(Нажимайте на дни, можно выбрать несколько)",
		)
		msgConfig.ReplyMarkup = WeekdaysKeyboard()
		h.bot.api.Send(msgConfig)

	case "monthly":
		// Для ежемесячной - спрашиваем дни месяца
		h.contextManager.SetData(userID, "selected_monthdays", "")
		msgConfig := tgbotapi.NewMessage(
			chatID,
			"📊 Выберите дни месяца для привычки:\n\n(Нажимайте на дни, можно выбрать несколько)",
		)
		msgConfig.ReplyMarkup = DayNumbersKeyboard()
		h.bot.api.Send(msgConfig)
	}
}

// handleWeekdaySelect обрабатывает выбор отдельного дня недели
func (h *BotHandlers) handleWeekdaySelect(userID int64, chatID int64, day string, messageID int) {
	// Получаем текущий список выбранных дней
	selectedDays := h.contextManager.GetData(userID, "selected_weekdays")

	// Проверяем, уже ли выбран этот день (для toggle функции)
	days := strings.Split(selectedDays, ",")
	dayFound := false
	newDays := []string{}

	for _, d := range days {
		if d != "" && d != day {
			newDays = append(newDays, d)
		} else if d == day {
			dayFound = true
		}
	}

	// Если день не был выбран - добавляем его
	if !dayFound {
		if selectedDays == "" {
			h.contextManager.SetData(userID, "selected_weekdays", day)
		} else {
			h.contextManager.SetData(userID, "selected_weekdays", selectedDays+","+day)
		}
	} else {
		// Если был выбран - убираем его (toggle)
		h.contextManager.SetData(userID, "selected_weekdays", strings.Join(newDays, ","))
	}

	// Отправляем обновленный список выбранных дней
	updatedDays := h.contextManager.GetData(userID, "selected_weekdays")
	dayNames := []string{"Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"}
	var selectedDayNames []string
	for _, dayStr := range strings.Split(updatedDays, ",") {
		if dayStr != "" {
			if dayIdx, err := strconv.Atoi(dayStr); err == nil && dayIdx >= 1 && dayIdx <= 7 {
				selectedDayNames = append(selectedDayNames, dayNames[dayIdx-1])
			}
		}
	}

	statusText := "📅 *Выбранные дни:* "
	if len(selectedDayNames) > 0 {
		statusText += strings.Join(selectedDayNames, ", ")
	} else {
		statusText += "_нет_"
	}

	// Редактируем существующее сообщение вместо отправки нового
	editConfig := tgbotapi.NewEditMessageText(chatID, messageID, statusText)
	editConfig.ParseMode = tgbotapi.ModeMarkdown
	keyboard := WeekdaysKeyboard()
	editConfig.ReplyMarkup = &keyboard
	if _, err := h.bot.api.Send(editConfig); err != nil {
		h.logger.Error("Failed to edit weekday selection message: %v", err)
	}
}

// handleWeekdaysDone обрабатывает завершение выбора дней недели
func (h *BotHandlers) handleWeekdaysDone(userID int64, chatID int64, messageID int) {
	selectedDays := h.contextManager.GetData(userID, "selected_weekdays")

	if selectedDays == "" {
		h.bot.sendMessage(chatID, "❌ Пожалуйста, выберите хотя бы один день недели")
		return
	}

	habitName := h.contextManager.GetData(userID, "habit_name")

	// Показываем выбранные дни пользователю
	dayNames := []string{"Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"}
	var selectedDayNames []string
	for _, dayStr := range strings.Split(selectedDays, ",") {
		if dayStr != "" {
			if dayIdx, err := strconv.Atoi(dayStr); err == nil && dayIdx >= 1 && dayIdx <= 7 {
				selectedDayNames = append(selectedDayNames, dayNames[dayIdx-1])
			}
		}
	}

	// Сохраняем выбранные дни и переходим к опциональным данным
	h.contextManager.SetData(userID, "selected_weekdays", selectedDays)
	h.askForOptionalData(userID, chatID, habitName, "weekly")

	h.logger.Info("User %d selected weekdays: %s", userID, selectedDays)
}

// handleMonthdaySelect обрабатывает выбор дня месяца
func (h *BotHandlers) handleMonthdaySelect(userID int64, chatID int64, day string, messageID int) {
	// Получаем текущий список выбранных дней
	selectedDays := h.contextManager.GetData(userID, "selected_monthdays")

	// Проверяем, уже ли выбран этот день (для toggle функции)
	days := strings.Split(selectedDays, ",")
	dayFound := false
	newDays := []string{}

	for _, d := range days {
		if d != "" && d != day {
			newDays = append(newDays, d)
		} else if d == day {
			dayFound = true
		}
	}

	// Если день не был выбран - добавляем его
	if !dayFound {
		if selectedDays == "" {
			h.contextManager.SetData(userID, "selected_monthdays", day)
		} else {
			h.contextManager.SetData(userID, "selected_monthdays", selectedDays+","+day)
		}
	} else {
		// Если был выбран - убираем его (toggle)
		h.contextManager.SetData(userID, "selected_monthdays", strings.Join(newDays, ","))
	}

	// Обновляем существующее сообщение
	updatedDays := h.contextManager.GetData(userID, "selected_monthdays")
	var selectedDayNumbers []string
	for _, dayStr := range strings.Split(updatedDays, ",") {
		if dayStr != "" {
			selectedDayNumbers = append(selectedDayNumbers, dayStr)
		}
	}

	statusText := "📊 *Выбранные дни месяца:* "
	if len(selectedDayNumbers) > 0 {
		statusText += strings.Join(selectedDayNumbers, ", ")
	} else {
		statusText += "_нет_"
	}

	// Редактируем существующее сообщение вместо отправки нового
	editConfig := tgbotapi.NewEditMessageText(chatID, messageID, statusText)
	editConfig.ParseMode = tgbotapi.ModeMarkdown
	keyboard := DayNumbersKeyboard()
	editConfig.ReplyMarkup = &keyboard
	if _, err := h.bot.api.Send(editConfig); err != nil {
		h.logger.Error("Failed to edit monthday selection message: %v", err)
	}
}

// handleMonthdaysDone обрабатывает завершение выбора дней месяца
func (h *BotHandlers) handleMonthdaysDone(userID int64, chatID int64, messageID int) {
	selectedDays := h.contextManager.GetData(userID, "selected_monthdays")

	if selectedDays == "" {
		h.bot.sendMessage(chatID, "❌ Пожалуйста, выберите хотя бы один день месяца")
		return
	}

	habitName := h.contextManager.GetData(userID, "habit_name")

	// Сохраняем выбранные дни и переходим к опциональным данным
	h.contextManager.SetData(userID, "selected_monthdays", selectedDays)
	h.askForOptionalData(userID, chatID, habitName, "monthly")

	h.logger.Info("User %d selected monthdays: %s", userID, selectedDays)
}

// askForOptionalData спрашивает описание и цель привычки
func (h *BotHandlers) askForOptionalData(userID int64, chatID int64, habitName string, frequency string) {
	h.contextManager.SetState(userID, "waiting_optional_data")

	text := fmt.Sprintf(
		"ℹ️ Привычка: *%s*\nЧастота: %s\n\n📝 Введите описание (опционально):\n_Оставьте пусто, если не нужно_",
		habitName,
		formatFrequencyText(frequency),
	)

	msgConfig := tgbotapi.NewMessage(chatID, text)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
	h.bot.api.Send(msgConfig)
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
	h.bot.api.Send(msgConfig)
}

// handleTimeSelect обрабатывает выбор времени напоминания
func (h *BotHandlers) handleTimeSelect(userID int64, chatID int64, timeKey string) {
	h.contextManager.SetData(userID, "reminder_time", timeKey)

	text := fmt.Sprintf("⏰ Время напоминания: %s\n\nПерейти к главному меню?", timeKey)
	msgConfig := tgbotapi.NewMessage(chatID, text)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = ConfirmKeyboard("menu_main", "menu_main")

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
		msgConfig := tgbotapi.NewMessage(chatID, "🏠 Главное меню")
		msgConfig.ReplyMarkup = MainMenuKeyboard()
		h.bot.api.Send(msgConfig)
	case "habits":
		h.HandleGetHabits(&tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}})
	case "archive":
		h.handleHabitArchive(userID, chatID)
	}
}

// isHabitActiveTodayWithContext проверяет активна ли привычка на сегодня
// используя данные из контекста при создании привычки
func (h *BotHandlers) isHabitActiveTodayWithContext(frequency string, weekdays string, monthdays string, date time.Time) bool {
	today := date.Weekday()
	dayOfMonth := date.Day()

	switch frequency {
	case "daily":
		// Ежедневная привычка всегда активна
		return true

	case "weekly":
		// Еженедельная привычка - проверяем день недели
		if weekdays == "" {
			return false
		}
		// weekdays формат: "1,3,5" (1=Пн, 7=Вс)
		// Go Weekday: 0=Вс, 1=Пн, ..., 6=Сб
		goDay := int(today)
		if goDay == 0 {
			goDay = 7 // Вс = 7 в нашей системе
		}

		selectedDays := strings.Split(weekdays, ",")
		for _, dayStr := range selectedDays {
			dayStr = strings.TrimSpace(dayStr)
			if dayStr != "" && dayStr == fmt.Sprintf("%d", goDay) {
				return true
			}
		}
		return false

	case "monthly":
		// Ежемесячная привычка - проверяем день месяца
		if monthdays == "" {
			return false
		}
		// monthdays формат: "1,15,30"
		selectedDays := strings.Split(monthdays, ",")
		for _, dayStr := range selectedDays {
			dayStr = strings.TrimSpace(dayStr)
			if dayStr != "" && dayStr == fmt.Sprintf("%d", dayOfMonth) {
				return true
			}
		}
		return false

	default:
		return false
	}
}

// handleCreateHabitConfirm обрабатывает подтверждение создания привычки
func (h *BotHandlers) handleCreateHabitConfirm(userID int64, chatID int64) {
	ctx := context.Background()

	// Получаем или создаём пользователя (для получения внутреннего ID)
	user, err := h.userService.GetOrCreateUser(ctx, userID, "", "", "", "")
	if err != nil {
		h.logger.Error("Failed to get/create user: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось инициализировать профиль"))
		return
	}

	// Получаем данные привычки из контекста
	habitName := h.contextManager.GetData(userID, "habit_name")
	frequency := h.contextManager.GetData(userID, "habit_frequency")
	description := h.contextManager.GetData(userID, "habit_description")
	goal := h.contextManager.GetData(userID, "habit_goal")
	selectedWeekdays := h.contextManager.GetData(userID, "selected_weekdays")
	selectedMonthdays := h.contextManager.GetData(userID, "selected_monthdays")

	if habitName == "" || frequency == "" {
		h.logger.Error("Missing habit data for user %d", userID)
		h.bot.sendMessage(chatID, FormatError("не удалось найти данные привычки"))
		return
	}

	// Преобразуем частоту в требуемый формат
	var frequencyStr string
	switch frequency {
	case "daily":
		frequencyStr = "daily"
	case "weekly":
		frequencyStr = "weekly"
	case "monthly":
		frequencyStr = "monthly"
	default:
		frequencyStr = "daily"
	}

	// Создаём привычку через сервис (используем внутренний ID пользователя)
	createReq := service.CreateHabitRequest{
		UserID:      int(user.ID),
		Name:        habitName,
		Frequency:   frequencyStr,
		Description: description,
		Goal:        goal,
	}

	h.logger.Info("Creating habit for user %d: name=%s, freq=%s", user.ID, habitName, frequencyStr)

	habit, err := h.habitService.CreateHabit(ctx, createReq)
	if err != nil {
		h.logger.Error("Failed to create habit: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось создать привычку"))
		return
	}

	h.logger.Info("Habit created with ID %d", habit.ID)

	// Если выбрана еженедельная привычка - сохраняем дни
	if frequencyStr == "weekly" && selectedWeekdays != "" {
		err = h.habitService.SetWeeklyDays(ctx, int(habit.ID), selectedWeekdays)
		if err != nil {
			h.logger.Error("Failed to set weekly days: %v", err)
			// Не возвращаем ошибку, привычка уже создана
		}
	}

	// Если выбрана ежемесячная привычка - сохраняем дни
	if frequencyStr == "monthly" && selectedMonthdays != "" {
		err = h.habitService.SetMonthlyDays(ctx, int(habit.ID), selectedMonthdays)
		if err != nil {
			h.logger.Error("Failed to set monthly days: %v", err)
			// Не возвращаем ошибку, привычка уже создана
		}
	}

	// Генерируем напоминание СРАЗУ если привычка активна на сегодня
	today := time.Now()
	isActiveToday := h.isHabitActiveTodayWithContext(frequencyStr, selectedWeekdays, selectedMonthdays, today)

	h.logger.Info("Habit %s (ID: %d) - frequency: %s, weekdays: '%s', monthdays: '%s', isActiveToday: %v",
		habit.Name, habit.ID, frequencyStr, selectedWeekdays, selectedMonthdays, isActiveToday)

	if isActiveToday {
		// Создаём напоминание на сегодня через GenerateRemindersForToday
		// (это более надёжный способ, чем CreateReminder)
		remindersCreated, reminderErr := h.reminderService.GenerateRemindersForToday(ctx, int(user.ID))
		if reminderErr != nil {
			h.logger.Error("Failed to generate reminders for new habit: %v", reminderErr)
			// Не показываем ошибку пользователю, привычка уже создана
		} else {
			h.logger.Info("Generated %d reminders for user %d after creating habit %d", len(remindersCreated), user.ID, habit.ID)
		}
	} else {
		h.logger.Info("Habit %s not active today, skipping reminder generation", habit.Name)
	}

	// Очищаем контекст
	h.contextManager.ClearContext(userID)

	// Отправляем сообщение о успешном создании
	text := FormatHabitCreated(habit.Name, frequencyStr)
	if isActiveToday {
		text += "\n\n✅ Привычка активна на сегодня - можно выполнить в /today"
	}
	msgConfig := tgbotapi.NewMessage(chatID, text)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = MainMenuKeyboard()

	h.bot.api.Send(msgConfig)
	h.logger.Info("User %d created habit: %s (ID: %d)", userID, habit.Name, habit.ID)
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
	msgConfig.ReplyMarkup = MainMenuKeyboard()
	h.bot.api.Send(msgConfig)

	h.logger.Info("User %d performed quick action: %s", userID, action)
}

// Helper функции

func formatFrequencyText(freq string) string {
	switch freq {
	case "daily":
		return "📅 Каждый день"
	case "weekly":
		return "📆 Каждую неделю"
	case "monthly":
		return "📊 Каждый месяц"
	default:
		return freq
	}
}

func formatOptional(text string) string {
	if text == "" {
		return "_не указано_"
	}
	return text
}

// handleHabitCompleteWithComment обрабатывает выполнение привычки и спрашивает комментарий
func (h *BotHandlers) handleHabitCompleteWithComment(userID int64, chatID int64, habitID int) {
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
	msgConfig.ReplyMarkup = MainMenuKeyboard()

	h.bot.api.Send(msgConfig)

	// Очищаем контекст
	h.contextManager.ClearContext(userID)
	h.logger.Info("User %d saved comment for habit %d: %s", userID, habitID, comment)
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
	msgConfig.ReplyMarkup = HabitsListKeyboard(habitIDs, habitNames)

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
