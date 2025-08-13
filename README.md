# Валютный кошелек и сервис обмена валют (gw-currency-wallet и gw-exchanger)
Обзор проекта
Этот проект состоит из двух микросервисов, взаимодействующих через gRPC:

1. gw-currency-wallet - основной сервис для управления пользовательскими кошельками и операций с валютами

2. gw-exchanger - сервис для получения и хранения курсов валют

## Дополнительные сервисы:

* PostgreSQL - для хранения данных

* Redis - для кэширования

* Telegram бот - для получения курсов валют через Telegram

## Функциональность

### Сервис кошелька (gw-currency-wallet)

* Регистрация и аутентификация пользователей (JWT)

* Управление балансом (пополнение, снятие)

* Обмен валют по текущему курсу

* Получение текущих курсов валют

* Telegram бот для просмотра курсов

### Сервис обмена (gw-exchanger)

* Получение курсов валют от Центрального Банка РФ

* Хранение курсов в PostgreSQL

* Предоставление курсов через gRPC API

* Автоматическое обновление курсов по расписанию

## Технологический стек

* Языки: Go 1.24

* Фреймворки:

  - Gin (HTTP сервер)

  - gRPC (межсервисное взаимодействие)

  - JWT (аутентификация)

* Базы данных:

  - PostgreSQL (основное хранилище)

  - Redis (кэширование)

* Документация: Swagger

* Инфраструктура:

  - Docker

  - Docker Compose

## Запуск проекта

### Требования

* Docker и Docker Compose

* Go 1.24+ (для локальной разработки)

### 1. Запуск через Docker Compose
```bash
make build
make up
```
Сервисы будут доступны:

* Swagger UI: http://localhost:8080/swagger/index.html

* gRPC сервер обмена: localhost:50051

### 2. Локальный запуск (для разработки)
```bash
# Запуск сервиса обмена
cd gw-exchanger
go run cmd/main.go

# Запуск сервиса кошелька (в другом терминале)
cd gw-currency-wallet
go run cmd/main.go
```

## API документация
Документация API доступна через Swagger UI после запуска сервиса:

http://localhost:8080/swagger/index.html

## Основные endpoints:

### Аутентификация

