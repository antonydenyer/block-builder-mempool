package example

import (
	"github.com/antonydenyer/block-builder-mempool/app"
	"net/http"

	"github.com/uptrace/bunrouter"
)

type OrgHandler struct {
	app *app.App
}

func NewOrgHandler(app *app.App) *OrgHandler {
	return &OrgHandler{
		app: app,
	}
}

func (h *OrgHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	var orgs []Org
	if err := h.app.DB().NewSelect().Model(&orgs).Relation("Owner").Scan(ctx); err != nil {
		return err
	}

	return bunrouter.JSON(w, bunrouter.H{
		"orgs": orgs,
	})
}
