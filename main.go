package main

import (
	"net/http"
	"final/pkg/db"
	"final/pkg/api"
)

func main() {
	dbFile := "scheduler.db"
	if err := db.Init(dbFile); err != nil {
		panic(err)
	}

	webRoot := "./web"
	serverPort := ":7540"

	handler := http.NewServeMux()
	api.RegisterHandlers(handler)
	handler.Handle("/", http.FileServer(http.Dir(webRoot)))

	if err := http.ListenAndServe(serverPort, handler); err != nil {
		panic(err)
	}
}