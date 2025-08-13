package config

import (
	"github.com/joho/godotenv" // Пакет для загрузки .env файлов
	"os"
	"strconv"
	"time"
)

// Config структура содержит все конфигурационные параметры приложения
type Config struct {
	ServerAddress       string        // Адрес и порт HTTP сервера (например: ":8080")
	JWTSecret           string        // Секретный ключ для генерации JWT токенов
	DBHost              string        // Хост PostgreSQL сервера
	DBPort              string        // Порт PostgreSQL сервера
	DBUser              string        // Имя пользователя PostgreSQL
	DBPassword          string        // Пароль пользователя PostgreSQL
	DBName              string        // Имя базы данных
	DBSSLMode           string        // Режим SSL для подключения к БД (disable/require/verify-full)
	ExchangeServiceAddr string        // Адрес gRPC сервиса обмена валют
	TokenExpiration     time.Duration // Время жизни JWT токена (например: "24h")
	CacheTTL            time.Duration // Время жизни кэша в Redis (например: "5m")
	TelegramToken       string        // Токен Telegram бота (если пустой - бот не запускается)
	RedisAddr           string        // Адрес Redis сервера (host:port)
	RedisPassword       string        // Пароль Redis (если требуется)
	RedisDB             int           // Номер базы данных Redis
}

// LoadConfig загружает конфигурацию из .env файла и возвращает структуру Config
// Принимает имя файла конфигурации (например "config2.env")
// Возвращает указатель на Config или ошибку, если загрузка не удалась
func LoadConfig(filename string) (*Config, error) {
	// Загружаем переменные окружения из указанного файла
	err := godotenv.Load("config2.env")
	if err != nil {
		return nil, err // Возвращаем ошибку, если файл не найден или нечитаем
	}

	// Парсим продолжительность жизни токена из переменной окружения
	// По умолчанию 24 часа, если переменная не задана или невалидна
	tokenExp, err := time.ParseDuration(getEnv("TOKEN_EXPIRATION", "24h"))
	if err != nil {
		return nil, err
	}

	// Парсим время жизни кэша из переменной окружения
	// По умолчанию 5 минут, если переменная не задана или невалидна
	cacheTTL, err := time.ParseDuration(getEnv("CACHE_TTL", "5m"))
	if err != nil {
		return nil, err
	}

	// Создаем и возвращаем структуру конфигурации
	// Для каждого параметра используется значение из переменной окружения или значение по умолчанию
	return &Config{
		ServerAddress:       getEnv("SERVER_ADDRESS", ":8080"),                  // Адрес сервера
		JWTSecret:           getEnv("JWT_SECRET", "default-secret"),             // Секрет JWT
		DBHost:              getEnv("DB_HOST", "localhost"),                     // Хост БД
		DBPort:              getEnv("DB_PORT", "5432"),                          // Порт БД
		DBUser:              getEnv("DB_USER", "postgres"),                      // Пользователь БД
		DBPassword:          getEnv("DB_PASSWORD", "ахха не скажу"),             // Пароль БД
		DBName:              getEnv("DB_NAME", "wallet_db"),                     // Имя БД
		DBSSLMode:           getEnv("DB_SSLMODE", "disable"),                    // Режим SSL
		ExchangeServiceAddr: getEnv("EXCHANGE_SERVICE_ADDR", "localhost:50051"), // Адрес сервиса обмена
		TokenExpiration:     tokenExp,                                           // Время жизни токена
		CacheTTL:            cacheTTL,                                           // Время жизни кэша
		TelegramToken:       getEnv("TELEGRAM_TOKEN", ""),                       // Токен бота
		RedisAddr:           getEnv("REDIS_ADDR", "localhost:6379"),             // Адрес Redis
		RedisPassword:       getEnv("REDIS_PASSWORD", ""),                       // Пароль Redis
		RedisDB:             getEnvAsInt("REDIS_DB", 0),                         // Номер БД Redis
	}, nil
}

// GetDBConnString формирует строку подключения к PostgreSQL
// Возвращает строку в формате "host=... port=... user=... password=... dbname=... sslmode=..."
func (c *Config) GetDBConnString() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}

// getEnv вспомогательная функция для получения переменной окружения
// Принимает ключ переменной и значение по умолчанию
// Возвращает значение переменной, если она существует, или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt вспомогательная функция для получения целочисленной переменной окружения
// Принимает имя переменной и значение по умолчанию
// Возвращает значение переменной как int, если она существует и валидна, или значение по умолчанию
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}
