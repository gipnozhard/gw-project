package telegram

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Handler представляет обработчик Telegram-бота, который управляет входящими командами
// и взаимодействует с сервисом для получения курсов валют.
type Handler struct {
	bot             *tgbotapi.BotAPI // Клиент Telegram Bot API для отправки сообщений
	exchangeService *ExchangeService // Сервис для работы с курсами валют
}

// NewHandler создает новый экземпляр Handler с заданными зависимостями.
// Параметры:
//   - bot: клиент Telegram Bot API
//   - exchangeService: сервис для работы с курсами валют
//
// Возвращает:
//   - Указатель на созданный Handler
func NewHandler(bot *tgbotapi.BotAPI, exchangeService *ExchangeService) *Handler {
	return &Handler{
		bot:             bot,
		exchangeService: exchangeService,
	}
}

// HandleCommand обрабатывает входящую команду от пользователя и отправляет соответствующий ответ.
// В зависимости от команды (например, "/start" или "/rates"), формируется ответное сообщение.
// Параметры:
//   - msg: входящее сообщение от пользователя
func (h *Handler) HandleCommand(msg *tgbotapi.Message) {
	// Создаем новое сообщение для ответа в тот же чат
	response := tgbotapi.NewMessage(msg.Chat.ID, "")

	// Обрабатываем команду из сообщения
	switch msg.Command() {
	case "start":
		// Ответ на команду /start
		response.Text = "Привет! Я бот для отслеживания курсов валют. Используй команду /rates чтобы получить текущие курсы."

	case "rates":
		// Ответ на команду /rates: получение и отображение текущих курсов валют
		rates, err := h.exchangeService.GetAllRates()
		if err != nil {
			response.Text = "Не удалось получить курсы валют. Попробуйте позже."
			break
		}

		// Формируем строку с курсами валют
		var sb strings.Builder
		sb.WriteString("Текущие курсы валют в рублях:\n\n")

		// Добавляем каждую валюту и ее курс в ответ
		for currency, rate := range rates {
			flag := GetCurrencyFlag(currency) // Получаем флаг для валюты (например, 🇺🇸 для USD)
			sb.WriteString(fmt.Sprintf("%s %s: %.4f\n", flag, currency, rate))
		}

		response.Text = sb.String()

	default:
		// Ответ на неизвестную команду
		response.Text = "Я не знаю такой команды. Доступные команды: /start, /rates"
	}

	// Отправляем ответ пользователю
	if _, err := h.bot.Send(response); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err) // Логируем ошибку, если отправка не удалась
	}
}
