package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"     // Веб-фреймворк Gin
	"github.com/golang-jwt/jwt/v5" // JWT реализация
	"net/http"
	"strings"
	"time"
)

// JWTClaims - кастомная структура claims для JWT токена
// Содержит ID пользователя и стандартные зарегистрированные claims
type JWTClaims struct {
	UserID               int `json:"user_id"` // ID пользователя - основная информация в токене
	jwt.RegisteredClaims     // Стандартные claims (exp, iat и др.)
}

// JWTAuthMiddleware - middleware для JWT аутентификации
// Принимает секретный ключ для верификации токенов
// Возвращает Gin-обработчик, который:
// 1. Проверяет наличие и формат токена
// 2. Валидирует подпись и срок действия
// 3. Добавляет userID в контекст при успешной аутентификации
func JWTAuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Извлечение токена из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Требуется заголовок Authorization",
			})
			return
		}

		// 2. Проверка формата: "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Неверный формат заголовка Authorization",
			})
			return
		}

		tokenString := tokenParts[1] // Сам токен без префикса

		// 3. Парсинг и валидация токена
		claims, err := parseToken(tokenString, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Неверный токен: " + err.Error(),
			})
			return
		}

		// 4. Проверка срока действия токена
		if time.Now().After(claims.ExpiresAt.Time) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Срок действия токена истек",
			})
			return
		}

		// 5. Успешная аутентификация - добавляем userID в контекст
		c.Set("userID", claims.UserID)

		// Передаем управление следующему обработчику
		c.Next()
	}
}

// parseToken - внутренняя функция для парсинга и валидации JWT токена
// Принимает:
// - tokenString: строка с JWT токеном
// - secret: секретный ключ для проверки подписи
// Возвращает:
// - *JWTClaims: распарсенные claims при успехе
// - error: ошибку при неудачной проверке
func parseToken(tokenString, secret string) (*JWTClaims, error) {
	// Парсим токен с указанием структуры для claims
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Проверяем, что используется ожидаемый алгоритм подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("неожиданный метод подписи")
			}
			return []byte(secret), nil // Возвращаем ключ для верификации
		},
	)

	if err != nil {
		return nil, err
	}

	// Проверяем валидность claims
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("неверные данные токена")
}

// GenerateJWTToken - генерирует новый JWT токен
// Принимает:
// - userID: идентификатор пользователя
// - secret: секретный ключ для подписи
// - expiration: время жизни токена
// Возвращает:
// - string: подписанный токен
// - error: ошибку при генерации
func GenerateJWTToken(userID int, secret string, expiration time.Duration) (string, error) {
	// Создаем claims с userID и временем expiration
	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
		},
	}

	// Создаем токен с алгоритмом HS256 и claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен секретным ключом
	return token.SignedString([]byte(secret))
}
