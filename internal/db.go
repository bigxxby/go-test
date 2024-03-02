// db.go
package gotest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
)

// DBHandler - интерфейс для работы с базой данных.
type DBHandler interface {
	Connect() error
	Close()
	CreateProjectsTable() error // Добавляем метод для создания таблицы projects
	CreateGoodsTable() error
	CreateIndex() error // Добавляем метод для создания таблицы projects
	GetGoods() ([]Good, error)
	GetProjects() ([]Project, error)
	CheckIfProjectExists(id int) (bool, error)
	CheckIfGoodExists(id int, projectID int) (bool, error)
	CreateGoods(projectId int, name string) (*Good, error)
	UpdateGoods(projectID int, id int, name string, description string) (*Good, error)
	DeleteGoods(projectID int, id int) error
}

// SingletonDB - структура, реализующая интерфейс DBHandler.
type SingletonDB struct {
	db          *sql.DB
	redisClient *redis.Client
	dbHost      string
	dbPort      string
	dbUser      string
	dbPass      string
	dbName      string
}

// Connect - метод для подключения к базе данных.
func (s *SingletonDB) Connect() error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		s.dbHost, s.dbPort, s.dbUser, s.dbPass, s.dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("Ошибка при подключении к базе данных: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("Ошибка при проверке соединения с базой данных: %v", err)
	}

	s.db = db
	log.Println("Соединение с базой данных успешно установлено")
	return nil
}

// Close - метод для закрытия соединения с базой данных.
func (s *SingletonDB) Close() {
	if s.db != nil {
		s.db.Close()
		log.Println("Соединение с базой данных закрыто")
	}
}

// CreateProjectsTable - метод для создания таблицы projects.
func (s *SingletonDB) CreateProjectsTable() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id SERIAL PRIMARY KEY,
			name TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		);

		INSERT INTO projects ( name, created_at) VALUES ( 'john', NOW());
	`)
	if err != nil {
		return fmt.Errorf("Ошибка при создании таблицы projects: %v", err)
	}
	log.Println("Таблица projects успешно создана")
	return nil
}

func (s *SingletonDB) CreateGoodsTable() error {
	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS goods (
		id SERIAL PRIMARY KEY,
		project_id INTEGER REFERENCES projects(id),
		name VARCHAR(255),
		description TEXT DEFAULT '',
		priority INTEGER DEFAULT 1,
		removed BOOLEAN DEFAULT false,
		created_at TIMESTAMP DEFAULT NOW()
	);
	`)
	if err != nil {
		return fmt.Errorf("Ошибка при создании таблицы projects: %v", err)
	}
	log.Println("Таблица goods успешно создана")
	return nil
}

func (s *SingletonDB) CheckIfProjectExists(id int) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM projects WHERE id=$1)"
	var exists bool
	err := s.db.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *SingletonDB) CreateIndex() error {
	_, err := s.db.Exec(`
	CREATE INDEX IF NOT EXISTS goods_id_index ON goods (id);
	CREATE INDEX IF NOT EXISTS goods_project_id_index ON goods (project_id);
	CREATE INDEX IF NOT EXISTS projects_name_index ON projects (name);
	`)
	if err != nil {
		return fmt.Errorf("Ошибка при создании индексов: %v", err)
	}
	return nil
}

