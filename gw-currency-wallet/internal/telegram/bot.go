package telegram

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5" // Официальная обертка Telegram Bot API
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

// Bot представляет Telegram бота и содержит его основные компоненты
type Bot struct {
	botAPI *tgbotapi.BotAPI // Клиент Telegram Bot API
	config Config           // Конфигурация бота
}

// Config содержит настройки для инициализации бота
type Config struct {
	Token               string        // Токен бота от @BotFather
	ExchangeServiceAddr string        // Адрес gRPC сервиса курсов валют
	UpdateTimeout       time.Duration // Таймаут получения обновлений
}

// New создает новый экземпляр Telegram бота
// Параметры:
//   - config: конфигурация бота (токен, адреса сервисов)
//
// Возвращает:
//   - *Bot: инициализированный бот
//   - error: ошибка при создании (например, невалидный токен)
func New(config Config) (*Bot, error) {
	// Инициализация клиента Telegram API
	botAPI, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания бота: %w", err)
	}

	return &Bot{
		botAPI: botAPI,
		config: config,
	}, nil
}

// Start запускает бота в работу и начинает обработку входящих сообщений
// Параметры:
//   - ctx: контекст для управления жизненным циклом бота
//
// Возвращает:
//   - error: ошибка при работе бота
func (b *Bot) Start(ctx context.Context) error {
	b.botAPI.Debug = true // Включаем режим отладки
	log.Printf("Авторизован как %s", b.botAPI.Self.UserName)

	// 1. Подключение к gRPC сервису курсов валют
	conn, err := grpc.NewClient(b.config.ExchangeServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Без TLS
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second, // Минимальное время попытки подключения
		}),
	)

	if err != nil {
		return fmt.Errorf("ошибка подключения к сервису курсов валют: %w", err)
	}
	defer conn.Close() // Гарантированное закрытие соединения

	// 2. Инициализация клиента для работы с курсами валют
	exchangeService := NewExchangeService(conn)

	// 3. Создание обработчиков сообщений с передачей зависимостей
	handler := NewHandler(b.botAPI, exchangeService)

	// 4. Настройка канала обновлений
	u := tgbotapi.NewUpdate(0) // offset=0 - получаем все обновления
	u.Timeout = int(b.config.UpdateTimeout.Seconds())

	// Получаем канал обновлений от Telegram
	updates := b.botAPI.GetUpdatesChan(u)

	// 5. Главный цикл обработки сообщений
	for {
		select {
		case <-ctx.Done():
			// Завершаем работу по сигналу контекста
			log.Println("Бот завершает работу...")
			return nil
		case update := <-updates:
			// Игнорируем не-сообщения (например, обновления чатов)
			if update.Message == nil {
				continue
			}

			// Обрабатываем только команды (сообщения, начинающиеся с '/')
			if update.Message.IsCommand() {
				handler.HandleCommand(update.Message)
			}
		}
	}
}
