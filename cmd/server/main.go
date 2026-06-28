package main

import (
	"html/template"

	"github.com/KonstantinPavlov/metric-service/internal/handler"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/labstack/echo/v4"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	handler := handler.MetricHandler{
		Repository: repository.NewMemStorage(),
		Template:   template.Must(template.ParseGlob("views/*.html")),
	}

	httpServer := echo.New()
	httpServer.Renderer = &handler
	httpServer.POST("/update/:type/:name/:value", handler.HandleUpdate)
	httpServer.GET("/value/:type/:name", handler.HandleValue)
	httpServer.GET("/", handler.HandleList)
	return httpServer.Start(":8080")
}
