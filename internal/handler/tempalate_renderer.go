package handler

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)


type TemplateRenderer struct {
	Template *template.Template
}

func (r *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return r.Template.ExecuteTemplate(w, name, data)
}