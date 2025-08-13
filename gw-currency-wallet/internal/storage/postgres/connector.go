package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq" // Импорт драйвера PostgreSQL без прямого использования
	"gw-currency-wallet/internal/models"
	"gw-currency-wallet/internal/storage"
	"log"
	"os"
	"time"
)

// userRepository реализует интерфейс UserRepository для работы с пользователями в PostgreSQL
type userRepository struct {
	db *sql.DB // Подключение к базе данных
}

// PostgresStorage объединяет все репозитории для работы с PostgreSQL
type PostgresStorage struct {
	db *sql.DB // Общее подключение к БД
}

// walletRepository реализует интерфейс WalletRepository для работы с кошельками
type walletRepository struct {
	db *sql.DB // Подключение к базе данных
}

// CreateUser создает нового пользователя в базе данных
func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	// SQL-запрос с возвратом ID созданного пользователя
	query := `INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, user.Username, user.Email, user.PasswordHash).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("ошибка создания пользователя: %w", err)
	}
	return nil
}

// GetUserByUsername находит пользователя по имени пользователя
func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE username = $1`
	return r.queryUser(ctx, query, username)
}

// GetUserByEmail находит пользователя по email
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE email = $1`
	return r.queryUser(ctx, query, email)
}

// GetUserByID находит пользователя по ID
func (r *userRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE id = $1`
	return r.queryUser(ctx, query, id)
}

// queryUser общий метод для выполнения запросов пользователей
func (r *userRepository) queryUser(ctx context.Context, query string, args ...interface{}) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Пользователь не найден - не ошибка
		}
		return nil, fmt.Errorf("ошибка запроса пользователя: %w", err)
	}
	return &user, nil
}

// GetBalance возвращает баланс пользователя
func (r *walletRepository) GetBalance(ctx context.Context, userID int) (*models.Balance, error) {
	query := `SELECT usd, rub, eur FROM wallets WHERE user_id = $1`
	row := r.db.QueryRowContext(ctx, query, userID)

	var balance models.Balance
	err := row.Scan(&balance.USD, &balance.RUB, &balance.EUR)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := r.CreateWallet(ctx, userID); err != nil {
				return nil, err
			}
			return &models.Balance{USD: 0, RUB: 0, EUR: 0}, nil
		}
		return nil, fmt.Errorf("ошибка получения баланса: %w", err)
	}
	return &balance, nil
}

// CreateWallet создает новый кошелек для пользователя
func (r *walletRepository) CreateWallet(ctx context.Context, userID int) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO wallets (user_id) VALUES ($1)", userID)
	if err != nil {
		return fmt.Errorf("ошибка создания кошелька: %w", err)
	}
	return nil
}

// UpdateBalance обновляет баланс пользователя для указанной валюты
func (r *walletRepository) UpdateBalance(ctx context.Context, userID int, currency string, amount float64) (*models.Balance, error) {
	var query string
	switch currency {
	case "USD":
		query = `UPDATE wallets SET usd = usd + $1 WHERE user_id = $2 RETURNING usd, rub, eur`
	case "RUB":
		query = `UPDATE wallets SET rub = rub + $1 WHERE user_id = $2 RETURNING usd, rub, eur`
	case "EUR":
		query = `UPDATE wallets SET eur = eur + $1 WHERE user_id = $2 RETURNING usd, rub, eur`
	default:
		return nil, fmt.Errorf("неподдерживаемая валюта: %s", currency)
	}

	row := r.db.QueryRowContext(ctx, query, amount, userID)
	var balance models.Balance
	err := row.Scan(&balance.USD, &balance.RUB, &balance.EUR)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления баланса: %w", err)
	}
	return &balance, nil
}

// Transfer выполняет перевод средств между пользователями в рамках транзакции
func (r *walletRepository) Transfer(
	ctx context.Context,
	fromUserID int,
	toUserID int,
	currency string,
	amount float64,
) (*models.Balance, *models.Balance, error) {
	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback() // Откат при ошибке

	// Снимаем средства у отправителя
	fromBalance, err := r.updateBalanceTx(ctx, tx, fromUserID, currency, -amount)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка списания у отправителя: %w", err)
	}

	// Зачисляем средства получателю
	toBalance, err := r.updateBalanceTx(ctx, tx, toUserID, currency, amount)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка зачисления получателю: %w", err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("ошибка подтверждения транзакции: %w", err)
	}

	return fromBalance, toBalance, nil
}