// InitDB - функция для инициализации подключения к базе данных.
func InitDB(dbHost, dbPort, dbUser, dbPass, dbName, redisAddr, redisPass string) (DBHandler, error) {
	db := &SingletonDB{
		dbHost: dbHost,
		dbPort: dbPort,
		dbUser: dbUser,
		dbPass: dbPass,
		dbName: dbName,
	}

	// Инициализация клиента Redis
	db.redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       0,
	})

	if err := db.Connect(); err != nil {
		return nil, err
	}
	log.Println("Подключение c redis завершено")

	return db, nil
}
func (s *SingletonDB) CheckIfGoodExists(id int, projectID int) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM goods WHERE id=$1 AND project_id=$2)"
	var exists bool
	err := s.db.QueryRow(query, id, projectID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
func (s *SingletonDB) GetGoods() ([]Good, error) {
	// Используем контекст по умолчанию

	// Проверяем наличие данных в кеше Redis
	// Проверяем наличие данных в кеше Redis
	goodsJSON, err := s.redisClient.Get("goods").Result()

	if err == redis.Nil {
		// Если ключ отсутствует в кеше, получаем данные из базы данных
		goodsFromDB, err := s.fetchGoodsFromDB()
		if err != nil {
			return nil, err
		}

		// Сохраняем данные в кеш Redis
		goodsJSON, err := json.Marshal(goodsFromDB)
		if err != nil {
			return nil, err
		}
		err = s.redisClient.Set("goods", goodsJSON, 10*time.Minute).Err()
		if err != nil {
			return nil, err
		}

		return goodsFromDB, nil
	} else if err != nil {
		// Обработка ошибки при работе с кешем Redis
		return nil, err
	}

	// Декодируем данные из JSON обратно в структуру Good
	var goods []Good
	err = json.Unmarshal([]byte(goodsJSON), &goods)
	if err != nil {
		return nil, err
	}

	return goods, nil
}

func (s *SingletonDB) fetchGoodsFromDB() ([]Good, error) {
	rows, err := s.db.Query("SELECT id, project_id, name, description, priority, removed, created_at FROM goods")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goods []Good
	for rows.Next() {
		var good Good

		err := rows.Scan(&good.ID, &good.ProjectID, &good.Name, &good.Description, &good.Priority, &good.Removed, &good.CreatedAt)
		if err != nil {
			return nil, err
		}
		goods = append(goods, good)
	}
	return goods, nil
}

func (s *SingletonDB) GetProjects() ([]Project, error) {
	rows, err := s.db.Query("SELECT id, name, created_at FROM projects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		err := rows.Scan(&project.ID, &project.Name, &project.CreatedAt)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, nil
}

func (s *SingletonDB) CreateGoods(projectID int, name string) (*Good, error) {
	// Начинаем транзакцию
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %v", err)
	}

	query := "INSERT INTO goods (project_id, name) VALUES ($1, $2) RETURNING id, project_id, name, description, priority, removed, created_at"
	var good Good
	err = tx.QueryRow(query, projectID, name).Scan(&good.ID, &good.ProjectID, &good.Name, &good.Description, &good.Priority, &good.Removed, &good.CreatedAt)
	if err != nil {
		// Если произошла ошибка при выполнении запроса, откатываем транзакцию и возвращаем ошибку
		tx.Rollback()
		return nil, fmt.Errorf("error inserting goods: %v", err)
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		// Если произошла ошибка при коммите транзакции, возвращаем ошибку
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	// Обновляем данные в Redis после успешного добавления товара
	err = s.updateGoodsCache()
	if err != nil {
		fmt.Println("Error updating goods cache:", err)
	}

	fmt.Println("Data inserted successfully into goods table.")
	return &good, nil
}

func (s *SingletonDB) updateGoodsCache() error {
	// Получаем все товары из базы данных
	goods, err := s.fetchGoodsFromDB()
	if err != nil {
		return fmt.Errorf("error fetching goods from database: %v", err)
	}

	// Преобразуем данные в формат JSON
	goodsJSON, err := json.Marshal(goods)
	if err != nil {
		return fmt.Errorf("error marshaling goods to JSON: %v", err)
	}

	// Обновляем данные в кеше Redis
	err = s.redisClient.Set("goods", goodsJSON, 10*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("error updating goods cache: %v", err)
	}

	return nil
}

func (s *SingletonDB) UpdateGoods(projectID int, id int, name string, description string) (*Good, error) {
	// Начинаем транзакцию
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %v", err)
	}

	query := "UPDATE goods SET name = $1, description = $2, priority = priority + 1 WHERE id = $3 AND project_id = $4 RETURNING id, project_id, name, description, priority, removed, created_at"
	var good Good
	err = tx.QueryRow(query, name, description, id, projectID).Scan(&good.ID, &good.ProjectID, &good.Name, &good.Description, &good.Priority, &good.Removed, &good.CreatedAt)
	if err != nil {
		// Если произошла ошибка при выполнении запроса, откатываем транзакцию и возвращаем ошибку
		tx.Rollback()
		return nil, fmt.Errorf("error updating goods: %v", err)
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		// Если произошла ошибка при коммите транзакции, возвращаем ошибку
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	// Обновляем данные в Redis
	updatedGoodJSON, err := json.Marshal(good)
	if err != nil {
		return nil, err
	}
	err = s.redisClient.Set(fmt.Sprintf("good:%d", good.ID), updatedGoodJSON, 10*time.Minute).Err()
	if err != nil {
		return nil, err
	}

	err = s.updateGoodsCache()
	if err != nil {
		fmt.Println("Error updating goods cache:", err)
	}
	fmt.Println("Data updated successfully in goods table and Redis.")
	return &good, nil
}

func (s *SingletonDB) DeleteGoods(projectID int, id int) error {
	// Начинаем транзакцию
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}

	query := "DELETE FROM goods WHERE project_id = $1 AND id = $2"
	_, err = tx.Exec(query, projectID, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting goods: %v", err)
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	// Удаляем данные из Redis
	key := fmt.Sprintf("good:%d", id)
	err = s.redisClient.Del(key).Err()
	if err != nil {
		return fmt.Errorf("error deleting data from Redis: %v", err)
	}
	err = s.updateGoodsCache()
	if err != nil {
		fmt.Println("Error updating goods cache:", err)
	}
	fmt.Println("Data deleted successfully from goods table and Redis.")
	return nil
}
