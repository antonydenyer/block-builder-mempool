package web

import (
	"embed"
	"github.com/antonydenyer/block-builder-mempool/app"
	"net/http"
	"text/template"

	"github.com/uptrace/bunrouter"
)

//go:embed templates/*
var templates embed.FS

type HomeHandler struct {
	app *app.App
	tpl *template.Template
}

func NewHomeHandler(app *app.App) *HomeHandler {
	tpl, err := template.New("").ParseFS(templates, "templates/*.html")
	if err != nil {
		panic(err)
	}

	return &HomeHandler{
		app: app,
		tpl: tpl,
	}
}

func (h *HomeHandler) Home(w http.ResponseWriter, _ bunrouter.Request) error {
	if err := h.tpl.ExecuteTemplate(w, "home.html", nil); err != nil {
		return err
	}
	return nil
}
