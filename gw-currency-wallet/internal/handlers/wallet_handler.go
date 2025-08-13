package handlers

import (
	"github.com/gin-gonic/gin"
	"gw-currency-wallet/internal/models"
	"gw-currency-wallet/internal/services"
	"log"
	"net/http"
)

// GetBalance godoc
// @Summary Получить баланс
// @Description Возвращает баланс пользователя по всем валютам
// @Tags Wallet
// @Security BearerAuth - Требуется JWT токен
// @Produce json
// @Success 200 {object} models.Balance - Успешный ответ с балансом
// @Failure 401 {object} models.ErrorResponse - Ошибка аутентификации
// @Failure 500 {object} models.ErrorResponse - Внутренняя ошибка сервера
// @Router /balance [get] - GET endpoint
func GetBalance(walletService *services.WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Извлекаем userID из контекста (устанавливается middleware аутентификации)
		userID := c.MustGet("userID").(int)

		// Получаем баланс через сервисный слой
		balance, err := walletService.GetBalance(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения баланса"})
			return
		}

		// Возвращаем баланс в формате JSON
		c.JSON(http.StatusOK, gin.H{"balance": balance})
	}
}

// Deposit godoc
// @Summary Пополнить баланс
// @Description Пополнение баланса пользователя в указанной валюте
// @Tags Wallet
// @Security BearerAuth
// @Accept json - Ожидаем JSON в теле запроса
// @Produce json
// @Param input body models.DepositRequest true "Данные для пополнения"
// @Success 200 {object} models.TransactionResponse - Ответ с новым балансом
// @Failure 400 {object} models.ErrorResponse - Некорректный запрос
// @Failure 401 {object} models.ErrorResponse - Ошибка аутентификации
// @Failure 500 {object} models.ErrorResponse - Ошибка сервера
// @Router /wallet/deposit [post] - POST endpoint
func Deposit(walletService *services.WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Структура для парсинга входящего запроса
		var request struct {
			Amount   float64 `json:"amount"`   // Сумма пополнения
			Currency string  `json:"currency"` // Код валюты (USD, EUR и т.д.)
		}

		// Парсим JSON тело запроса
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный запрос"})
			return
		}

		// Извлекаем userID из контекста
		userID := c.MustGet("userID").(int)

		// Вызываем сервис для пополнения баланса
		newBalance, err := walletService.Deposit(
			c.Request.Context(),
			userID,
			request.Currency,
			request.Amount,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Возвращаем успешный ответ с новым балансом
		c.JSON(http.StatusOK, gin.H{
			"message":     "Баланс успешно пополнен",
			"new_balance": newBalance,
		})
	}
}

// Withdraw godoc
// @Summary Снять средства
// @Description Снятие средств с баланса пользователя
// @Tags Wallet
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body models.WithdrawRequest true "Данные для снятия"
// @Success 200 {object} models.TransactionResponse
// @Failure 400 {object} models.ErrorResponse - Недостаточно средств/некорректная валюта
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /wallet/withdraw [post]
func Withdraw(walletService *services.WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Используем модель WithdrawRequest для валидации
		var request models.WithdrawRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный запрос"})
			return
		}

		userID := c.MustGet("userID").(int)

		// Вызываем сервис для снятия средств
		newBalance, err := walletService.Withdraw(
			c.Request.Context(),
			userID,
			request.Currency,
			request.Amount,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "Средства успешно сняты",
			"new_balance": newBalance,
		})
	}
}

// GetExchangeRates godoc
// @Summary Получить курсы валют
// @Description Возвращает текущие курсы обмена валют
// @Tags Exchange
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.ExchangeRatesResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse - Сервис обмена недоступен
// @Router /exchange/rates [get]
func GetExchangeRates(exchangeService *services.ExchangeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверка инициализации сервиса
		if exchangeService == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Сервис обмена не инициализирован",
				"message": "Сервис обмена недоступен",
			})
			return
		}

		// Получаем текущие курсы валют
		rates, err := exchangeService.GetRates(c.Request.Context())
		if err != nil {
			log.Printf("Ошибка получения курсов валют: %v", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Ошибка получения курсов валют",
				"message": "Сервис обмена недоступен",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"rates": rates})
	}
}

// ExchangeCurrency godoc
// @Summary Обмен валют
// @Description Обменивает указанную сумму из одной валюты в другую по текущему курсу
// @Tags Exchange
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body models.ExchangeRequest true "Данные для обмена"
// @Success 200 {object} models.ExchangeResponse - Результат обмена
// @Failure 400 {object} models.ErrorResponse - Недостаточно средств/некорректные данные
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /exchange [post]
func ExchangeCurrency(
	walletService *services.WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request models.ExchangeRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный запрос"})
			return
		}

		userID := c.MustGet("userID").(int)

		// Выполняем обмен через сервисный слой
		response, err := walletService.Exchange(
			c.Request.Context(),
			userID,
			request.FromCurrency,
			request.ToCurrency,
			request.Amount,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}
