package example

import (
	"github.com/antonydenyer/block-builder-mempool/app"
	"net/http"

	"github.com/uptrace/bunrouter"
)

type UserHandler struct {
	app *app.App
}

func NewUserHandler(app *app.App) *UserHandler {
	return &UserHandler{
		app: app,
	}
}

func (h *UserHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	var users []User

	if err := h.app.DB().NewSelect().Model(&users).Scan(ctx); err != nil {
		return err
	}

	return bunrouter.JSON(w, bunrouter.H{
		"users": users,
	})
}

func (h *UserHandler) Get(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	id := req.Param("id")

	var user User
	if err := h.app.DB().NewSelect().Where("id = ?", id).Model(&user).Scan(ctx); err != nil {
		return err
	}

	return bunrouter.JSON(w, user)
}
