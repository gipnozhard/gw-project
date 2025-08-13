package postgres

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Драйвер PostgreSQL (импорт для side effects)
	"log"
	"os"
	"path/filepath"
	"time"
)

// PostgresStorage представляет хранилище данных в PostgreSQL
type PostgresStorage struct {
	db             *sql.DB       // Подключение к базе данных
	apiURL         string        // URL API Центробанка для получения курсов
	updateInterval time.Duration // Интервал обновления курсов
}

// NewPostgresStorage создает и инициализирует новое подключение к PostgreSQL
// Параметры:
//   - connStr: строка подключения к основной БД
//   - apiURL: URL API Центробанка
//   - updateInterval: интервал обновления курсов
//
// Возвращает:
//   - *PostgresStorage: инициализированное хранилище
//   - error: ошибка при создании
func NewPostgresStorage(connStr string, apiURL string, updateInterval time.Duration) (*PostgresStorage, error) {
	// 1. Подключение к служебной БД postgres для проверки/создания нужной БД
	adminConnStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		os.Getenv("DB_HOST"),     // Хост БД
		os.Getenv("DB_PORT"),     // Порт БД
		os.Getenv("DB_USER"),     // Имя пользователя
		os.Getenv("DB_PASSWORD"), // Пароль
	)

	adminDb, err := sql.Open("postgres", adminConnStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к служебной БД: %v", err)
	}
	defer adminDb.Close()

	// 2. Проверка существования БД и создание при необходимости
	var exists bool
	err = adminDb.QueryRowContext(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'exchange_rates')").Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки существования БД: %v", err)
	}

	if !exists {
		_, err = adminDb.ExecContext(context.Background(), "CREATE DATABASE exchange_rates")
		if err != nil {
			return nil, fmt.Errorf("ошибка создания БД: %v", err)
		}
	}

	// 2. Подключение к основной БД
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к основной БД: %v", err)
	}

	// 3. Проверка подключения с таймаутом 3 секунды
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ошибка проверки подключения: %v", err)
	}

	// 4. Применение миграций
	if err := applyMigrations(db); err != nil {
		return nil, fmt.Errorf("ошибка применения миграций: %v", err)
	}

	log.Println("Успешное подключение к PostgreSQL")

	storage := &PostgresStorage{
		db:             db,
		apiURL:         apiURL,
		updateInterval: updateInterval,
	}

	// 5. Запуск фонового обновления курсов
	go storage.startRateUpdater()

	return storage, nil
}

// applyMigrations применяет SQL-миграции из файла
func applyMigrations(db *sql.DB) error {
	// Получаем путь к файлу миграции
	migrationPath := filepath.Join("migrations", "001_init.sql")

	log.Printf("Путь к миграции: %s", migrationPath)
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		return fmt.Errorf("файл миграции не найден: %v", err)
	}

	// Чтение файла миграции
	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла миграции: %v", err)
	}

	// Выполнение SQL-запросов
	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("ошибка выполнения миграции: %v", err)
	}

	return nil
}

// Close закрывает подключение к БД
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}
