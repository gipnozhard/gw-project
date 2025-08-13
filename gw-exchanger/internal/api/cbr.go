package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CBRResponse представляет структуру ответа от API Центрального Банка России
// Содержит дату актуальности курсов и словарь с курсами валют
type CBRResponse struct {
	Date  string            `json:"Date"`   // Дата обновления курсов в формате строки
	Rates map[string]CBRate `json:"Valute"` // Словарь курсов валют, где ключ - код валюты
}

// CBRate содержит детальную информацию о курсе конкретной валюты
type CBRate struct {
	CharCode string  `json:"CharCode"` // Буквенный код валюты (например: USD, EUR)
	Nominal  int     `json:"Nominal"`  // Номинал (например: 1 USD, 10 JPY)
	Name     string  `json:"Name"`     // Название валюты на русском
	Value    float64 `json:"Value"`    // Стоимость номинала в рублях
}

// FetchCBExchangeRates получает актуальные курсы валют от API ЦБ РФ
// Параметры:
//   - url: адрес API Центробанка (например: "https://www.cbr-xml-daily.ru/daily_json.js")
//
// Возвращает:
//   - map[string]float64: словарь с курсами валют (ключ - код валюты, значение - курс к рублю)
//   - error: ошибка при получении или обработке данных
func FetchCBExchangeRates(url string) (map[string]float64, error) {
	// 1. Отправка HTTP GET запроса к API Центробанка
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения курсов: %v", err)
	}
	defer resp.Body.Close() // Гарантированное закрытие тела ответа

	// 2. Чтение тела ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	// 3. Парсинг JSON данных в структуру CBRResponse
	var data CBRResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("ошибка разбора JSON: %v", err)
	}

	// 4. Подготовка результата - нормализация курсов к 1 единице валюты
	rates := make(map[string]float64)
	for _, rate := range data.Rates {
		// Пересчитываем курс на 1 единицу валюты (делим на номинал)
		rates[rate.CharCode] = rate.Value / float64(rate.Nominal)
	}

	// 5. Добавляем рубль с курсом 1.0 для консистентности
	rates["RUB"] = 1.0

	return rates, nil
}
