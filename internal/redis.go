package gotest

import (
	"fmt"

	"github.com/go-redis/redis"
)

func ConnectToRedis() error {
	// Создаем новый клиент Redis
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Адрес сервера Redis
		Password: "",           // Пароль, если используется аутентификация
		DB:       0,            // Номер базы данных (по умолчанию 0)
	})

	// Закрываем соединение с сервером Redis в случае ошибки или после использования
	defer client.Close()

	// Проверяем соединение с сервером Redis
	pong, err := client.Ping().Result()
	if err != nil {
		return err
	}
	fmt.Println("Соединение с сервером Redis установлено:", pong)

	return nil
}
