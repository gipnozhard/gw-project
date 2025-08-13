package handlers

import (
	"github.com/gin-gonic/gin" // Веб-фреймворк Gin
	"gw-currency-wallet/internal/models"
	"gw-currency-wallet/internal/services"
	"net/http"
)

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя в системе
// @Tags Auth - Группа методов в Swagger
// @Accept json - Ожидаемый Content-Type
// @Produce json - Возвращаемый Content-Type
// @Param input body models.CreateUserRequest true "Данные для регистрации"
// @Success 201 {object} models.SuccessMessage - Успешный ответ
// @Failure 400 {object} models.ErrorResponse - Ошибка валидации
// @Router /register [post] - Путь и HTTP метод
func Register(authService *services.AuthService) gin.HandlerFunc {
	// Возвращаем функцию-обработчик Gin
	return func(c *gin.Context) {
		// 1. Парсинг входных данных
		var req models.CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			// При ошибке парсинга возвращаем 400 Bad Request
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный запрос"})
			return
		}

		// 2. Вызов сервиса регистрации
		user, err := authService.Register(c.Request.Context(), req)
		if err != nil {
			// При ошибке регистрации возвращаем 400 с описанием ошибки
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 3. Успешный ответ
		c.JSON(http.StatusCreated, gin.H{
			"message": "Пользователь успешно зарегистрирован",
			"user_id": user.ID, // Возвращаем ID созданного пользователя
		})
	}
}

// Login godoc
// @Summary Аутентификация пользователя
// @Description Вход в систему с получением JWT токена
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body models.LoginRequest true "Данные для входа"
// @Success 200 {object} models.LoginResponse - Успешный ответ с токеном
// @Failure 401 {object} models.ErrorResponse - Ошибка аутентификации
// @Router /login [post]
func Login(authService *services.AuthService) gin.HandlerFunc {
	// Возвращаем функцию-обработчик Gin
	return func(c *gin.Context) {
		// 1. Парсинг входных данных
		var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			// При ошибке парсинга возвращаем 400
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный запрос"})
			return
		}

		// 2. Вызов сервиса аутентификации
		token, err := authService.Login(c.Request.Context(), req.Username, req.Password)
		if err != nil {
			// При ошибке аутентификации возвращаем 401 Unauthorized
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// 3. Успешный ответ с JWT токеном
		c.JSON(http.StatusOK, gin.H{
			"token": token, // Возвращаем сгенерированный токен
		})
	}
}
