package bot

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/grpc"
)

type Bot struct {
	api    *tgbotapi.BotAPI
	grpc   *grpc.ClientConn
	logger Logger
}

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

func New(token string, grpcConn *grpc.ClientConn, logger Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot api: %w", err)
	}

	api.Debug = false

	return &Bot{
		api:    api,
		grpc:   grpcConn,
		logger: logger,
	}, nil
}

func (b *Bot) Start() error {
	b.logger.Info("Bot started, username: %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		go b.handleUpdate(update)
	}

	return nil
}

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		b.handleMessage(update.Message)
	case update.CallbackQuery != nil:
		b.handleCallbackQuery(update.CallbackQuery)
	}
}

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	userID := message.From.ID
	b.logger.Debug("Received message from user %d: %s", userID, message.Text)

	if message.IsCommand() {
		b.handleCommand(message)
	} else if message.ReplyToMessage != nil {
		b.handleReply(message)
	} else {
		b.sendMessage(message.Chat.ID, "Извините, я не понимаю эту команду. Используйте /help для справки.")
	}
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID
	command := message.Command()

	switch command {
	case "start":
		b.handleStart(userID, chatID, message.From)
	case "help":
		b.handleHelp(chatID)
	case "habits":
		b.handleHabits(userID, chatID)
	case "add":
		b.handleAddHabit(userID, chatID)
	case "today":
		b.handleToday(userID, chatID)
	case "stats":
		b.handleStats(userID, chatID)
	default:
		b.sendMessage(chatID, fmt.Sprintf("Неизвестная команда: /%s\nИспользуйте /help для справки.", command))
	}
}

func (b *Bot) handleReply(message *tgbotapi.Message) {
	// Обработка ответов на вопросы бота (например, название привычки)
	replyTo := message.ReplyToMessage
	if replyTo == nil || replyTo.From.ID != b.api.Self.ID {
		return
	}

	b.logger.Debug("Handling reply to message: %s", replyTo.Text)
	// Здесь будет обработка контекста ответа
}

func (b *Bot) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	data := query.Data

	b.logger.Debug("Received callback from user %d: %s", userID, data)

	// Закрываем уведомление загрузки
	b.answerCallbackQuery(query.ID, "")

	// Обрабатываем callback в зависимости от типа
	b.handleCallback(userID, chatID, data)
}

func (b *Bot) handleStart(userID int64, chatID int64, from *tgbotapi.User) {
	msg := fmt.Sprintf(`Добро пожаловать, %s! 🎉

Я помогу вам отслеживать ваши привычки и достигать целей.

Используйте /help для полного списка команд.`, from.FirstName)

	b.sendMessage(chatID, msg)
}

func (b *Bot) handleHelp(chatID int64) {
	helpText := `📖 Доступные команды:

/start - Начало работы
/habits - Мои привычки
/add - Добавить новую привычку
/today - Напоминания на сегодня
/stats - Статистика привычек
/help - Справка

Примеры использования:
• /add Exercise - добавить привычку "Exercise"
• /today - показать все напоминания на сегодня`

	b.sendMessage(chatID, helpText)
}

func (b *Bot) handleHabits(userID int64, chatID int64) {
	msg := "Здесь будут ваши привычки с кнопками для управления..."
	b.sendMessage(chatID, msg)
}

func (b *Bot) handleAddHabit(userID int64, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Введите название привычки:")
	b.api.Send(msg)
}

func (b *Bot) handleToday(userID int64, chatID int64) {
	msg := "Напоминания на сегодня:\n\n(Будет получено из сервиса)"
	b.sendMessage(chatID, msg)
}

func (b *Bot) handleStats(userID int64, chatID int64) {
	msg := "Ваша статистика:\n\n(Будет получено из сервиса)"
	b.sendMessage(chatID, msg)
}

func (b *Bot) handleCallback(userID int64, chatID int64, data string) {
	// Обработка различных callback'ов
	// Например: habit_1_complete, habit_1_skip, habit_1_info и т.д.
	b.logger.Debug("Processing callback: %s", data)
}

func (b *Bot) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := b.api.Send(msg)
	return err
}

func (b *Bot) SendMessage(chatID int64, text string) error {
	return b.sendMessage(chatID, text)
}

func (b *Bot) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return b.api.GetUpdatesChan(config)
}

func (b *Bot) answerCallbackQuery(queryID, text string) error {
	config := tgbotapi.NewCallback(queryID, text)
	_, err := b.api.Request(config)
	return err
}

func (b *Bot) editMessage(chatID int64, messageID int, text string, keyboard interface{}) error {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown

	if keyboard != nil {
		if replyMarkup, ok := keyboard.(tgbotapi.InlineKeyboardMarkup); ok {
			msg.ReplyMarkup = &replyMarkup
		}
	}

	_, err := b.api.Send(msg)
	return err
}

func (b *Bot) sendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = keyboard
	_, err := b.api.Send(msg)
	return err
}

func (b *Bot) DeleteMessage(chatID int64, messageID int) error {
	delete := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := b.api.Send(delete)
	return err
}

func Int64ToString(id int64) string {
	return strconv.FormatInt(id, 10)
}

func StringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
