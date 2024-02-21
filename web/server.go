package web

import (
	"context"
	"embed"
	"github.com/antonydenyer/block-builder-mempool/app"
	"github.com/uptrace/bunrouter"
	"net/http"
)

//go:embed static
var filesFS embed.FS

func RegisterRoutes() {
	app.OnStart("web.init", func(ctx context.Context, app *app.App) error {
		router := app.Router()
		api := app.APIRouter()

		htmlHandler := NewHtmlHandler(app)
		apiHandler := NewApiHandler(app)

		router.GET("/", htmlHandler.Home)
		router.GET("/blocks/:blockNumber", htmlHandler.Block)

		api.GET("/blocks", apiHandler.List)
		api.GET("/blocks/:blockNumber", apiHandler.Get)

		fileServer := http.FileServer(http.FS(filesFS))
		router.GET("/static/*path", bunrouter.HTTPHandler(fileServer))

		return nil
	})
}
