package storage

import (
	"context"
	"gw-currency-wallet/internal/models"
)

// UserRepository определяет контракт для работы с данными пользователей
// Интерфейс абстрагирует работу с хранилищем и позволяет легко подменять реализации
type UserRepository interface {
	// CreateUser создает нового пользователя в хранилище
	// Принимает:
	//   - ctx: контекст для контроля времени выполнения и отмены
	//   - user: указатель на объект пользователя для создания
	// Возвращает:
	//   - error: ошибка при создании (например, если пользователь уже существует)
	CreateUser(ctx context.Context, user *models.User) error

	// GetUserByUsername находит пользователя по имени пользователя
	// Принимает:
	//   - ctx: контекст выполнения
	//   - username: имя пользователя для поиска
	// Возвращает:
	//   - *models.User: найденный пользователь или nil если не найден
	//   - error: ошибка при выполнении запроса
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)

	// GetUserByEmail находит пользователя по email
	// Принимает:
	//   - ctx: контекст выполнения
	//   - email: email для поиска
	// Возвращает:
	//   - *models.User: найденный пользователь или nil если не найден
	//   - error: ошибка при выполнении запроса
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)

	// GetUserByID находит пользователя по идентификатору
	// Принимает:
	//   - ctx: контекст выполнения
	//   - id: числовой идентификатор пользователя
	// Возвращает:
	//   - *models.User: найденный пользователь или nil если не найден
	//   - error: ошибка при выполнении запроса
	GetUserByID(ctx context.Context, id int) (*models.User, error)
}

// WalletRepository определяет контракт для работы с финансовыми операциями
// Интерфейс обеспечивает абстракцию над конкретной реализацией хранилища кошельков
type WalletRepository interface {
	// GetBalance возвращает текущий баланс пользователя по всем валютам
	// Принимает:
	//   - ctx: контекст выполнения
	//   - userID: идентификатор пользователя
	// Возвращает:
	//   - *models.Balance: структура с балансами по валютам
	//   - error: ошибка при получении баланса
	GetBalance(ctx context.Context, userID int) (*models.Balance, error)

	// CreateWallet создает новый кошелек для пользователя
	// Принимает:
	//   - ctx: контекст выполнения
	//   - userID: идентификатор пользователя
	// Возвращает:
	//   - error: ошибка при создании
	CreateWallet(ctx context.Context, userID int) error

	// UpdateBalance изменяет баланс пользователя для указанной валюты
	// Принимает:
	//   - ctx: контекст выполнения
	//   - userID: идентификатор пользователя
	//   - currency: валюта для изменения (USD, RUB, EUR)
	//   - amount: сумма для изменения (может быть отрицательной)
	// Возвращает:
	//   - *models.Balance: новый баланс после изменения
	//   - error: ошибка при обновлении
	UpdateBalance(ctx context.Context, userID int, currency string, amount float64) (*models.Balance, error)

	// Transfer выполняет перевод средств между пользователями
	// Должен выполняться атомарно в рамках транзакции
	// Принимает:
	//   - ctx: контекст выполнения
	//   - fromUserID: ID отправителя
	//   - toUserID: ID получателя
	//   - currency: валюта перевода
	//   - amount: сумма перевода
	// Возвращает:
	//   - *models.Balance: новый баланс отправителя
	//   - *models.Balance: новый баланс получателя
	//   - error: ошибка при переводе
	Transfer(
		ctx context.Context,
		fromUserID int,
		toUserID int,
		currency string,
		amount float64,
	) (*models.Balance, *models.Balance, error)

	// Exchange выполняет обмен валюты для пользователя
	// Должен выполняться атомарно в рамках транзакции
	// Принимает:
	//   - ctx: контекст выполнения
	//   - userID: ID пользователя
	//   - fromCurrency: исходная валюта
	//   - toCurrency: целевая валюта
	//   - amount: сумма для обмена
	//   - rate: курс обмена
	// Возвращает:
	//   - *models.Balance: новый баланс после обмена
	//   - error: ошибка при обмене
	Exchange(
		ctx context.Context,
		userID int,
		fromCurrency string,
		toCurrency string,
		amount float64,
		rate float64,
	) (*models.Balance, error)
}
