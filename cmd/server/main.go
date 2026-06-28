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
	webHandler := handler.MetricHandler{
		Repository: repository.NewMemStorage(),
	}

	renderer := &handler.TemplateRenderer{
		Template: template.Must(template.ParseGlob("views/*.html")),
	}

	httpServer := echo.New()
	httpServer.Renderer = renderer
	httpServer.POST("/update/:type/:name/:value", webHandler.HandleUpdate)
	httpServer.GET("/value/:type/:name", webHandler.HandleValue)
	httpServer.GET("/", webHandler.HandleList)
	return httpServer.Start("0.0.0.0:8080")
}
