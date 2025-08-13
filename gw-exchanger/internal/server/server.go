package server

import (
	"context"
	"fmt"
	"google.golang.org/grpc"                 // Фреймворк для работы с gRPC
	"gw-exchanger/internal/storage/postgres" // Реализация хранилища данных
	"gw-proto/proto"                         // Сгенерированный Protobuf код
	"log"
	"net"
)

// ExchangeServer реализует gRPC сервис для работы с курсами валют
type ExchangeServer struct {
	proto.UnimplementedExchangeServiceServer                           // Обязательная встроенная реализация
	storage                                  *postgres.PostgresStorage // Хранилище данных (PostgreSQL)
}

// NewServer создает новый экземпляр gRPC сервера
// Параметры:
//   - storage: подключение к хранилищу данных
//
// Возвращает:
//   - *ExchangeServer: готовый к работе сервер
func NewServer(storage *postgres.PostgresStorage) *ExchangeServer {
	return &ExchangeServer{storage: storage}
}

// GetExchangeRates возвращает все текущие курсы валют
// Параметры:
//   - ctx: контекст выполнения
//   - req: пустой запрос (proto.Empty)
//
// Возвращает:
//   - *proto.ExchangeRatesResponse: список всех курсов валют
//   - error: ошибка при получении данных
func (s *ExchangeServer) GetExchangeRates(ctx context.Context, req *proto.Empty) (*proto.ExchangeRatesResponse, error) {
	// Получаем курсы из хранилища
	rates, err := s.storage.GetAllRates(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения курсов: %v", err)
	}

	// Конвертируем map[string]float64 в map[string]float32 для gRPC
	response := make(map[string]float32, len(rates))
	for currency, rate := range rates {
		response[currency] = float32(rate)
	}

	return &proto.ExchangeRatesResponse{Rates: response}, nil
}

// GetExchangeRateForCurrency возвращает курс для конкретной пары валют
// Параметры:
//   - ctx: контекст выполнения
//   - req: запрос с кодами валют (proto.CurrencyRequest)
//
// Возвращает:
//   - *proto.ExchangeRateResponse: курс обмена между валютами
//   - error: ошибка при получении данных
func (s *ExchangeServer) GetExchangeRateForCurrency(ctx context.Context, req *proto.CurrencyRequest) (*proto.ExchangeRateResponse, error) {
	// Получаем курс из хранилища
	rate, err := s.storage.GetRate(ctx, req.FromCurrency, req.ToCurrency)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения курса: %v", err)
	}

	return &proto.ExchangeRateResponse{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		Rate:         float32(rate),
	}, nil
}

// Start запускает gRPC сервер на указанном порту
// Параметры:
//   - port: порт для прослушивания (например "50051")
//   - storage: подключение к хранилищу данных
func Start(port string, storage *postgres.PostgresStorage) {
	// Создаем TCP listener на указанном порту
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("ошибка запуска сервера: %v", err)
	}

	// Создаем новый экземпляр gRPC сервера
	grpcServer := grpc.NewServer()

	// Регистрируем наш сервис ExchangeService
	proto.RegisterExchangeServiceServer(grpcServer, NewServer(storage))

	log.Printf("Сервер запущен на порту %s", port)

	// Запускаем сервер (блокирующая операция)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("ошибка работы сервера: %v", err)
	}
}
