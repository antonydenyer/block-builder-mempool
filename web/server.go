package web

import (
	"context"
	"github.com/antonydenyer/block-builder-mempool/app"
)

func RegisterRoutes() {
	app.OnStart("web.init", func(ctx context.Context, app *app.App) error {
		router := app.Router()
		api := app.APIRouter()

		homeHandler := NewHomeHandler(app)
		blockHandler := NewBlockHandler(app)

		router.GET("/", homeHandler.Home)

		api.GET("/blocks", blockHandler.List)

		return nil
	})
}
