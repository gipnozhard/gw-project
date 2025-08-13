package main

import (
	"context"
	_ "gw-currency-wallet/docs" // Импорт сгенерированной документации Swagger (важно оставить подчеркивание для side-effect импорта)
	"gw-currency-wallet/internal/config"
	"gw-currency-wallet/internal/services"
	"gw-currency-wallet/internal/storage/postgres"
	"gw-currency-wallet/internal/telegram"
	"gw-currency-wallet/routes"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Аннотации Swagger для генерации документации API

// @title Валютный Кошелек
// @description API для управления пользовательскими кошельками и обмена валют
// @tagsOrder Auth, Wallet, Exchange
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите "Bearer" пробел и ваш токен (например: Bearer abc123...)
// @tokenUrl /login
func main() {
	// 1. Загрузка конфигурации приложения
	// Используется файл config2.env для загрузки параметров (JWT секрет, настройки БД и т.д.)
	cfg, err := config.LoadConfig("config2.env")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err) // Критическая ошибка - выход приложения
	}

	// 2. Инициализация подключения к базе данных PostgreSQL
	// Используется строка подключения из конфигурации
	db, err := postgres.NewPostgresStorage(cfg.GetDBConnString())
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err) // Критическая ошибка - выход
	}
	defer db.Close() // Гарантированное закрытие подключения при завершении приложения

	// 3. Инициализация сервисов приложения

	// Сервис аутентификации (JWT)
	// Использует репозиторий пользователей и параметры из конфига
	authService := services.NewAuthService(db.GetUserRepository(), cfg.JWTSecret, cfg.TokenExpiration)

	// Сервис обмена валют
	// Подключается к внешнему сервису обмена и использует Redis для кэширования
	exchangeService, err := services.NewExchangeService(
		cfg.ExchangeServiceAddr, // Адрес сервиса обмена валют
		cfg.RedisAddr,           // Адрес Redis из конфига
		cfg.CacheTTL,            // Время жизни кэша
	)
	if err != nil {
		log.Fatalf("Ошибка создания сервиса обмена валют: %v", err) // Критическая ошибка
	}
	defer exchangeService.Close() // Закрытие соединений при завершении

	// Сервис работы с кошельками
	// Использует репозиторий кошельков и сервис обмена валют
	walletService := services.NewWalletService(db.GetWalletRepository(), exchangeService)

	// 4. Настройка маршрутизатора HTTP
	// Передаем все сервисы и JWT секрет для middleware аутентификации
	router := routes.SetupRouter(authService, walletService, exchangeService, cfg.JWTSecret)

	// 5. Запуск Telegram бота (если указан токен в конфиге)
	if cfg.TelegramToken != "" {
		bot, err := telegram.New(telegram.Config{
			Token:               cfg.TelegramToken,
			ExchangeServiceAddr: cfg.ExchangeServiceAddr,
			UpdateTimeout:       60 * time.Second,
		})
		if err != nil {
			log.Printf("Ошибка создания Telegram бота: %v", err) // Не критическая ошибка
		} else {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Запуск бота в отдельной горутине
			go func() {
				if err := bot.Start(ctx); err != nil {
					log.Printf("Ошибка в работе Telegram бота: %v", err)
				}
			}()
		}
	}

	// 6. Настройка graceful shutdown
	// Канал для получения сигналов завершения работы
	quit := make(chan os.Signal, 1)
	// Регистрация обработчиков сигналов SIGINT (Ctrl+C) и SIGTERM (kill)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 7. Запуск HTTP сервера в отдельной горутине
	go func() {
		log.Printf("Запуск сервера на %s", cfg.ServerAddress)
		if err := http.ListenAndServe(cfg.ServerAddress, router); err != nil {
			log.Fatalf("Ошибка запуска сервера: %v", err) // Критическая ошибка
		}
	}()

	// Ожидание сигнала завершения
	<-quit
	log.Println("Завершение работы сервера...")

	// Создание контекста с таймаутом для graceful shutdown
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Сервер остановлен")
}
