package main

import (
	"fmt"
	gotest "gotest/internal"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var (
	dbHost     = os.Getenv("POSTGRES_HOST")
	dbPort     = os.Getenv("POSTGRES_PORT")
	dbUser     = os.Getenv("POSTGRES_USER")
	dbName     = os.Getenv("POSTGRES_DB")
	dbPassword = os.Getenv("POSTGRES_PASSWORD")
	redisHost  = os.Getenv("REDIS_HOST")
)

func main() {
	// Подключение к Redis
	var err error
	for i := 0; i < 10; i++ {
		err = gotest.ConnectToRedis()
		if err == nil {
			log.Println("Error connecting to redis: ", err)
			break
		}
		fmt.Printf("Ошибка при соединении с сервером Redis: %v. Повторная попытка через 1 секунду\n", err)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		fmt.Println("Невозможно подключиться к серверу Redis:", err)
		return
	}
	addr := redisHost + ":6379"
	// Подключение к базе данных
	db, err := gotest.InitDB(dbHost, dbPort, dbUser, dbPassword, dbName, addr, "")
	if err != nil {
		panic(err)
	}
	handler := gotest.NewHandler(db)
	err = db.CreateProjectsTable()
	err = db.CreateGoodsTable()
	err = db.CreateIndex()
	if err != nil {
		panic(err)
	}
	// Регистрируем хендлеры
	http.HandleFunc("/", handler.Main)
	http.HandleFunc("/good/get", handler.GET)
	http.HandleFunc("/good/create", handler.POST)
	http.HandleFunc("/good/update", handler.PATCH)
	http.HandleFunc("/good/remove", handler.DELETE)
	log.Println("Started - http://localhost:8080/")
	// Запускаем сервер
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
