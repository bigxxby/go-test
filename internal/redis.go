package gotest

import (
	"fmt"

	"github.com/go-redis/redis"
)

func ConnectToRedis() error {
	// Создаем новый клиент Redis
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Адрес сервера Redis
		Password: "",               // Пароль, если используется аутентификация
		DB:       0,                // Номер базы данных (по умолчанию 0)
	})

	// Проверяем соединение с сервером Redis
	pong, err := client.Ping().Result()
	if err != nil {
		return err
	}
	fmt.Println("Соединение с сервером Redis установлено:", pong)

	// Закрываем соединение с сервером Redis

	defer client.Close()
	return nil
}
