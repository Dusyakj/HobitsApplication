package bot

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MainMenuKeyboard возвращает главное меню
func MainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📋 Мои привычки"),
			tgbotapi.NewKeyboardButton("➕ Добавить"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔔 Напоминания"),
			tgbotapi.NewKeyboardButton("📊 Статистика"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⚙️ Настройки"),
			tgbotapi.NewKeyboardButton("❓ Помощь"),
		),
	)
}

// HabitsListKeyboard возвращает клавиатуру со списком привычек
func HabitsListKeyboard(habitIDs []int, habitNames []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for i, habitID := range habitIDs {
		button := tgbotapi.NewInlineKeyboardButtonData(
			habitNames[i],
			fmt.Sprintf("habit_%d_view", habitID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	// Добавляем кнопку "Назад"
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "menu_main"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// HabitActionsKeyboard возвращает клавиатуру для действий с привычкой
func HabitActionsKeyboard(habitID int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Выполнено", fmt.Sprintf("habit_%d_complete", habitID)),
			tgbotapi.NewInlineKeyboardButtonData("⏭️ Пропустить", fmt.Sprintf("habit_%d_skip", habitID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Комментарий", fmt.Sprintf("habit_%d_comment", habitID)),
			tgbotapi.NewInlineKeyboardButtonData("📊 Статистика", fmt.Sprintf("habit_%d_stats", habitID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚫 Деактивировать", fmt.Sprintf("habit_%d_deactivate", habitID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "menu_habits"),
		),
	)
}

// HabitDetailKeyboardDeactivateOnly возвращает клавиатуру с кнопкой деактивации (для выполненных или неактивных привычек)
func HabitDetailKeyboardDeactivateOnly(habitID int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚫 Деактивировать", fmt.Sprintf("habit_%d_deactivate", habitID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "menu_habits"),
		),
	)
}

// FrequencyKeyboard возвращает клавиатуру для выбора частоты привычки
func FrequencyKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 Ежедневно", "freq_daily"),
			tgbotapi.NewInlineKeyboardButtonData("📆 Еженедельно", "freq_weekly"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Ежемесячно", "freq_monthly"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "menu_main"),
		),
	)
}

// ConfirmKeyboard возвращает клавиатуру подтверждения
func ConfirmKeyboard(confirmCallback, cancelCallback string) tgbotapi.InlineKeyboardMarkup {
	// Если confirmCallback = "confirm_create_habit", преобразуем в правильный формат
	if confirmCallback == "confirm_create_habit" {
		confirmCallback = "confirm_create_habit"
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да", confirmCallback),
			tgbotapi.NewInlineKeyboardButtonData("❌ Нет", cancelCallback),
		),
	)
}

// WeekdaysKeyboard возвращает клавиатуру для выбора дней недели
func WeekdaysKeyboard() tgbotapi.InlineKeyboardMarkup {
	days := []string{"Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"}
	var rows [][]tgbotapi.InlineKeyboardButton

	// Первый ряд (Пн-Сб)
	var row1 []tgbotapi.InlineKeyboardButton
	for i := 0; i < 6; i++ {
		row1 = append(row1, tgbotapi.NewInlineKeyboardButtonData(
			days[i],
			fmt.Sprintf("weekday_%d", i+1),
		))
	}
	rows = append(rows, row1)

	// Второй ряд (Вс и кнопки)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Вс", "weekday_7"),
	))

	// Кнопки действий
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("✅ Готово", "weekdays_done"),
		tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "menu_main"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// DayNumbersKeyboard возвращает клавиатуру для выбора дней месяца
func DayNumbersKeyboard() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Создаём кнопки для дней 1-31
	for day := 1; day <= 31; day += 7 {
		var row []tgbotapi.InlineKeyboardButton
		for i := 0; i < 7 && day+i <= 31; i++ {
			dayNum := day + i
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				strconv.Itoa(dayNum),
				fmt.Sprintf("monthday_%d", dayNum),
			))
		}
		rows = append(rows, row)
	}

	// Кнопки действий
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("✅ Готово", "monthdays_done"),
		tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "menu_main"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// ReminderTimeKeyboard возвращает клавиатуру для выбора времени напоминания
func ReminderTimeKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌅 08:00", "time_0800"),
			tgbotapi.NewInlineKeyboardButtonData("☀️ 12:00", "time_1200"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌆 14:00", "time_1400"),
			tgbotapi.NewInlineKeyboardButtonData("🌙 20:00", "time_2000"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏰ Свое время", "time_custom"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "menu_main"),
		),
	)
}

// QuickActionsKeyboard возвращает клавиатуру быстрых действий
func QuickActionsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Все выполнено", "habits_all_complete"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏭️ Все пропустить", "habits_all_skip"),
		),
	)
}

// SettingsKeyboard возвращает клавиатуру настроек
func SettingsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔔 Напоминания", "settings_reminders"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌐 Язык", "settings_language"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "menu_main"),
		),
	)
}

// StatsPeriodKeyboard возвращает клавиатуру для выбора периода статистики
func StatsPeriodKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 Неделя", "stats_week"),
			tgbotapi.NewInlineKeyboardButtonData("📊 Месяц", "stats_month"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 Все время", "stats_all"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "menu_main"),
		),
	)
}
