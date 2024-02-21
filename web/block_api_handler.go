package web

import (
	"github.com/antonydenyer/block-builder-mempool/app"
	"github.com/antonydenyer/block-builder-mempool/service"
	"net/http"
	"strconv"

	"github.com/uptrace/bunrouter"
)

type ApiHandler struct {
	app                     *app.App
	blockTransactionService *service.BlockTransactionService
}

func NewApiHandler(app *app.App) *ApiHandler {
	return &ApiHandler{
		app:                     app,
		blockTransactionService: service.NewBlockTransactionsService(app.DB()),
	}
}

func (h *ApiHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	blocks, err := h.blockTransactionService.Get(ctx)
	if err != nil {
		return err
	}

	return bunrouter.JSON(w, bunrouter.H{
		"blocks": blocks,
	})
}

func (h *ApiHandler) Get(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	blockNumber, err := strconv.ParseUint(req.Param("blockNumber"), 10, 64)
	if err != nil {
		return err
	}
	blocks, err := h.blockTransactionService.GetByNumber(ctx, blockNumber)
	if err != nil {
		return err
	}

	return bunrouter.JSON(w, bunrouter.H{
		"blocks": blocks,
	})
}
