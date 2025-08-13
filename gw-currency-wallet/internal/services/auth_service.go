package services

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt" // Пакет для безопасного хеширования паролей
	"gw-currency-wallet/internal/middleware"
	"gw-currency-wallet/internal/models"
	"gw-currency-wallet/internal/storage"
	"time"
)

// AuthService предоставляет функционал для регистрации и аутентификации пользователей.
// Содержит зависимости:
// - repo: для операций с хранилищем пользователей
// - jwtSecret: секретный ключ для подписи JWT
// - tokenExpiration: срок действия токена
type AuthService struct {
	repo            storage.UserRepository
	jwtSecret       string
	tokenExpiration time.Duration
}

// NewAuthService - конструктор для создания экземпляра AuthService.
// Инициализирует сервис с переданными параметрами:
// - repo: реализация интерфейса работы с хранилищем пользователей
// - jwtSecret: секретный ключ для генерации/верификации токенов
// - tokenExpiration: длительность жизни токена (например 24h)
//
// Возвращает готовый к использованию экземпляр AuthService.
func NewAuthService(repo storage.UserRepository, jwtSecret string, tokenExpiration time.Duration) *AuthService {
	return &AuthService{
		repo:            repo,
		jwtSecret:       jwtSecret,
		tokenExpiration: tokenExpiration,
	}
}

// Register регистрирует нового пользователя в системе.
// Последовательность операций:
// 1. Проверка уникальности имени пользователя
// 2. Хеширование пароля
// 3. Создание записи пользователя
//
// Параметры:
// - ctx: контекст выполнения
// - req: данные для регистрации (username, email, password)
//
// Возвращает:
// - *models.User: данные зарегистрированного пользователя
// - error: ошибка при возникновении проблем
func (s *AuthService) Register(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
	// Проверка существующего пользователя
	existing, _ := s.repo.GetUserByUsername(ctx, req.Username)
	if existing != nil {
		return nil, errors.New("пользователь с таким именем уже существует") // Сообщение об ошибке
	}

	// Генерация хеша пароля с стандартной стоимостью вычисления
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err // Возвращаем ошибку хеширования
	}

	// Создание объекта пользователя
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword), // Сохранение хеша вместо пароля
	}

	// Сохранение пользователя в хранилище
	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err // Ошибка сохранения
	}

	return user, nil
}

// Login выполняет аутентификацию пользователя и генерирует JWT токен.
// Алгоритм работы:
// 1. Поиск пользователя по username
// 2. Сравнение хеша пароля
// 3. Генерация токена при успешной проверке
//
// Параметры:
// - ctx: контекст выполнения
// - username: имя пользователя
// - password: пароль пользователя
//
// Возвращает:
// - string: JWT токен для доступа
// - error: ошибка аутентификации
func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	// Получение пользователя из хранилища
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil || user == nil {
		// Обобщенное сообщение для безопасности (не раскрываем детали)
		return "", errors.New("неверные учетные данные")
	}

	// Сравнение хеша пароля с предоставленным паролем
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("неверные учетные данные")
	}

	// Генерация JWT токена с указанными параметрами
	token, err := middleware.GenerateJWTToken(
		user.ID,           // ID пользователя в claims
		s.jwtSecret,       // Секретный ключ
		s.tokenExpiration, // Время жизни токена
	)
	if err != nil {
		return "", errors.New("ошибка генерации токена")
	}

	return token, nil
}
