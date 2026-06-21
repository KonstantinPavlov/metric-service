package main

import (
	"net/http"

	"github.com/KonstantinPavlov/metric-service/internal/handler"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	handler := handler.MetricHandler{
		Repository: repository.NewMemStorage(),
	}
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/{type}/{name}/{value}`, handler.HandleUpdate)
	return http.ListenAndServe(":8080", mux)
}