// Exchange выполняет обмен валюты в рамках транзакции
func (r *walletRepository) Exchange(
	ctx context.Context,
	userID int,
	fromCurrency string,
	toCurrency string,
	amount float64,
	rate float64,
) (*models.Balance, error) {
	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback()

	// Снимаем средства в исходной валюте
	_, err = r.updateBalanceTx(ctx, tx, userID, fromCurrency, -amount)
	if err != nil {
		return nil, fmt.Errorf("ошибка списания %s: %w", fromCurrency, err)
	}

	// Зачисляем средства в целевой валюте
	exchangedAmount := amount * rate
	balance, err := r.updateBalanceTx(ctx, tx, userID, toCurrency, exchangedAmount)
	if err != nil {
		return nil, fmt.Errorf("ошибка зачисления %s: %w", toCurrency, err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("ошибка подтверждения транзакции: %w", err)
	}

	return balance, nil
}

// updateBalanceTx вспомогательный метод для обновления баланса в транзакции
func (r *walletRepository) updateBalanceTx(
	ctx context.Context,
	tx *sql.Tx,
	userID int,
	currency string,
	amount float64,
) (*models.Balance, error) {
	var query string
	switch currency {
	case "USD":
		query = `UPDATE wallets SET usd = usd + $1 WHERE user_id = $2 RETURNING usd, rub, eur`
	case "RUB":
		query = `UPDATE wallets SET rub = rub + $1 WHERE user_id = $2 RETURNING usd, rub, eur`
	case "EUR":
		query = `UPDATE wallets SET eur = eur + $1 WHERE user_id = $2 RETURNING usd, rub, eur`
	default:
		return nil, fmt.Errorf("неподдерживаемая валюта: %s", currency)
	}

	row := tx.QueryRowContext(ctx, query, amount, userID)
	var balance models.Balance
	err := row.Scan(&balance.USD, &balance.RUB, &balance.EUR)
	if err != nil {
		return nil, err
	}

	return &balance, nil
}

// NewPostgresStorage создает новое подключение к PostgreSQL
func NewPostgresStorage(connString string) (*PostgresStorage, error) {
	// Получаем параметры подключения из переменных окружения
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Основная строка подключения
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	adminConn, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}
	defer adminConn.Close()

	// Проверка существования БД
	var exists bool
	err = adminConn.QueryRowContext(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'wallet_db')").Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки существования БД: %w", err)
	}

	// Создание БД если не существует
	if !exists {
		_, err = adminConn.ExecContext(context.Background(), "CREATE DATABASE wallet_db")
		if err != nil {
			return nil, fmt.Errorf("ошибка создания БД: %w", err)
		}
	}

	// Подключение к конкретной БД
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}

	// Проверка подключения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ошибка проверки соединения: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Применение миграций
	if err := applyMigrations(db); err != nil {
		return nil, fmt.Errorf("ошибка применения миграций: %w", err)
	}

	log.Println("Успешное подключение к PostgreSQL")

	return &PostgresStorage{db: db}, nil
}

// applyMigrations создает таблицы если они не существуют
func applyMigrations(db *sql.DB) error {
	// Создание таблицы пользователей
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(100) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы пользователей: %w", err)
	}

	// Создание таблицы кошельков
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS wallets (
			user_id INTEGER PRIMARY KEY REFERENCES users(id),
			usd DECIMAL(15, 2) DEFAULT 0.00,
			rub DECIMAL(15, 2) DEFAULT 0.00,
			eur DECIMAL(15, 2) DEFAULT 0.00,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы кошельков: %w", err)
	}

	return nil
}

// Close закрывает подключение к базе данных
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

// GetUserRepository возвращает реализацию UserRepository
func (s *PostgresStorage) GetUserRepository() storage.UserRepository {
	return &userRepository{db: s.db}
}

// GetWalletRepository возвращает реализацию WalletRepository
func (s *PostgresStorage) GetWalletRepository() storage.WalletRepository {
	return &walletRepository{db: s.db}
}
