# vk-internship

[![CI Status](https://github.com/username/vk-internship/workflows/Go/badge.svg)](https://github.com/username/vk-internship/actions)

Этот проект использует непрерывную интеграцию для автоматического выполнения тестов и проверки кода при каждом пул-реквесте.

## Установка

Для установки проекта выполните следующие шаги:

1. Клонируйте репозиторий:
   ```sh
   git clone https://github.com/username/vk-internship.git
   ```
2. Перейдите в директорию проекта:
   ```sh
   cd vk-internship
   ```
3. Установите зависимости:
   ```sh
   go mod download
   ```

## Использование

Для запуска проекта выполните следующую команду:
   ```sh
   go run main.go
   ```

## Тестирование

Для выполнения тестов используйте следующую команду:
   ```sh
   go test ./...
   ```