* POST /api/v1/register - регистрация пользователя

  Метод: POST

  URL: /api/v1/register

  Тело запроса:

  ```
  {
    "username": "string",
    "password": "string",
    "email": "string"
  }
  ```
  Ответ:
  
  • Успех: 201 Created

  ```
  { 
    "message": "Пользователь успешно зарегистрирован"
  }
  ````
  • Ошибка: 400 Bad Request

  ```
  {
    "error": "Некорректный запрос"
  }
  ```

  ▎Описание

  Регистрация нового пользователя. Проверяется уникальность имени пользователя и адреса электронной почты. Пароль должен быть зашифрован перед сохранением в базе данных.

--------------------------------------------

* POST /api/v1/login - вход в систему (получение JWT)

  Метод: POST

  URL: /api/v1/login

  Тело запроса:

  ```
  {
  "username": "string",
  "password": "string"
  }
  ```
  
  Ответ:
  
  • Успех: 200 OK

  ```
  {
    "token": "JWT_TOKEN"
  }
  ```
  
  • Ошибка: 401 Unauthorized

  ```
  {
    "error": "Некорректный запрос"
  }
  ```
  
  ▎Описание
  
  
  Авторизация пользователя. При успешной авторизации возвращается JWT-токен, который будет использоваться для аутентификации последующих запросов.

--------------------------------------------

### Кошелек

* GET /api/v1/balance - получение баланса

  Метод: GET
  
  URL: /api/v1/balance
  
  Заголовки:
  
  Authorization: Bearer JWT_TOKEN
  
  Ответ:
  
  • Успех: 200 OK

  ```
  {
    "balance":
    {
    "USD": "float",
    "RUB": "float",
    "EUR": "float"
    }
  }
  ```

--------------------------------------------

* POST /api/v1/wallet/deposit - пополнение счета

  Метод: POST
  
  URL: /api/v1/wallet/deposit

  Заголовки:

  Authorization: Bearer JWT_TOKEN
  
  Тело запроса:

  ```
  {
    "amount": 100.00,
    "currency": "USD" // (USD, RUB, EUR)
  }
  ```
  
  Ответ:
  
  • Успех: 200 OK

  ```
  {
    "message": "Баланс успешно пополнен",
    "new_balance": {
      "USD": "float",
      "RUB": "float",
      "EUR": "float"
    }
  }
  ```
  
  • Ошибка: 400 Bad Request

  ```
  {
  "error": "Некорректный запрос"
  }
  ```
  
  ▎Описание
  
  Позволяет пользователю пополнить свой счет. Проверяется корректность суммы и валюты. Обновляется баланс пользователя в базе данных.

--------------------------------------------

* POST /api/v1/wallet/withdraw - снятие средств

  Метод: POST
  
  URL: /api/v1/wallet/withdraw

  Заголовки:

  Authorization: Bearer JWT_TOKEN
  
  Тело запроса:

  ```
  {
      "amount": 50.00,
      "currency": "USD" // USD, RUB, EUR)
  }
  ```
  
  Ответ:
  
  • Успех: 200 OK

  ```
  {
    "message": "Средства успешно сняты",
    "new_balance": {
      "USD": "float",
      "RUB": "float",
      "EUR": "float"
    }
  }
  ```
  
  • Ошибка: 400 Bad Request

  ```
  {
    "error": "Некорректный запрос"
  }
  ```
  
  ▎Описание
  
  Позволяет пользователю вывести средства со своего счета. Проверяется наличие достаточного количества средств и корректность суммы.

--------------------------------------------

### Обмен валют

* GET /api/v1/exchange/rates - получение текущих курсов

Метод: GET

URL: /api/v1/exchange/rates

Заголовки:

Authorization: Bearer JWT_TOKEN

Ответ:

• Успех: 200 OK

```
{
    "rates": 
    {
      "USD": "float",
      "RUB": "float",
      "EUR": "float"
    }
}
```

• Ошибка: 500 Internal Server Error

```
{
  "error": "Сервис обмена не инициализирован",
  "message": "Сервис обмена недоступен",
}
```

▎Описание

Получение актуальных курсов валют из внешнего gRPC-сервиса. Возвращает курсы всех поддерживаемых валют.

--------------------------------------------

* POST /api/v1/exchange - обмен валюты

Метод: POST

URL: /api/v1/exchange

Заголовки:

Authorization: Bearer JWT_TOKEN

Тело запроса:

```
{
  "from_currency": "USD",
  "to_currency": "EUR",
  "amount": 100.00
}
```

Ответ:

• Успех: 200 OK

```
{
  "message": "Обмен выполнен успешно",
  "exchanged_amount": 85.00,
  "new_balance":
  {
  "USD": 0.00,
  "EUR": 85.00
  }
}
```

• Ошибка: 400 Bad Request

```
{
  "error": "Некорректный запрос"
}
```

▎Описание

Курс валют осуществляется по данным сервиса exchange (если в течении небольшого времени был запрос от клиента курса валют (/api/v1/exchange) до обмена, то брать курс из кэша, если же запроса курса валют не было или он запрашивался слишком давно, то нужно осуществить gRPC-вызов к внешнему сервису, который предоставляет актуальные курсы валют) Проверяется наличие средств для обмена, и обновляется баланс пользователя.

--------------------------------------------

## Конфигурация

Основные настройки задаются через переменные окружения:

### Сервис кошелька (gw-currency-wallet/config2.env)

```ini
SERVER_ADDRESS=:8080
JWT_SECRET=your-very-secret-key
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD="свой пароль ставить"
DB_NAME=wallet_db
EXCHANGE_SERVICE_ADDR=exchanger:50051
REDIS_ADDR=redis:6379
````

### Сервис обмена (gw-exchanger/config.env)

```ini
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD="свой пароль ставить"
DB_NAME=exchange_rates
CB_API_URL=https://www.cbr-xml-daily.ru/daily_json.js
UPDATE_INTERVAL_MINUTES=60
```
### Курсы валют получем с API ЦБ:

https://www.cbr-xml-daily.ru/daily_json.js

## Архитектура

### Сервис кошелька

```text
┌───────────────────────┐       ┌───────────────────────┐
│                       │       │                       │
│   HTTP API (Gin)      │──────▶│    Business Logic     │
│                       │       │      (Services)       │
└──────────┬────────────┘       └──────────┬────────────┘
           │                               │
           │                               │
           ▼                               ▼
┌───────────────────────┐       ┌───────────────────────┐
│     JWT Middleware    │       │       Storage         │
│                       │       │    (PostgreSQL)       │
└───────────────────────┘       └──────────┬────────────┘
                                           │
                                           │ gRPC
                                           ▼
                                 ┌───────────────────────┐
                                 │      gw-exchanger     │
                                 └───────────────────────┘
```


### Сервис обмена

```text
┌───────────────────────┐       ┌───────────────────────┐
│                       │       │                       │
│   gRPC Server         │──────▶│    Rate Updater       │
│                       │       │                       │
└──────────┬────────────┘       └──────────┬────────────┘
           │                               │
           │                               │
           ▼                               ▼
┌───────────────────────┐       ┌───────────────────────┐
│     Rate Provider     │       │     API Client        │
│    (PostgreSQL)       │       │   (ЦБ РФ API)         │
└───────────────────────┘       └───────────────────────┘
```

## Структура всего проекта

