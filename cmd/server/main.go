package main

import (
	"embed"
	"github.com/KonstantinPavlov/metric-service/internal/handler"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/labstack/echo/v4"
	"html/template"
)

//go:embed views/*
var viewsFS embed.FS

func main() {
	parseFlags()	
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	webHandler := handler.MetricHandler{
		Repository: repository.NewMemStorage(),
	}

	tmpl, err := template.ParseFS(viewsFS, "views/*.html")
	if err != nil {
		return err
	}

	renderer := &handler.TemplateRenderer{
		Template: tmpl,
	}

	httpServer := echo.New()
	httpServer.Renderer = renderer
	httpServer.POST("/update/:type/:name/:value", webHandler.HandleUpdate)
	httpServer.GET("/value/:type/:name", webHandler.HandleValue)
	httpServer.GET("/", webHandler.HandleList)
	return httpServer.Start(flagRunAddr)
}
