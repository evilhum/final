package api

import (
	"net/http"

	"final/pkg/db"
)

type TaskListResponse struct {
	Tasks []db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	list, err := db.ListTasks(50)
	if err != nil {
		sendResponse(w, map[string]string{"error": "Не удалось загрузить список задач"})
		return
	}
	if list == nil {
        list = []db.Task{}
    }
	sendResponse(w, TaskListResponse{
		Tasks: list,
	})
}

