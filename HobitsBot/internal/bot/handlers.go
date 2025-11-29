package bot

import (
	"strconv"
	"strings"

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
			} else if habitAction == "description" {
				h.handleHabitDescription(userID, chatID, habitID)
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
