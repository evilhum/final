package db

import (
	_ "database/sql"
	"fmt"
	"log" 
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func (t Task) Save() (int64, error) {
	log.Printf("DB: Попытка создания задачи: %s", t.Title)
	res, err := DB.Exec(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		t.Date, t.Title, t.Comment, t.Repeat,
	)
	if err != nil {
		return 0, fmt.Errorf("ошибка вставки: %w", err)
	}
	return res.LastInsertId()
}

func (t Task) Update() error {
	log.Printf("DB: Обновление задачи ID %s", t.ID)
	res, err := DB.Exec(
		`UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`,
		t.Date, t.Title, t.Comment, t.Repeat, t.ID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("запись не найдена")
	}
	return nil
}

func ListTasks(limit int) ([]Task, error) {
	log.Printf("DB: Запрос списка задач, лимит: %d", limit)
	rows, err := DB.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, err
		}
		records = append(records, t)
	}
	if err := rows.Err(); err != nil {
        return nil, err
    } 
	return records, nil
}

func GetTaskByID(id string) (Task, error) {
	var t Task
	err := DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id).
		Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	return t, err
}

func DeleteTask(id string) error {
	log.Printf("DB: Удаление задачи ID %s", id)
	_, err := DB.Exec("DELETE FROM scheduler WHERE id = ?", id)
	return err
}