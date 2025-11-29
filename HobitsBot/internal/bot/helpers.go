package bot

import (
	"HobitsBot/internal/service"
)

// Helper функции

// deleteOldMessage удаляет предыдущее сообщение с кнопками, если оно есть
func (h *BotHandlers) deleteOldMessage(userID int64, chatID int64) {
	oldChatID, oldMessageID := h.contextManager.GetMessageID(userID)
	if oldMessageID > 0 && oldChatID == chatID {
		go func() {
			err := h.bot.DeleteMessage(oldChatID, oldMessageID)
			if err != nil {
				h.logger.Debug("Could not delete old message %d: %v", oldMessageID, err)
			}
		}()
		h.contextManager.ClearMessageID(userID)
	}
}

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

func countCompletedReminders(reminders []*service.ReminderResponse) int {
	count := 0
	for _, r := range reminders {
		if r.IsCompleted {
			count++
		}
	}
	return count
}
