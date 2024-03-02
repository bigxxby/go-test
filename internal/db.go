// db.go
package gotest

import (
	"database/sql"
	"fmt"
	"log"
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
	db     *sql.DB
	dbHost string
	dbPort string
	dbUser string
	dbPass string
	dbName string
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
func InitDB(dbHost, dbPort, dbUser, dbPass, dbName string) (DBHandler, error) {
	db := &SingletonDB{
		dbHost: dbHost,
		dbPort: dbPort,
		dbUser: dbUser,
		dbPass: dbPass,
		dbName: dbName,
	}
	if err := db.Connect(); err != nil {
		return nil, err
	}
	log.Println("Подключение завершено")
	return db, nil
}

func (s *SingletonDB) GetGoods() ([]Good, error) {
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

func (s *SingletonDB) CheckIfGoodExists(id int, projectID int) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM goods WHERE id=$1 AND project_id=$2)"
	var exists bool
	err := s.db.QueryRow(query, id, projectID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
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

	fmt.Println("Data inserted successfully into goods table.")
	return &good, nil
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

	fmt.Println("Data updated successfully in goods table.")
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}
