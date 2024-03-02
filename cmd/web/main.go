package main

import (
	"fmt"
	gotest "gotest/internal"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

// var (
//
//	dbHost     = os.Getenv("POSTGRES_HOST")
//	dbPort     = os.Getenv("POSTGRES_PORT")
//	dbUser     = os.Getenv("POSTGRES_USER")
//	dbName     = os.Getenv("POSTGRES_DB")
//	dbPassword = os.Getenv("POSTGRES_PASSWORD")
//
// )
var (
	dbHost     = "localhost "
	dbPort     = "5432"
	dbUser     = "postgres"
	dbName     = "postgres"
	dbPassword = "postgres"
)

func main() {
	err := gotest.ConnectToRedis()
	if err != nil {
		fmt.Println("Ошибка при соединении с сервером Redis:", err)
		return
	}

	db, err := gotest.InitDB(dbHost, dbPort, dbUser, dbPassword, dbName, "127.0.0.1:6379", "")
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
