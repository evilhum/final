package api

import (
	"net/http"
	"io"
	"encoding/json"
	"fmt"
	"time"
	"log"
	"final/pkg/db"
)

func sendResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	resp, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Ошибка сериализации", http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTask(w, r)
	case http.MethodDelete:
		deleteTask(w, r)
	case http.MethodPost:
		postTask(w, r)
	case http.MethodPut: 
		putTask(w, r)
	default:
		sendResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "метод не поддерживается"})
	}
}

func getTask(w http.ResponseWriter, r *http.Request) {
	log.Println("API: Получен GET запрос /api/task")
	id := r.FormValue("id")
	if id == "" {
		sendResponse(w, http.StatusBadRequest, map[string]string{"error": "не указан id"})
		return
	}
	item, err := db.GetTaskByID(id)
	if err != nil {
		sendResponse(w, http.StatusNotFound, map[string]string{"error": "задача не найдена"})
		return
	}
	sendResponse(w, http.StatusOK, item)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	log.Println("API: Получен DELETE запрос /api/task")
	id := r.FormValue("id")
	if id == "" {
		sendResponse(w, http.StatusBadRequest, map[string]string{"error": "не указан id"})
		return
	}
	if _, err := db.GetTaskByID(id); err != nil {
		sendResponse(w, http.StatusNotFound, map[string]string{"error": "задача не найдена"})
		return
	}
	if err := db.DeleteTask(id); err != nil {
		sendResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	sendResponse(w, http.StatusOK, map[string]any{})
}

func postTask(w http.ResponseWriter, r *http.Request) {
	log.Println("API: Получен POST запрос /api/task")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, map[string]string{"error": "ошибка чтения"})
		return
	}

	var t db.Task
	if err := json.Unmarshal(body, &t); err != nil {
		sendResponse(w, http.StatusBadRequest, map[string]string{"error": "невалидный JSON"})
		return
	}

	if t.Title == "" {
		sendResponse(w, http.StatusBadRequest, map[string]string{"error": "не указан заголовок"})
		return
	}

	now := time.Now().Truncate(24 * time.Hour)
	if t.Date == "" {
		t.Date = now.Format(TimeLayout)
	}

	parsedDate, err := time.Parse(TimeLayout, t.Date)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, map[string]string{"error": "неверный формат даты"})
		return
	}

	if parsedDate.Before(now) {
		if t.Repeat == "" {
			t.Date = now.Format(TimeLayout)
		} else {
			next, err := NextDate(now, t.Date, t.Repeat)
			if err != nil {
				sendResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			t.Date = next
		}
	}

	newID, err := t.Save()
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	sendResponse(w, http.StatusOK, map[string]string{"id": fmt.Sprintf("%d", newID)})
}

func putTask(w http.ResponseWriter, r *http.Request) {
	log.Println("API: Получен PUT запрос /api/task")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, map[string]string{"error": "ошибка чтения"})
		return
	}

	var t db.Task
	if err := json.Unmarshal(body, &t); err != nil {
		sendResponse(w, http.StatusBadRequest, map[string]string{"error": "невалидный JSON"})
		return
	}

	if t.ID == "" {
		sendResponse(w, http.StatusBadRequest, map[string]string{"error": "не указан id"})
		return
	}

	if t.Title == "" {
		sendResponse(w, http.StatusBadRequest, map[string]string{"error": "не указан заголовок"})
		return
	}

	now := time.Now().Truncate(24 * time.Hour)
	if t.Date == "" {
		t.Date = now.Format(TimeLayout)
	}

	parsedDate, err := time.Parse(TimeLayout, t.Date)
	if err != nil {
		sendResponse(w, http.StatusBadRequest, map[string]string{"error": "неверный формат даты"})
		return
	}

	if parsedDate.Before(now) {
		if t.Repeat == "" {
			t.Date = now.Format(TimeLayout)
		} else {
			next, err := NextDate(now, t.Date, t.Repeat)
			if err != nil {
				sendResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			t.Date = next
		}
	}

	if err := t.Update(); err != nil {
		sendResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	sendResponse(w, http.StatusOK, map[string]any{})
}

func taskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не разрешен"})
		return
	}
	log.Println("API: Выполнение задачи /api/task/done")
	id := r.FormValue("id")
	if id == "" {
		sendResponse(w, http.StatusBadRequest, map[string]string{"error": "не указан id"})
		return
	}
	task, err := db.GetTaskByID(id)
	if err != nil {
		sendResponse(w, http.StatusNotFound, map[string]string{"error": "задача не найдена"})
		return
	}

	if task.Repeat == "" {
		db.DeleteTask(id)
	} else {
		now := time.Now().Truncate(24 * time.Hour)
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			sendResponse(w, http.StatusBadRequest, map[string]string{"error": "ошибка вычисления следующей даты"})
			return
		}
		task.Date = next
		task.Update()
	}
	sendResponse(w, http.StatusOK, map[string]any{})
}