package telegram

import (
	"context"

	"google.golang.org/grpc"
	pb "gw-proto/proto" // Импорт сгенерированного gRPC-кода
)

// ExchangeService представляет сервис для взаимодействия с gRPC-сервером,
// предоставляющим данные о курсах валют.
type ExchangeService struct {
	client pb.ExchangeServiceClient // gRPC-клиент для вызова методов сервера
}

// NewExchangeService создает новый экземпляр ExchangeService.
// Параметры:
//   - conn: установленное gRPC-соединение с сервером
//
// Возвращает:
//   - Указатель на созданный ExchangeService
func NewExchangeService(conn *grpc.ClientConn) *ExchangeService {
	return &ExchangeService{
		client: pb.NewExchangeServiceClient(conn), // Инициализация gRPC-клиента
	}
}

// GetAllRates запрашивает у gRPC-сервера все текущие курсы валют и возвращает их в виде map.
// Возвращаемые значения:
//   - map[string]float64: словарь, где ключ — код валюты (например, "USD"), значение — курс к рублю
//   - error: ошибка, если запрос к серверу не удался
func (s *ExchangeService) GetAllRates() (map[string]float64, error) {
	// Вызов gRPC-метода GetExchangeRates с пустым запросом (pb.Empty)
	rates, err := s.client.GetExchangeRates(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, err // Возвращаем ошибку, если запрос не удался
	}

	// Конвертируем полученные курсы из protobuf-формата в map[string]float64
	result := make(map[string]float64)
	for currency, rate := range rates.Rates {
		result[currency] = float64(rate) // Преобразуем тип rate (предположительно float32) в float64
	}

	return result, nil
}
