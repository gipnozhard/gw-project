package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gw-currency-wallet/internal/handlers"
	"gw-currency-wallet/internal/middleware"
	"gw-currency-wallet/internal/services"
)

// SetupRouter создает и настраивает маршруты для HTTP-сервера с использованием Gin.
// Параметры:
//   - authService: сервис для аутентификации и регистрации пользователей
//   - walletService: сервис для операций с кошельком (баланс, депозит, снятие)
//   - exchangeService: сервис для работы с курсами валют
//   - jwtSecret: секретный ключ для подписи JWT-токенов
//
// Возвращает:
//   - *gin.Engine: настроенный роутер Gin
func SetupRouter(
	authService *services.AuthService,
	walletService *services.WalletService,
	exchangeService *services.ExchangeService,
	jwtSecret string,
) *gin.Engine {
	router := gin.Default() // Создаем экземпляр Gin с дефолтными middleware (логгирование, восстановление после паники)

	// Настройка Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.DefaultModelsExpandDepth(-1), // Отключаем отображение моделей в Swagger UI
		ginSwagger.PersistAuthorization(true),   // Сохраняем авторизацию между перезагрузками страницы
	))

	// Группа публичных маршрутов (не требуют аутентификации)
	public := router.Group("/api/v1")
	{
		public.POST("/register", handlers.Register(authService)) // Регистрация нового пользователя
		public.POST("/login", handlers.Login(authService))       // Аутентификация пользователя
	}

	// Группа защищенных маршрутов (требуют JWT-аутентификации)
	protected := router.Group("/api/v1")
	protected.Use(middleware.JWTAuthMiddleware(jwtSecret)) // Подключаем middleware для проверки JWT
	{
		// Операции с кошельком
		protected.GET("/balance", handlers.GetBalance(walletService))        // Получение текущего баланса
		protected.POST("/wallet/deposit", handlers.Deposit(walletService))   // Пополнение кошелька
		protected.POST("/wallet/withdraw", handlers.Withdraw(walletService)) // Снятие средств с кошелька

		// Операции с обменом валют
		protected.GET("/exchange/rates", handlers.GetExchangeRates(exchangeService)) // Получение текущих курсов валют
		protected.POST("/exchange", handlers.ExchangeCurrency(walletService))        // Обмен одной валюты на другую
	}

	return router
}
