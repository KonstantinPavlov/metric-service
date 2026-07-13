package main

import (
	"embed"
	"html/template"

	"github.com/KonstantinPavlov/metric-service/internal/handler"
	"github.com/KonstantinPavlov/metric-service/internal/logger"
	"github.com/KonstantinPavlov/metric-service/internal/repository"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

//go:embed views/*
var viewsFS embed.FS

func main() {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	parseFlags()
	if err := run(zapLogger); err != nil {
		panic(err)
	}
}

func run(zapLogger *zap.Logger) error {

	webHandler := handler.NewMetricHandler(
		repository.NewMemStorage(),
		zapLogger,
	)

	tmpl, err := template.ParseFS(viewsFS, "views/*.html")
	if err != nil {
		return err
	}

	renderer := &handler.TemplateRenderer{
		Template: tmpl,
	}

	httpServer := echo.New()
	httpServer.Use(logger.ZapMiddleware(zapLogger))
	httpServer.Renderer = renderer
	httpServer.POST("/update/:type/:name/:value", webHandler.HandleParamUpdate)
	httpServer.POST("/update/", webHandler.HandleBodyUpdate)	
	httpServer.GET("/value/:type/:name", webHandler.HandleGetValue)
	httpServer.POST("/value/", webHandler.HandlePostValue)
	httpServer.GET("/", webHandler.HandleList)
	return httpServer.Start(flagRunAddr)
}
