package main

import (
	"bhc_test_task/api/handler"
	"bhc_test_task/manager"
	"bhc_test_task/model/data_storage"
	"log"
	"net/http"
)

func main() {
	dataStorage := data_storage.NewLocalDataStorage()
	clientsManager := manager.NewClientsManager(dataStorage)
	handler.SetupHandler(clientsManager)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
