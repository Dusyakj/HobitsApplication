package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger глобальный логгер приложения
var Logger *zap.Logger

// Init инициализирует глобальный логгер с поддержкой разных уровней
func Init(environment string) error {
	var config zap.Config

	if environment == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	stdoutSink := zapcore.AddSync(os.Stdout)

	// Лог файл с ротацией
	fileWriter := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    100, // MB
		MaxBackups: 10,
		MaxAge:     30, // days
		Compress:   true,
	}

	// Для development добавляем цвета и короткие пути
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Для production используем JSON энкодер
	if environment == "production" {
		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Создаем core который пишет в оба места
	core := zapcore.NewTee(
		// Stdout - для контейнеров (Docker)
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			stdoutSink,
			zapcore.DebugLevel,
		),
		// File - для архивирования и анализа
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(fileWriter),
			zapcore.DebugLevel,
		),
	)

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	defer Logger.Sync()

	return nil
}

// GetLogger возвращает логгер для конкретного пакета/модуля
func GetLogger(name string) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}
	return Logger.Named(name)
}

// Удобные функции для логирования
func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	}
}
