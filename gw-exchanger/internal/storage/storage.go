package storages

import (
	"context"
	"time"
)

// Storage - основной интерфейс хранилища курсов валют.
// Объединяет функциональность для работы с курсами (RateProvider),
// их обновления (Updater) и управления ресурсами (Closer).
type Storage interface {
	RateProvider  // Методы для получения курсов
	Updater       // Методы для обновления данных
	Close() error // Метод для освобождения ресурсов
}

// RateProvider предоставляет методы для доступа к курсам валют
type RateProvider interface {
	// GetRate возвращает курс конвертации между двумя валютами
	// Параметры:
	//   - ctx: контекст выполнения
	//   - from: код исходной валюты (например "USD")
	//   - to: код целевой валюты (например "RUB")
	// Возвращает:
	//   - float64: курс обмена
	//   - error: ошибка при получении курса
	GetRate(ctx context.Context, from, to string) (float64, error)

	// GetAllRates возвращает все доступные курсы валют
	// Параметры:
	//   - ctx: контекст выполнения
	// Возвращает:
	//   - map[string]float64: словарь курсов (ключ - код валюты)
	//   - error: ошибка при получении данных
	GetAllRates(ctx context.Context) (map[string]float64, error)
}

// Updater предоставляет методы для обновления курсов валют
type Updater interface {
	// UpdateRates выполняет обновление курсов из внешнего источника
	// Возвращает:
	//   - error: ошибка при обновлении данных
	UpdateRates() error
}

// UpdaterConfig содержит параметры для фонового обновления курсов
type UpdaterConfig struct {
	Enabled        bool          // Флаг активности автоматического обновления
	UpdateInterval time.Duration // Интервал между обновлениями (например 1h)
	InitialDelay   time.Duration // Задержка перед первым обновлением
}