```
gw-project
├── docker-compose.yml
├── go.work
├── go.work.sum
├── gw-currency-wallet
│   ├── cmd
│   │   └── main.go
│   ├── config2.env
│   ├── Dockerfile
│   ├── docs
│   │   ├── docs.go
│   │   ├── swagger.json
│   │   └── swagger.yaml
│   ├── go.mod
│   ├── go.sum
│   ├── internal
│   │   ├── config
│   │   │   └── config.go
│   │   ├── handlers
│   │   │   ├── auth_handlers.go
│   │   │   └── wallet_handler.go
│   │   ├── middleware
│   │   │   └── auth.go
│   │   ├── models
│   │   │   └── user.go
│   │   ├── services
│   │   │   ├── auth_service.go
│   │   │   ├── exchange_service.go
│   │   │   └── wallet_service.go
│   │   ├── storage
│   │   │   ├── postgres
│   │   │   │   └── connector.go
│   │   │   ├── redis
│   │   │   │   └── client.go
│   │   │   └── storage.go
│   │   └── telegram
│   │       ├── bot.go
│   │       ├── currency_flags.go
│   │       ├── handler.go
│   │       └── service.go
│   └── routes
│       └── router.go
├── gw-exchanger
│   ├── cmd
│   │   └── main.go
│   ├── config.env
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── internal
│   │   ├── api
│   │   │   └── cbr.go
│   │   ├── server
│   │   │   └── server.go
│   │   ├── storage
│   │   │   ├── model.go
│   │   │   ├── postgres
│   │   │   │   ├── connector.go
│   │   │   │   └── methods.go
│   │   │   └── storage.go
│   │   └── utils
│   │       └── currency_printer.go
│   └── migrations
│       └── 001_init.sql
├── gw-proto
│   ├── go.mod
│   ├── go.sum
│   └── proto
│       ├── exchange_grpc.pb.go
│       ├── exchange.pb.go
│       └── exchange.proto
├── init.sql
└── Makefile
```

### Telgram BOT

Можно создать свой код под него подйдет , только вписать токен, пример вывода:

```
Текущие курсы валют в рублях:

Лей 🇷🇴 RON: 18.2822
Песо 🇨🇺 CUP: 3.3280
Найра 🇳🇬 NGN: 0.0520
Быр 🇪🇹 ETB: 0.5707
Сум 🇺🇿 UZS: 0.0063
Фунт 🇪🇬 EGP: 1.6490
Риал 🇶🇦 QAR: 21.9427
Манат 🇹🇲 TMT: 22.8204
Юань 🇨🇳 CNY: 11.0609
Франк 🇨🇭 CHF: 98.6067
Лира 🇹🇷 TRY: 1.9639
Реал 🇧🇷 BRL: 14.6634
Риал 🇮🇷 IRR: 0.0001
Рэнд 🇿🇦 ZAR: 4.5117
Крона 🇩🇰 DKK: 12.4381
Рубль 🇷🇺 RUB: 1.0000
Канадский доллар 🇨🇦 CAD: 57.9408
Тугрик 🇲🇳 MNT: 0.0222
Динар 🇧🇭 BHD: 212.3781
Драм 🇦🇲 AMD: 0.2081
Гонконгский доллар 🇭🇰 HKD: 10.1929
Крона 🇨🇿 CZK: 3.7879
Бат 🇹🇭 THB: 2.4705
Тенге 🇰🇿 KZT: 0.1472
Лев 🇧🇬 BGN: 47.4614
Боливиано 🇧🇴 BOB: 11.5588
Донг 🇻🇳 VND: 0.0032
Риял 🇸🇦 SAR: 21.2990
Крона 🇸🇪 SEK: 8.3014
Злотый 🇵🇱 PLN: 21.7621
Евро 🇪🇺 EUR: 92.8583
Динар 🇷🇸 RSD: 0.7926
Манат 🇦🇿 AZN: 46.9832
Така 🇧🇩 BDT: 0.6560
Иена 🇯🇵 JPY: 0.5382
СДР 🌐 XDR: 109.1163
Лей 🇲🇩 MDL: 4.7510
Рупия 🇮🇩 IDR: 0.0049
Сомони 🇹🇯 TJS: 8.5314
Рупия 🇮🇳 INR: 0.9110
Сом 🇰🇬 KGS: 0.9144
Форинт 🇭🇺 HUF: 0.2343
Рубль 🇧🇾 BYN: 26.8214
Вона 🇰🇷 KRW: 0.0575
Крона 🇳🇴 NOK: 7.7973
Новозеландский доллар 🇳🇿 NZD: 47.4636
Кьят 🇲🇲 MMK: 0.0380
Динар 🇩🇿 DZD: 0.6145
Дирхам 🇦🇪 AED: 21.7485
Фунт стерлингов 🇬🇧 GBP: 107.0916
Австралийский доллар 🇦🇺 AUD: 51.9004
Сингапурский доллар 🇸🇬 SGD: 62.0987
Гривна 🇺🇦 UAH: 1.9280
Риал 🇴🇲 OMR: 207.7280
Лари 🇬🇪 GEL: 29.6391
Доллар США 🇺🇸 USD: 79.8714
```
