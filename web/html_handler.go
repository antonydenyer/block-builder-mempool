package web

import (
	"embed"
	"github.com/antonydenyer/block-builder-mempool/app"
	"github.com/antonydenyer/block-builder-mempool/service"
	"html/template"
	"net/http"
	"strconv"

	"github.com/uptrace/bunrouter"
)

//go:embed templates/*
var templates embed.FS

type HtmlHandler struct {
	app                     *app.App
	tpl                     *template.Template
	blockTransactionService *service.BlockTransactionService
}

func NewHtmlHandler(app *app.App) *HtmlHandler {
	tpl, err := template.New("").ParseFS(templates, "templates/*.html")

	if err != nil {
		panic(err)
	}

	return &HtmlHandler{
		app:                     app,
		tpl:                     tpl,
		blockTransactionService: service.NewBlockTransactionsService(app.DB()),
	}
}

func (h *HtmlHandler) Home(w http.ResponseWriter, req bunrouter.Request) error {
	blocks, err := h.blockTransactionService.Get(req.Context())
	if err != nil {
		return err
	}

	if err := h.tpl.ExecuteTemplate(w, "home.html", blocks); err != nil {
		return err
	}
	return nil
}

func (h *HtmlHandler) Block(w http.ResponseWriter, req bunrouter.Request) error {
	blockNumber, err := strconv.ParseUint(req.Param("blockNumber"), 10, 64)
	if err != nil {
		return err
	}
	blocks, err := h.blockTransactionService.GetByNumber(req.Context(), blockNumber)
	if err != nil {
		return err
	}
	if err := h.tpl.ExecuteTemplate(w, "block.html", blocks); err != nil {
		return err
	}
	return nil
}
