package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"HobitsBot/internal/bot"
	"HobitsBot/internal/config"
	"HobitsBot/internal/grpc"
	"HobitsBot/internal/logger"
	"HobitsBot/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel)

	if cfg.TelegramBotToken == "" {
		log.Error("TELEGRAM_BOT_TOKEN not set")
		os.Exit(1)
	}

	log.Info("HobitsBot starting...")

	grpcClient, err := grpc.New(cfg.GRPCServerAddr)
	if err != nil {
		log.Error("Failed to connect to gRPC server: %v", err)
		os.Exit(1)
	}
	defer grpcClient.Close()

	log.Info("Connected to gRPC server at %s", cfg.GRPCServerAddr)

	botInstance, err := bot.New(cfg.TelegramBotToken, grpcClient.GetConn(), log)
	if err != nil {
		log.Error("Failed to create bot: %v", err)
		os.Exit(1)
	}

	// Создаем сервисы
	habitService := service.NewHabitService(grpcClient.GetConn(), log)
	userService := service.NewUserService(grpcClient.GetConn(), log)
	logService := service.NewLogService(grpcClient.GetConn(), log)
	reminderService := service.NewReminderService(grpcClient.GetConn(), log)

	// Создаем менеджер контекстов (TTL - 30 минут)
	contextManager := bot.NewContextManager(30 * time.Minute)

	// Создаем обработчик команд
	handlers := bot.NewBotHandlers(
		botInstance,
		habitService,
		userService,
		logService,
		reminderService,
		contextManager,
		log,
	)

	// Запускаем обработку обновлений
	go runBotUpdateLoop(botInstance, handlers, log)

	// Запускаем планировщик для 9:00 AM уведомлений
	scheduler := bot.NewScheduler(botInstance, reminderService, userService, log)
	scheduler.Start()

	log.Info("Bot is running. Press Ctrl+C to stop.")

	// Ожидаем сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down...")
	scheduler.Stop()
	os.Exit(0)
}

func runBotUpdateLoop(
	botInstance *bot.Bot,
	handlers *bot.BotHandlers,
	log *logger.Logger,
) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := botInstance.GetUpdatesChan(u)

	for update := range updates {
		go handleUpdate(update, botInstance, handlers, log)
	}
}

func handleUpdate(
	update tgbotapi.Update,
	botInstance *bot.Bot,
	handlers *bot.BotHandlers,
	log *logger.Logger,
) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered from panic: %v", r)
		}
	}()

	switch {
	case update.Message != nil:
		handleMessage(update.Message, botInstance, handlers, log)
	case update.CallbackQuery != nil:
		handlers.HandleCallbackQuery(update.CallbackQuery)
	}
}

func handleMessage(
	message *tgbotapi.Message,
	botInstance *bot.Bot,
	handlers *bot.BotHandlers,
	log *logger.Logger,
) {
	userID := message.From.ID
	log.Debug("Received message from user %d: %s", userID, message.Text)

	if message.IsCommand() {
		handleCommand(message, botInstance, handlers, log)
	} else if message.ReplyToMessage != nil && message.ReplyToMessage.From.IsBot {
		// Проверяем состояние пользователя
		state := handlers.GetContextManager().GetState(userID)
		log.Debug("User %d state: %s", userID, state)

		switch state {
		case "waiting_habit_name":
			handlers.HandleHabitName(message)
		case "waiting_optional_data":
			handlers.HandleOptionalData(message)
		case "waiting_goal":
			handlers.HandleGoal(message)
		case "waiting_habit_comment":
			handlers.HandleHabitComment(message)
		default:
			botInstance.SendMessage(message.Chat.ID, "Не понимаю. Используйте /help для справки.")
		}
	} else {
		// Обрабатываем нажатия кнопок главного меню (ЗАКОММЕНТИРОВАНО - используются inline кнопки)
		// handleMenuButton(message, botInstance, handlers, log)
		botInstance.SendMessage(message.Chat.ID, "Используйте inline кнопки под сообщениями для управления ботом.")
	}
}

// handleMenuButton обрабатывает нажатия reply кнопок (ЗАКОММЕНТИРОВАНО - используются inline кнопки)
// func handleMenuButton(
// 	message *tgbotapi.Message,
// 	botInstance *bot.Bot,
// 	handlers *bot.BotHandlers,
// 	log *logger.Logger,
// ) {
// 	userID := message.From.ID
// 	chatID := message.Chat.ID
// 	text := message.Text
//
// 	switch text {
// 	case "📋 Мои привычки":
// 		log.Debug("User %d clicked 'Мои привычки' button", userID)
// 		handlers.HandleGetHabits(message)
// 	case "➕ Добавить":
// 		log.Debug("User %d clicked 'Добавить' button", userID)
// 		handlers.HandleAddHabit(message)
// 	case "🔔 Напоминания":
// 		log.Debug("User %d clicked 'Напоминания' button", userID)
// 		handlers.HandleGetToday(message)
// 	case "📊 Статистика":
// 		log.Debug("User %d clicked 'Статистика' button", userID)
// 		handlers.HandleGetStats(message)
// 	case "⚙️ Настройки":
// 		log.Debug("User %d clicked 'Настройки' button", userID)
// 		botInstance.SendMessage(chatID, "⚙️ Настройки находятся в разработке")
// 	case "❓ Помощь":
// 		log.Debug("User %d clicked 'Помощь' button", userID)
// 		showHelp(botInstance, chatID)
// 	default:
// 		log.Debug("User %d sent text: %s", userID, text)
// 		botInstance.SendMessage(chatID, "Не понимаю эту команду. Используйте кнопки меню или /help для справки.")
// 	}
// }

func handleCommand(
	message *tgbotapi.Message,
	botInstance *bot.Bot,
	handlers *bot.BotHandlers,
	log *logger.Logger,
) {
	command := message.Command()
	userID := message.From.ID
	log.Debug("Received command from user %d: /%s", userID, command)

	switch command {
	case "start":
		handlers.HandleStart(message)
	case "help":
		showHelp(botInstance, message.Chat.ID)
	case "habits":
		handlers.HandleGetHabits(message)
	case "add":
		handlers.HandleAddHabit(message)
	case "today":
		handlers.HandleGetToday(message)
	case "stats":
		handlers.HandleGetStats(message)
	default:
		botInstance.SendMessage(
			message.Chat.ID,
			fmt.Sprintf("Неизвестная команда: /%s\nИспользуйте /help для справки.", command),
		)
	}
}

func showHelp(botInstance *bot.Bot, chatID int64) {
	helpText := `📖 *Доступные команды:*

/start - Начало работы
/habits - Мои привычки
/add - Добавить новую привычку
/today - Напоминания на сегодня
/stats - Статистика привычек
/help - Справка

*Примеры использования:*
• Нажмите /add и следуйте инструкциям
• Используйте кнопки под сообщениями для управления
• /today показывает все напоминания на сегодня`

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	botInstance.SendMessage(chatID, helpText)
}
