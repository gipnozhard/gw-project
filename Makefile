.PHONY: build up down migrate test

# Сборка всех сервисов
build:
	docker-compose build

# Запуск всей системы
up:
	docker-compose up -d

# Остановка системы
down:
	docker-compose down

# Перезапуск системы
restart:
	down up

# Запуск миграций
migrate:
	docker-compose run --rm wallet ./wallet migrate

# Запуск тестов
test:
	docker-compose run --rm wallet go test ./...

# Просмотр логов wallet
logs-wallet:
	docker-compose logs -f wallet

# Просмотр логов в ообщем
logs:
	docker-compose logs -f

# Просмотр логов exchanger
logs-exchanger:
	docker-compose logs -f exchanger

# Очистка (удаление контейнеров, образов и томов)
clean:
	docker-compose down -v --rmi all