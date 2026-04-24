package api

import (
	"net/http"

	"final/pkg/db"
)

const Limit = 50

type TaskListResponse struct {
	Tasks []db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не разрешен"})
		return
	}

	list, err := db.ListTasks(Limit)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось загрузить список задач"})
		return
	}
	if list == nil {
		list = []db.Task{}
	}

	sendResponse(w, http.StatusOK, TaskListResponse{
		Tasks: list,
	})
}