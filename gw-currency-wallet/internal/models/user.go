package models

import (
	"time"
)

// User представляет основную модель пользователя в системе
// swagger:model User
type User struct {
	ID           int       `json:"id" db:"id"`                 // Уникальный идентификатор пользователя
	Username     string    `json:"username" db:"username"`     // Логин пользователя (уникальный)
	Email        string    `json:"email" db:"email"`           // Email пользователя (уникальный)
	PasswordHash string    `json:"-" db:"password_hash"`       // Хэш пароля (никогда не возвращается в API)
	CreatedAt    time.Time `json:"created_at" db:"created_at"` // Дата создания записи
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"` // Дата последнего обновления
}

// CreateUserRequest - запрос на регистрацию нового пользователя
// swagger:model CreateUserRequest
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"` // Логин (3-50 символов)
	Email    string `json:"email" validate:"required,email"`           // Валидный email
	Password string `json:"password" validate:"required,min=8"`        // Пароль (мин. 8 символов)
}

// LoginRequest - запрос на аутентификацию пользователя
// swagger:model LoginRequest
type LoginRequest struct {
	Username string `json:"username" validate:"required"` // Логин пользователя
	Password string `json:"password" validate:"required"` // Пароль пользователя
}

// Balance - модель баланса пользователя по валютам
// swagger:model Balance
type Balance struct {
	USD float64 `json:"USD" db:"USD"` // Сумма в долларах
	RUB float64 `json:"RUB" db:"RUB"` // Сумма в рублях
	EUR float64 `json:"EUR" db:"EUR"` // Сумма в евро
}

// DepositRequest - запрос на пополнение баланса
// swagger:model DepositRequest
type DepositRequest struct {
	Amount   float64 `json:"amount" validate:"required,gt=0"`                // Сумма пополнения (>0)
	Currency string  `json:"currency" validate:"required,oneof=USD RUB EUR"` // Валюта (USD/RUB/EUR)
}

// WithdrawRequest - запрос на снятие средств
// swagger:model WithdrawRequest
type WithdrawRequest struct {
	Amount   float64 `json:"amount" validate:"required,gt=0"`                // Сумма снятия (>0)
	Currency string  `json:"currency" validate:"required,oneof=USD RUB EUR"` // Валюта (USD/RUB/EUR)
}

// ExchangeRatesResponse - ответ с текущими курсами валют
// swagger:model ExchangeRatesResponse
type ExchangeRatesResponse struct {
	Rates map[string]float64 `json:"rates"` // Карта курсов (например: {"USD":1,"RUB":75.5})
}

// ExchangeRequest - запрос на обмен валюты
// swagger:model ExchangeRequest
type ExchangeRequest struct {
	FromCurrency string  `json:"from_currency" validate:"required,oneof=USD RUB EUR"` // Исходная валюта
	ToCurrency   string  `json:"to_currency" validate:"required,oneof=USD RUB EUR"`   // Целевая валюта
	Amount       float64 `json:"amount" validate:"required,gt=0"`                     // Сумма для обмена (>0)
}

// ExchangeResponse - результат операции обмена валют
// swagger:model ExchangeResponse
type ExchangeResponse struct {
	Message         string   `json:"message"`          // Сообщение о результате
	ExchangedAmount float64  `json:"exchanged_amount"` // Полученная сумма
	NewBalance      *Balance `json:"new_balance"`      // Обновленный баланс
	Rate            float64  `json:"rate"`             // Примененный курс обмена
}

// Wallet - модель кошелька пользователя в БД
// swagger:model Wallet
type Wallet struct {
	UserID    int       `json:"user_id" db:"user_id"`       // Ссылка на пользователя
	USD       float64   `json:"USD" db:"USD"`               // Баланс USD
	RUB       float64   `json:"RUB" db:"RUB"`               // Баланс RUB
	EUR       float64   `json:"EUR" db:"EUR"`               // Баланс EUR
	CreatedAt time.Time `json:"created_at" db:"created_at"` // Дата создания
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Дата обновления
}

// ErrorResponse - стандартный ответ при ошибке
// swagger:model ErrorResponse
type ErrorResponse struct {
	Error string `json:"error"` // Описание ошибки
}

// SuccessMessage - стандартный успешный ответ
// swagger:model SuccessMessage
type SuccessMessage struct {
	Message string `json:"message"` // Информационное сообщение
	UserID  int    `json:"user_id"` // ID пользователя (если применимо)
}

// LoginResponse - ответ с JWT токеном при успешной аутентификации
// swagger:model LoginResponse
type LoginResponse struct {
	Token string `json:"token"` // JWT токен для авторизации
}

// TransactionRequest - обобщенный запрос для операций с балансом
// swagger:model TransactionRequest
type TransactionRequest struct {
	Amount   float64 `json:"amount" validate:"required,gt=0"`                // Сумма операции (>0)
	Currency string  `json:"currency" validate:"required,oneof=USD RUB EUR"` // Валюта операции
}

// TransactionResponse - обобщенный ответ для операций с балансом
// swagger:model TransactionResponse
type TransactionResponse struct {
	Message    string   `json:"message"`     // Сообщение о результате
	NewBalance *Balance `json:"new_balance"` // Обновленный баланс
}
