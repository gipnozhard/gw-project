package storages

import "time"

// ExchangeRate представляет запись о курсе валюты в хранилище
// Содержит поля, соответствующие структуре таблицы в БД
type ExchangeRate struct {
	Currency  string    `json:"currency" db:"currency"`     // Код валюты (например "USD")
	Rate      float64   `json:"rate" db:"rate"`             // Текущий курс к базовой валюте
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Время последнего обновления
}

// RateRequest содержит параметры запроса курса обмена
// Используется в API для входящих запросов
type RateRequest struct {
	From string `json:"from"` // Код исходной валюты
	To   string `json:"to"`   // Код целевой валюты
}

// RateResponse представляет ответ с курсом обмена
// Используется в API для исходящих ответов
type RateResponse struct {
	From      string  `json:"from"`      // Исходная валюта
	To        string  `json:"to"`        // Целевая валюта
	Rate      float64 `json:"rate"`      // Рассчитанный курс обмена
	Timestamp int64   `json:"timestamp"` // Время расчета в Unix timestamp
}

// AllRatesResponse представляет ответ со всеми текущими курсами
// Используется в API для исходящих ответов
type AllRatesResponse struct {
	Rates     map[string]float64 `json:"rates"`     // Словарь всех курсов (ключ - код валюты)
	Timestamp int64              `json:"timestamp"` // Время актуальности данных в Unix timestamp
}
