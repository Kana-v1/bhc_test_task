package handler

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "bhc_test_task/docs"
)

func SetupDocsHandler() {
	http.HandleFunc("/docs/", httpSwagger.Handler())
}
