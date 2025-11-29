package bot

import (
	"fmt"
	"time"
)

// FormatWelcomeMessage форматирует приветственное сообщение
func FormatWelcomeMessage(firstName string) string {
	return fmt.Sprintf(`👋 Добро пожаловать, %s!

Я ваш личный помощник для отслеживания привычек. 🎯

📋 Мои привычки - просмотрите все ваши привычки
➕ Добавить - создайте новую привычку
🔔 Напоминания - сегодняшние задачи
📊 Статистика - анализ вашего прогресса
⚙️ Настройки - параметры приложения
❓ Помощь - справка по использованию

Выберите команду из меню ниже 👇`, firstName)
}

// FormatHabitList форматирует список привычек
func FormatHabitList(habits []string) string {
	if len(habits) == 0 {
		return `📭 У вас нет привычек.

Создайте первую привычку, нажав кнопку ➕ Добавить`
	}

	text := "📋 *Ваши привычки:*\n\n"
	for i, habit := range habits {
		text += fmt.Sprintf("%d. %s\n", i+1, habit)
	}
	return text
}

// FormatHabitDetail форматирует подробную информацию о привычке
func FormatHabitDetail(name string, frequency string, streak, bestStreak, completed, total int, goal string) string {
	frequencyStr := map[string]string{
		"daily":   "📅 Ежедневно",
		"weekly":  "📆 Еженедельно",
		"monthly": "📊 Ежемесячно",
	}[frequency]

	streakEmoji := "🔥"
	if streak == 0 {
		streakEmoji = "❄️"
	}

	result := fmt.Sprintf(`*%s*

Частота: %s
%s Текущая серия: %d дней
🏆 Лучшая серия: %d дней
✅ Выполнено: %d/%d раз`, name, frequencyStr, streakEmoji, streak, bestStreak, completed, total)

	// Добавляем цель если она есть
	if goal != "" {
		result += fmt.Sprintf("\n\n🎯 *Цель:* %s", goal)
	}

	result += "\n\nОтличная работа! Продолжайте в том же духе!"

	return result
}

// FormatTodayReminders форматирует напоминания на сегодня
func FormatTodayReminders(reminders []string, completed int, total int) string {
	if len(reminders) == 0 {
		return "🎉 Все напоминания выполнены или их нет на сегодня!"
	}

	text := fmt.Sprintf("🔔 *Напоминания на сегодня* (%d/%d)\n\n", completed, total)
	for i, reminder := range reminders {
		text += fmt.Sprintf("%d. %s\n", i+1, reminder)
	}
	return text
}

// FormatStats форматирует статистику
func FormatStats(habitName string, completionRate float64, bestStreak, currentStreak int, period string) string {
	periodText := map[string]string{
		"week":  "неделю",
		"month": "месяц",
		"all":   "все время",
	}[period]

	progressBar := formatProgressBar(int(completionRate))

	return fmt.Sprintf(`*Статистика: %s* (%s)

Процент выполнения: %s %d%%
🔥 Текущая серия: %d дней
🏆 Лучшая серия: %d дней

Продолжайте работать!`, habitName, periodText, progressBar, int(completionRate), currentStreak, bestStreak)
}

// FormatHabitCreated форматирует сообщение о создании привычки
func FormatHabitCreated(habitName, frequency string) string {
	frequencyStr := map[string]string{
		"daily":   "📅 Ежедневно",
		"weekly":  "📆 Еженедельно",
		"monthly": "📊 Ежемесячно",
	}[frequency]

	return fmt.Sprintf(`✅ Привычка создана!

*%s*
%s

Теперь я буду отправлять вам напоминания. Удачи! 🚀`, habitName, frequencyStr)
}

// FormatHabitCompleted форматирует сообщение о выполнении привычки
func FormatHabitCompleted(habitName string, newStreak int32) string {
	emoji := "🔥"
	if newStreak > 1 && newStreak%7 == 0 {
		emoji = "🎉"
	}
	if newStreak > 1 && newStreak%30 == 0 {
		emoji = "🏆"
	}

	return fmt.Sprintf(`✅ Отлично!

*%s* выполнена!

%s Серия: %d дней подряд

Продолжайте так же! 💪`, habitName, emoji, newStreak)
}

// FormatHabitSkipped форматирует сообщение о пропуске привычки
func FormatHabitSkipped(habitName string) string {
	return fmt.Sprintf(`⏭️ Привычка пропущена

*%s*

Не волнуйтесь, это случается. Главное - вернуться к выполнению привычки завтра. 💪`, habitName)
}

// FormatError форматирует сообщение об ошибке
func FormatError(errorMsg string) string {
	return fmt.Sprintf(`❌ Ошибка

%s

Пожалуйста, попробуйте позже или используйте /help.`, errorMsg)
}

// FormatLoading форматирует сообщение загрузки
func FormatLoading() string {
	return "⏳ Загрузка..."
}

// FormatProgressBar форматирует прогресс-бар
func formatProgressBar(percent int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := (percent * 10) / 100
	empty := 10 - filled

	bar := "█"
	for i := 1; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}

	return bar
}

// FormatDayOfWeek форматирует день недели
func FormatDayOfWeek(day time.Weekday) string {
	days := map[time.Weekday]string{
		time.Monday:    "Понедельник",
		time.Tuesday:   "Вторник",
		time.Wednesday: "Среда",
		time.Thursday:  "Четверг",
		time.Friday:    "Пятница",
		time.Saturday:  "Суббота",
		time.Sunday:    "Воскресенье",
	}
	return days[day]
}

// FormatTime форматирует время в удобный формат
func FormatTime(t time.Time) string {
	return t.Format("15:04")
}

// FormatDate форматирует дату в удобный формат
func FormatDate(t time.Time) string {
	return t.Format("02.01.2006")
}

// FormatStreakMilestone форматирует достижение серии
func FormatStreakMilestone(habitName string, streak int) string {
	emoji := ""
	message := ""

	switch {
	case streak == 7:
		emoji = "🌟"
		message = "неделю"
	case streak == 14:
		emoji = "✨"
		message = "две недели"
	case streak == 30:
		emoji = "🏆"
		message = "месяц"
	case streak == 100:
		emoji = "👑"
		message = "100 дней"
	default:
		return ""
	}

	return fmt.Sprintf(`%s *Поздравления!*

Вы выполняли привычку *%s* %s подряд!

%s`, emoji, habitName, message, emoji)
}
