package handler

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type TemplateRenderer struct {
	Template *template.Template
}

func (tr *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/html")
	return tr.Template.ExecuteTemplate(w, name, data)
}
