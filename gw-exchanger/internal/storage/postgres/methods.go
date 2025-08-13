package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"gw-exchanger/internal/api"
	"log"
	"time"
)

// startRateUpdater запускает фоновое обновление курсов валют
func (s *PostgresStorage) startRateUpdater() {
	ticker := time.NewTicker(s.updateInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.UpdateRatesFromCB(); err != nil {
			log.Printf("Ошибка обновления курсов: %v", err)
		}
	}
}

// UpdateRatesFromCB обновляет курсы валют из API Центробанка
func (s *PostgresStorage) UpdateRatesFromCB() error {
	if s.apiURL == "" {
		return fmt.Errorf("URL API не настроен")
	}

	log.Println("Обновление курсов валют...")

	// 1. Получение курсов от API
	rates, err := api.FetchCBExchangeRates(s.apiURL)
	if err != nil {
		return fmt.Errorf("ошибка получения курсов: %v", err)
	}

	// 2. Начало транзакции
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %v", err)
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {

		}
	}(tx)

	// 3. Обновление курсов в БД
	for currency, rate := range rates {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO exchange_rates (currency, rate) 
			 VALUES ($1, $2)
			 ON CONFLICT (currency) DO UPDATE SET rate = $2, updated_at = NOW()`,
			currency, rate)
		if err != nil {
			return fmt.Errorf("ошибка обновления курса %s: %v", currency, err)
		}
	}

	// 4. Фиксация транзакции
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка фиксации транзакции: %v", err)
	}

	log.Println("Курсы валют успешно обновлены")
	return nil
}

// GetRate возвращает курс обмена между двумя валютами
// Параметры:
//   - ctx: контекст выполнения
//   - from: исходная валюта
//   - to: целевая валюта
//
// Возвращает:
//   - float64: курс обмена
//   - error: ошибка при получении
func (s *PostgresStorage) GetRate(ctx context.Context, from, to string) (float64, error) {
	if from == to {
		return 1.0, nil // Курс одинаковых валют всегда 1
	}

	// Случай 1: Одна из валют - USD
	if from == "USD" || to == "USD" {
		targetCurrency := to
		if from == "USD" {
			targetCurrency = to
		} else {
			targetCurrency = from
		}

		query := "SELECT rate FROM exchange_rates WHERE currency = $1"
		var rate float64
		err := s.db.QueryRowContext(ctx, query, targetCurrency).Scan(&rate)

		if err != nil {
			return 0, fmt.Errorf("курс для %s не найден: %v", targetCurrency, err)
		}

		if from == "USD" {
			return rate, nil // Прямой курс (USD -> другая валюта)
		}
		return 1 / rate, nil // Обратный курс (другая валюта -> USD)
	}

	// Случай 2: Кросс-курс (через USD)
	rateFromUSD, err := s.GetRate(ctx, from, "USD")
	if err != nil {
		return 0, err
	}
	rateToUSD, err := s.GetRate(ctx, "USD", to)
	if err != nil {
		return 0, err
	}
	return rateFromUSD * rateToUSD, nil // Расчет кросс-курса
}

// GetAllRates возвращает все текущие курсы валют
func (s *PostgresStorage) GetAllRates(ctx context.Context) (map[string]float64, error) {
	query := "SELECT currency, rate FROM exchange_rates"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса курсов: %v", err)
	}
	defer rows.Close()

	rates := make(map[string]float64)
	for rows.Next() {
		var currency string
		var rate float64
		if err := rows.Scan(&currency, &rate); err != nil {
			return nil, fmt.Errorf("ошибка чтения данных: %v", err)
		}
		rates[currency] = rate
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка обработки результатов: %v", err)
	}

	return rates, nil
}
