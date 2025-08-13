package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gw-currency-wallet/internal/storage/redis"
	pb "gw-proto/proto" // Импорт сгенерированного protobuf кода
	"time"
)

// ExchangeService предоставляет функционал для работы с курсами валют
// Использует:
// - gRPC клиент для получения актуальных курсов
// - Redis для кэширования результатов
type ExchangeService struct {
	client        pb.ExchangeServiceClient // gRPC клиент для сервиса курсов
	conn          *grpc.ClientConn         // gRPC соединение
	redisClient   *redis.Client            // Клиент Redis для кэширования
	cacheDuration time.Duration            // Время жизни кэша
}

// NewExchangeService создает новый экземпляр ExchangeService
// Параметры:
//   - addr: адрес gRPC сервиса курсов валют
//   - redisAddr: адрес Redis сервера
//   - cacheDuration: время жизни кэша (например 5m)
//
// Возвращает:
//   - *ExchangeService: инициализированный сервис
//   - error: ошибка при создании
func NewExchangeService(addr string, redisAddr string, cacheDuration time.Duration) (*ExchangeService, error) {
	if addr == "" {
		return nil, errors.New("адрес сервиса обмена не может быть пустым")
	}

	// Устанавливаем соединение с gRPC сервером
	// 1. Создаем контекст с таймаутом подключения
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 2. Устанавливаем соединение с современными параметрами
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second, // Минимальное время попытки подключения
		}),
	)

	// Инициализация Redis клиента
	redisClient, err := redis.New(redis.Options{
		Addr:     redisAddr,
		Password: "", // Пароль, если требуется
		DB:       0,  // Номер базы данных
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к Redis: %w", err)
	}

	return &ExchangeService{
		client:        pb.NewExchangeServiceClient(conn),
		conn:          conn,
		redisClient:   redisClient,
		cacheDuration: cacheDuration,
	}, nil
}

// GetRates возвращает текущие курсы валют
// Сначала проверяет кэш в Redis, если нет - запрашивает через gRPC
// Возвращает:
//   - map[string]float64: курс валют (например {"USD": 75.50})
//   - error: ошибка при получении
func (s *ExchangeService) GetRates(ctx context.Context) (map[string]float64, error) {
	if s == nil {
		return nil, errors.New("сервис обмена не инициализирован")
	}

	// Пробуем получить из кэша Redis
	cacheKey := "exchange:rates"
	cachedRates, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var rates map[string]float64
		if err := json.Unmarshal(cachedRates, &rates); err == nil {
			return s.filterRates(rates), nil // Возвращаем отфильтрованные курсы
		}
	}

	// Запрашиваем актуальные курсы через gRPC
	rates, err := s.client.GetExchangeRates(ctx, &pb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("ошибка получения курсов от gRPC сервиса: %w", err)
	}

	// Конвертируем protobuf в map
	result := make(map[string]float64)
	for k, v := range rates.Rates {
		result[k] = float64(v)
	}

	// Сохраняем в кэш
	ratesJSON, err := json.Marshal(result)
	if err == nil {
		s.redisClient.Set(ctx, cacheKey, ratesJSON, s.cacheDuration)
	}

	return s.filterRates(result), nil
}

// filterRates оставляет только поддерживаемые валюты (USD, EUR, RUB)
func (s *ExchangeService) filterRates(rates map[string]float64) map[string]float64 {
	res := make(map[string]float64)
	for _, currency := range []string{"USD", "EUR", "RUB"} {
		if rate, ok := rates[currency]; ok {
			res[currency] = rate
		}
	}
	return res
}

// GetRate возвращает курс обмена между двумя валютами
// Поддерживает конвертацию через RUB как базовую валюту
// Параметры:
//   - from: исходная валюта
//   - to: целевая валюта
//
// Возвращает:
//   - float64: курс обмена
//   - error: ошибка при расчете
func (s *ExchangeService) GetRate(ctx context.Context, from, to string) (float64, error) {
	// Если валюты одинаковые
	if from == to {
		return 1.0, nil
	}

	rates, err := s.GetRates(ctx)
	if err != nil {
		return 0, err
	}

	// Проверка доступности базовых курсов
	if rates["USD"] == 0 || rates["EUR"] == 0 {
		return 0, errors.New("базовые курсы недоступны")
	}

	// Логика конвертации через RUB
	switch {
	// Прямые котировки с RUB
	case from == "RUB" && to == "USD":
		return 1 / rates["USD"], nil // 1 RUB = (1/USD_RATE) USD
	case from == "USD" && to == "RUB":
		return rates["USD"], nil // 1 USD = USD_RATE RUB
	case from == "RUB" && to == "EUR":
		return 1 / rates["EUR"], nil // 1 RUB = (1/EUR_RATE) EUR
	case from == "EUR" && to == "RUB":
		return rates["EUR"], nil // 1 EUR = EUR_RATE RUB

	// Кросс-курсы через RUB
	case from == "USD" && to == "EUR":
		return (1 / rates["EUR"]) * rates["USD"], nil // USD->RUB->EUR
	case from == "EUR" && to == "USD":
		return (1 / rates["USD"]) * rates["EUR"], nil // EUR->RUB->USD

	default:
		return 0, errors.New("неподдерживаемая валютная пара")
	}
}

// Close освобождает ресурсы (gRPC и Redis соединения)
func (s *ExchangeService) Close() error {
	if s.conn != nil {
		_ = s.conn.Close()
	}
	if s.redisClient != nil {
		_ = s.redisClient.Close()
	}
	return nil
}
