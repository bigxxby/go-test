// handlers.go
package gotest

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

// Создаем структуру хендлера с полем db типа DBHandler
type Handler struct {
	db DBHandler
}

// Изменяем конструктор для хендлера, чтобы он принимал объект базы данных и клиент Redis
func NewHandler(db DBHandler) *Handler {
	return &Handler{
		db: db,
	}
}

type Good struct {
	ID          int    `json:"id"`
	ProjectID   int    `json:"project_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	Removed     bool   `json:"removed"`
	CreatedAt   string `json:"created_at"`
}

type Project struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func (h *Handler) Main(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	temp, err := template.ParseFiles("ui/templates/main.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", 500)
		return
	}
	err = temp.Execute(w, nil)
	if err != nil {

		http.Error(w, "Internal server error", 500)
		return
	}
}

var JsonData struct {
	Id          string `json:"id"`
	ProjectID   string `json:"projectId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int
}

var JsonDataDeleted struct {
	Id        string `json:"id"`
	ProjectID string `json:"projectId"`
	Removed   bool
}

func (h *Handler) POST(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&JsonData); err != nil {
		log.Println(err.Error())
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}
	if JsonData.ProjectID == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	projectId := JsonData.ProjectID
	idNum, err := strconv.Atoi(projectId)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	name := JsonData.Name
	if name == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	booelan, err := h.db.CheckIfProjectExists(idNum)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", 500)

		return
	}
	if !booelan {
		http.NotFound(w, r)
		return
	}
	good, err := h.db.CreateGoods(idNum, name)
	responseJSON, err := json.Marshal(good)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Установка заголовка Content-Type на application/json
	w.Header().Set("Content-Type", "application/json")
	// Отправка ответа JSON
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

func (h *Handler) PATCH(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&JsonData); err != nil {
		log.Println(err.Error())
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}
	if JsonData.Name == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	projectIdNum, err := strconv.Atoi(JsonData.ProjectID)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	idNum, err := strconv.Atoi(JsonData.Id)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	boolean, err := h.db.CheckIfGoodExists(idNum, projectIdNum)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", 500)

		return
	}
	if !boolean {
		http.NotFound(w, r)
		return
	}
	goods, err := h.db.UpdateGoods(projectIdNum, idNum, JsonData.Name, JsonData.Description)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", 500)
		return
	}
	JsonData.Priority = goods.Priority

	responseJSON, err := json.Marshal(JsonData)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Установка заголовка Content-Type на application/json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

////////////////////////////////////////////////////////////////////

func (h *Handler) DELETE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&JsonDataDeleted); err != nil {
		log.Println(err.Error())
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}
	projectIdNum, err := strconv.Atoi(JsonDataDeleted.ProjectID)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	idNum, err := strconv.Atoi(JsonDataDeleted.Id)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	boolean, err := h.db.CheckIfGoodExists(idNum, projectIdNum)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", 500)
		return
	}
	if !boolean {
		http.NotFound(w, r)
		return
	}
	err = h.db.DeleteGoods(projectIdNum, idNum)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", 500)
		return
	}
	JsonDataDeleted.Removed = true
	responseJSON, err := json.Marshal(JsonDataDeleted)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Установка заголовка Content-Type на application/json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

func (h *Handler) GET(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	goods, err := h.db.GetGoods()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	projects, err := h.db.GetProjects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"goods":    goods,
		"projects": projects,
	}

	// Преобразование данных в JSON
	JsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Установка заголовка Content-Type
	w.Header().Set("Content-Type", "application/json")

	// Возвращение данных в виде JSON
	fmt.Fprintf(w, "%s\n", JsonData)
}
