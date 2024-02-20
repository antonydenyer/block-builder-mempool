package example

import (
	"context"
	"github.com/antonydenyer/block-builder-mempool/app"
)

func init() {
	app.OnStart("example.init", func(ctx context.Context, app *app.App) error {
		router := app.Router()
		api := app.APIRouter()

		welcomeHandler := NewWelcomeHandler(app)
		userHandler := NewUserHandler(app)
		orgHandler := NewOrgHandler(app)

		router.GET("/", welcomeHandler.Welcome)
		router.GET("/hello", welcomeHandler.Hello)

		api.GET("/users", userHandler.List)
		api.GET("/users/:id", userHandler.Get)
		api.GET("/orgs", orgHandler.List)

		return nil
	})
}
