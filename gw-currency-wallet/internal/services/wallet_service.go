package services

import (
	"context"
	"errors"
	"fmt"
	"gw-currency-wallet/internal/models"
	"gw-currency-wallet/internal/storage"
	"log"
)

// RateProvider определяет интерфейс для работы с сервисом курсов валют
// Это позволяет абстрагироваться от конкретной реализации и легко подменять сервис курсов
type RateProvider interface {
	GetRate(ctx context.Context, fromCurrency, toCurrency string) (float64, error)
}

// WalletService реализует бизнес-логику работы с кошельком пользователя
type WalletService struct {
	repo        storage.WalletRepository // Репозиторий для работы с данными кошелька
	rateService RateProvider             // Сервис для получения курсов валют
}

// NewWalletService создает новый экземпляр WalletService
// Параметры:
//   - repo: репозиторий для работы с хранилищем кошельков
//   - rateService: сервис для получения курсов валют
//
// Возвращает:
//   - *WalletService: инициализированный сервис работы с кошельком
func NewWalletService(repo storage.WalletRepository, rateService RateProvider) *WalletService {
	return &WalletService{
		repo:        repo,
		rateService: rateService,
	}
}

// GetBalance возвращает баланс пользователя по всем валютам
// Параметры:
//   - ctx: контекст выполнения
//   - userID: идентификатор пользователя
//
// Возвращает:
//   - *models.Balance: текущий баланс
//   - error: ошибка при получении баланса
func (s *WalletService) GetBalance(ctx context.Context, userID int) (*models.Balance, error) {
	if userID <= 0 {
		return nil, errors.New("неверный ID пользователя") // Валидация входных данных
	}

	return s.repo.GetBalance(ctx, userID) // Делегируем получение баланса репозиторию
}

// Deposit пополняет баланс пользователя в указанной валюте
// Параметры:
//   - ctx: контекст выполнения
//   - userID: идентификатор пользователя
//   - currency: валюта пополнения (USD, RUB, EUR)
//   - amount: сумма пополнения
//
// Возвращает:
//   - *models.Balance: новый баланс после пополнения
//   - error: ошибка при выполнении операции
func (s *WalletService) Deposit(ctx context.Context, userID int, currency string, amount float64) (*models.Balance, error) {
	// Валидация входных параметров
	if userID <= 0 {
		return nil, errors.New("неверный ID пользователя")
	}

	if !isValidCurrency(currency) {
		return nil, fmt.Errorf("неподдерживаемая валюта: %s", currency)
	}

	if amount <= 0 {
		return nil, errors.New("сумма должна быть положительной")
	}

	// Выполняем операцию пополнения через репозиторий
	return s.repo.UpdateBalance(ctx, userID, currency, amount)
}

// Withdraw снимает средства с баланса пользователя
// Параметры:
//   - ctx: контекст выполнения
//   - userID: идентификатор пользователя
//   - currency: валюта снятия (USD, RUB, EUR)
//   - amount: сумма снятия
//
// Возвращает:
//   - *models.Balance: новый баланс после снятия
//   - error: ошибка при выполнении операции
func (s *WalletService) Withdraw(ctx context.Context, userID int, currency string, amount float64) (*models.Balance, error) {
	// Валидация входных параметров
	if userID <= 0 {
		return nil, errors.New("неверный ID пользователя")
	}

	if !isValidCurrency(currency) {
		return nil, fmt.Errorf("неподдерживаемая валюта: %s", currency)
	}

	if amount <= 0 {
		return nil, errors.New("сумма должна быть положительной")
	}

	// Получаем текущий баланс
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения баланса: %w", err)
	}

	// Проверяем достаточность средств
	currentBalance, err := getBalanceByCurrency(balance, currency)
	if err != nil {
		return nil, err
	}

	if currentBalance < amount {
		return nil, errors.New("недостаточно средств")
	}

	// Выполняем операцию снятия (передаем отрицательное значение)
	return s.repo.UpdateBalance(ctx, userID, currency, -amount)
}

// Exchange выполняет обмен валюты по текущему курсу
// Параметры:
//   - ctx: контекст выполнения
//   - userID: идентификатор пользователя
//   - fromCurrency: исходная валюта
//   - toCurrency: целевая валюта
//   - amount: сумма для обмена
//
// Возвращает:
//   - *models.ExchangeResponse: результат обмена
//   - error: ошибка при выполнении операции
func (s *WalletService) Exchange(
	ctx context.Context,
	userID int,
	fromCurrency string,
	toCurrency string,
	amount float64,
) (*models.ExchangeResponse, error) {
	// Валидация входных параметров
	if userID <= 0 {
		return nil, errors.New("неверный ID пользователя")
	}

	if !isValidCurrency(fromCurrency) {
		return nil, fmt.Errorf("неподдерживаемая исходная валюта: %s", fromCurrency)
	}

	if !isValidCurrency(toCurrency) {
		return nil, fmt.Errorf("неподдерживаемая целевая валюта: %s", toCurrency)
	}

	if amount <= 0 {
		return nil, errors.New("сумма должна быть положительной")
	}

	// Получаем текущий курс обмена
	rate, err := s.rateService.GetRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		log.Printf("Ошибка получения курса для %s->%s: %v", fromCurrency, toCurrency, err)
		return nil, fmt.Errorf("ошибка получения курса обмена: %w", err)
	}

	// Проверка реалистичности курса (защита от аномалий)
	if (fromCurrency == "RUB" && toCurrency == "USD" && rate > 0.05) ||
		(fromCurrency == "USD" && toCurrency == "RUB" && rate < 10) {
		return nil, errors.New("нереалистичный курс обмена, проверьте сервис")
	}

	// Логирование параметров операции
	log.Printf("Запрос обмена: %f %s в %s по курсу: %f", amount, fromCurrency, toCurrency, rate)

	// Выполняем обмен валюты в рамках транзакции
	newBalance, err := s.repo.Exchange(ctx, userID, fromCurrency, toCurrency, amount, rate)
	if err != nil {
		return nil, fmt.Errorf("ошибка обмена: %w", err)
	}

	// Дополнительные проверки курса
	maxRates := map[string]float64{
		"RUB/USD": 0.05,
		"USD/RUB": 100,
		"EUR/USD": 2.0,
		"USD/EUR": 2.0,
	}

	key := fmt.Sprintf("%s/%s", fromCurrency, toCurrency)
	if maxRate, ok := maxRates[key]; ok && rate > maxRate {
		return nil, fmt.Errorf("слишком высокий курс обмена: %f", rate)
	}

	// Логирование результата
	log.Printf("Обмен: %s->%s сумма: %.2f, курс: %.6f, результат: %.2f",
		fromCurrency, toCurrency, amount, rate, amount*rate)

	// Формируем ответ
	return &models.ExchangeResponse{
		Message:         "Обмен выполнен успешно",
		ExchangedAmount: amount * rate,
		NewBalance:      newBalance,
		Rate:            rate,
	}, nil
}

// Вспомогательные функции

// isValidCurrency проверяет, поддерживается ли указанная валюта
func isValidCurrency(currency string) bool {
	switch currency {
	case "USD", "RUB", "EUR":
		return true
	default:
		return false
	}
}

// getBalanceByCurrency возвращает баланс по конкретной валюте
func getBalanceByCurrency(balance *models.Balance, currency string) (float64, error) {
	switch currency {
	case "USD":
		return balance.USD, nil
	case "RUB":
		return balance.RUB, nil
	case "EUR":
		return balance.EUR, nil
	default:
		return 0, fmt.Errorf("неподдерживаемая валюта: %s", currency)
	}
}
