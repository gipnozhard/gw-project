package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"               // Для загрузки переменных окружения
	"gw-exchanger/internal/server"           // Пакет с логикой сервера
	"gw-exchanger/internal/storage/postgres" // Работа с PostgreSQL
	"gw-exchanger/internal/utils"            // Вспомогательные утилиты
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	// 1. Загрузка конфигурации из файла .env
	if err := godotenv.Load("config.env"); err != nil {
		log.Fatal("Ошибка загрузки файла config2.env") // Критическая ошибка - завершаем программу
	}

	// 2. Получение параметров для обновления курсов валют
	apiURL := os.Getenv("CB_API_URL") // URL API Центробанка
	updateInterval := time.Minute * time.Duration(
		getEnvAsInt("UPDATE_INTERVAL_MINUTES", 60)) // Интервал обновления (по умолчанию 60 минут)

	// 3. Формирование строки подключения к PostgreSQL
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),     // Хост базы данных
		os.Getenv("DB_PORT"),     // Порт (обычно 5432)
		os.Getenv("DB_USER"),     // Имя пользователя
		os.Getenv("DB_PASSWORD"), // Пароль
		os.Getenv("DB_NAME"),     // Имя базы данных
	)

	// 4. Проверка подключения к базе данных
	if err := checkDBConnection(connStr); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err) // Критическая ошибка
	}

	// 5. Инициализация хранилища данных с поддержкой периодического обновления
	storage, err := postgres.NewPostgresStorage(connStr, apiURL, updateInterval)
	if err != nil {
		log.Fatalf("Ошибка инициализации хранилища: %v", err) // Критическая ошибка
	}
	defer storage.Close() // Гарантированное закрытие подключения при завершении

	// 6. Первоначальное обновление курсов валют
	if err := storage.UpdateRatesFromCB(); err != nil {
		log.Printf("Ошибка первоначального обновления курсов: %v", err) // Не критическая ошибка
	}

	// 7. Вывод списка доступных валют
	utils.PrintAvailableCurrencies(storage)

	// 8. Запуск gRPC сервера
	log.Println("Запуск gRPC сервера...")
	server.Start("50051", storage) // Порт 50051 и инициализированное хранилище
}

// checkDBConnection проверяет подключение к базе данных
// Параметры:
//   - connStr: строка подключения к PostgreSQL
//
// Возвращает:
//   - error: ошибка подключения или nil при успехе
func checkDBConnection(connStr string) error {
	db, err := sql.Open("postgres", connStr) // Инициализация подключения
	if err != nil {
		return fmt.Errorf("ошибка открытия подключения: %w", err)
	}
	defer db.Close() // Гарантированное закрытие

	// Проверка подключения с таймаутом 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ошибка проверки подключения: %w", err)
	}
	return nil
}

// getEnvAsInt получает переменную окружения как целое число
// Параметры:
//   - name: имя переменной окружения
//   - defaultValue: значение по умолчанию
//
// Возвращает:
//   - int: значение переменной или значение по умолчанию
func getEnvAsInt(name string, defaultValue int) int {
	val := os.Getenv(name)
	if val == "" {
		return defaultValue
	}
	result, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return result
}
