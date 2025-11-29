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

// handleFrequencySelect обрабатывает выбор частоты привычки
func (h *BotHandlers) handleFrequencySelect(userID int64, chatID int64, frequency string) {
	// Удаляем предыдущее сообщение с кнопками
	h.deleteOldMessage(userID, chatID)

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
		sent, _ := h.bot.api.Send(msgConfig)
		// Сохраняем messageID для будущего удаления
		h.contextManager.SetMessageID(userID, chatID, sent.MessageID)

	case "monthly":
		// Для ежемесячной - спрашиваем дни месяца
		h.contextManager.SetData(userID, "selected_monthdays", "")
		msgConfig := tgbotapi.NewMessage(
			chatID,
			"📊 Выберите дни месяца для привычки:\n\n(Нажимайте на дни, можно выбрать несколько)",
		)
		msgConfig.ReplyMarkup = DayNumbersKeyboard()
		sent, _ := h.bot.api.Send(msgConfig)
		// Сохраняем messageID для будущего удаления
		h.contextManager.SetMessageID(userID, chatID, sent.MessageID)
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

	// Удаляем сообщение с выбором дней недели
	h.bot.DeleteMessage(chatID, messageID)

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

	// Удаляем сообщение с выбором дней месяца
	h.bot.DeleteMessage(chatID, messageID)

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

// handleTimeSelect обрабатывает выбор времени напоминания
func (h *BotHandlers) handleTimeSelect(userID int64, chatID int64, timeKey string) {
	h.contextManager.SetData(userID, "reminder_time", timeKey)

	text := fmt.Sprintf("⏰ Время напоминания: %s\n\nПерейти к главному меню?", timeKey)
	msgConfig := tgbotapi.NewMessage(chatID, text)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = ConfirmKeyboard("menu_main", "menu_main")

	h.bot.api.Send(msgConfig)
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
	// Удаляем предыдущее сообщение с подтверждением
	h.deleteOldMessage(userID, chatID)

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
		// Проверяем что дни недели выбраны
		if selectedWeekdays == "" {
			h.logger.Error("Weekly habit but no weekdays selected for user %d", userID)
			h.bot.sendMessage(chatID, FormatError("не выбраны дни недели для еженедельной привычки"))
			return
		}
	case "monthly":
		frequencyStr = "monthly"
		// Проверяем что дни месяца выбраны
		if selectedMonthdays == "" {
			h.logger.Error("Monthly habit but no monthdays selected for user %d", userID)
			h.bot.sendMessage(chatID, FormatError("не выбраны дни месяца для ежемесячной привычки"))
			return
		}
	default:
		frequencyStr = "daily"
	}

	// Создаём привычку через сервис (используем внутренний ID пользователя)
	// Передаем дни недели/месяца напрямую при создании
	createReq := service.CreateHabitRequest{
		UserID:      int(user.ID),
		Name:        habitName,
		Frequency:   frequencyStr,
		Description: description,
		Goal:        goal,
		WeeklyDays:  selectedWeekdays,
		MonthlyDays: selectedMonthdays,
	}

	h.logger.Info("Creating habit for user %d: name=%s, freq=%s, weekdays=%s, monthdays=%s",
		user.ID, habitName, frequencyStr, selectedWeekdays, selectedMonthdays)

	habit, err := h.habitService.CreateHabit(ctx, createReq)
	if err != nil {
		h.logger.Error("Failed to create habit: %v", err)
		h.bot.sendMessage(chatID, FormatError("не удалось создать привычку"))
		return
	}

	h.logger.Info("Habit created with ID %d", habit.ID)

	// Создаем первое напоминание для привычки (сервис сам определит нужна ли она на сегодня)
	reminder, reminderErr := h.reminderService.CreateInitialReminder(ctx, int(habit.ID))
	if reminderErr != nil {
		h.logger.Error("Failed to create initial reminder for habit %d: %v", habit.ID, reminderErr)
		// Не показываем критичную ошибку пользователю, привычка уже создана
	} else {
		h.logger.Info("Created initial reminder %d for habit %d", reminder.ID, habit.ID)
	}

	// Очищаем контекст
	h.contextManager.ClearContext(userID)

	// Отправляем сообщение о успешном создании
	text := FormatHabitCreated(habit.Name, frequencyStr)
	text += "\n\n💡 Напоминание создано! Проверьте раздел 🔔 Напоминания"

	msgConfig := tgbotapi.NewMessage(chatID, text)
	msgConfig.ParseMode = tgbotapi.ModeMarkdown
	msgConfig.ReplyMarkup = MainMenuInlineKeyboard()

	h.bot.api.Send(msgConfig)
	h.logger.Info("User %d created habit: %s (ID: %d)", userID, habit.Name, habit.ID)
}
