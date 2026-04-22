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


func sendResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
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
		log.Println("API: Получен GET запрос /api/task")
		id := r.FormValue("id")
		if len(id) == 0 {
			sendResponse(w, map[string]string{"error": "id не указан"})
			return
		}
		item, err := db.GetTaskByID(id)
		if err != nil {
			sendResponse(w, map[string]string{"error": "задача не найдена"})
			return
		}
		sendResponse(w, item)

	case http.MethodDelete:
		log.Println("API: Получен DELETE запрос /api/task")
		id := r.FormValue("id")
		if err := db.DeleteTask(id); err != nil {
			sendResponse(w, map[string]string{"error": err.Error()})
			return
		}
		sendResponse(w, map[string]any{})

	case http.MethodPost, http.MethodPut:
		log.Printf("API: Получен %s запрос /api/task", r.Method)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			sendResponse(w, map[string]string{"error": "ошибка чтения"})
			return
		}
		
		var t db.Task
		if err := json.Unmarshal(body, &t); err != nil {
			sendResponse(w, map[string]string{"error": "невалидный JSON"})
			return
		}

		if t.Title == "" {
			sendResponse(w, map[string]string{"error": "пустой заголовок"})
			return
		}

		now := time.Now().Truncate(24 * time.Hour)
		if t.Date == "" {
			t.Date = now.Format(TimeLayout)
		}

		parsedDate, err := time.Parse(TimeLayout, t.Date)
		if err != nil {
			sendResponse(w, map[string]string{"error": "формат даты 20060102"})
			return
		}

		if parsedDate.Before(now) {
			if t.Repeat == "" {
				t.Date = now.Format(TimeLayout)
			} else {
				next, err := NextDate(now, t.Date, t.Repeat)
				if err != nil {
					sendResponse(w, map[string]string{"error": err.Error()})
					return
				}
				t.Date = next
			}
		}

		if r.Method == http.MethodPut {
			if t.ID == "" {
				sendResponse(w, map[string]string{"error": "отсутствует id"})
				return
			}
			if err := t.Update(); err != nil {
				sendResponse(w, map[string]string{"error": err.Error()})
				return
			}
			sendResponse(w, map[string]any{})
		} else {
			newID, err := t.Save()
			if err != nil {
				sendResponse(w, map[string]string{"error": err.Error()})
				return
			}
			sendResponse(w, map[string]string{"id": fmt.Sprintf("%d", newID)})
		}

	default:
		log.Printf("API: Метод %s не поддерживается", r.Method)
		sendResponse(w, map[string]string{"error": "метод не разрешен"})
	}
}

func taskDoneHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("API: Выполнение задачи /api/task/done")
	id := r.FormValue("id")
	task, err := db.GetTaskByID(id)
	if err != nil {
		sendResponse(w, map[string]string{"error": "объект не найден"})
		return
	}

	if task.Repeat == "" {
		db.DeleteTask(id)
	} else {
		now := time.Now().Truncate(24 * time.Hour)
		next, _ := NextDate(now, task.Date, task.Repeat)
		task.Date = next
		task.Update()
	}
	sendResponse(w, map[string]any{})
}