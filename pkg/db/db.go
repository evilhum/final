package db

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

const schema = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(255) NOT NULL DEFAULT "",
    comment TEXT NOT NULL DEFAULT "",
    repeat VARCHAR(128) NOT NULL DEFAULT ""
);
CREATE INDEX idx_date ON scheduler (date);
`

func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	var install bool
	if err != nil {
		install = true
		log.Printf("DB: Файл %s не найден, будет создана новая база данных", dbFile)
	} else {
		log.Printf("DB: Обнаружен существующий файл базы данных %s", dbFile)
	}

	dbConn, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Printf("DB: Критическая ошибка при открытии БД: %v", err)
		return err
	}
	DB = dbConn

	if install {
		if _, err := DB.Exec(schema); err != nil {
			log.Printf("DB: Ошибка при установке схемы таблиц: %v", err)
			return err
		}
		log.Println("DB: Схема успешно установлена")
	}

	log.Println("DB: Соединение с базой данных успешно установлено")
	return nil
}