package utils

import (
	"context"
	"fmt"
	"gw-exchanger/internal/storage/postgres"
	"log"
	"time"
)

// PrintAvailableCurrencies выводит список доступных валют и их курсов к рублю
// Параметры:
//   - storage: подключение к хранилищу данных (PostgreSQL)
//
// Логика работы:
//  1. Создает контекст с таймаутом 3 секунды для запроса
//  2. Получает все курсы валют из хранилища
//  3. Форматирует и выводит результат:
//     - USD всегда выводится первым как базовая валюта
//     - Остальные валюты выводятся в алфавитном порядке
//  4. Обрабатывает возможные ошибки
func PrintAvailableCurrencies(storage *postgres.PostgresStorage) {
	// Создаем контекст с ограничением времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // Гарантированное освобождение ресурсов

	// Получаем все курсы валют из хранилища
	currencies, err := storage.GetAllRates(ctx)
	if err != nil {
		log.Printf("Ошибка получения списка валют: %v", err)
		return
	}

	// Форматированный вывод заголовка
	fmt.Println("\n=== Доступные валюты ===")

	// Обработка случая пустой базы данных
	if len(currencies) == 0 {
		fmt.Println("В базе данных не найдено курсов валют!")
	} else {
		// Специальный вывод для USD (базовая валюта)
		if usdRate, exists := currencies["USD"]; exists {
			fmt.Printf("%.4f RUB (базовая валюта) = 1 USD\n", usdRate)
		}

		// Вывод остальных валют в алфавитном порядке
		for currency, rate := range currencies {
			if currency != "USD" {
				fmt.Printf("%.4f RUB = 1 %s\n", rate, currency)
			}
		}
	}

	// Нижний разделитель
	fmt.Println("===========================")
}
