package web

import (
	"fmt"
	"github.com/antonydenyer/block-builder-mempool/app"
	"github.com/antonydenyer/block-builder-mempool/httputil"
	"github.com/antonydenyer/block-builder-mempool/web"
	"github.com/urfave/cli/v2"
	"log"
	"net/http"
	"time"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "web",
		Usage: "start API web",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "addr",
				Value: "0.0.0.0:8000",
				Usage: "serve address",
			},
		},
		Action: func(c *cli.Context) error {
			web.RegisterRoutes()
			ctx, app, err := app.Start(c.Context, "web", c.String("env"))
			if err != nil {
				return err
			}
			defer app.Stop()

			var handler http.Handler
			handler = app.Router()
			handler = httputil.ExitOnPanicHandler{Next: handler}

			srv := &http.Server{
				Addr:         c.String("addr"),
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  60 * time.Second,
				Handler:      handler,
			}
			go func() {
				if err := srv.ListenAndServe(); err != nil && !isServerClosed(err) {
					log.Printf("ListenAndServe failed: %s", err)
				}
			}()

			fmt.Printf("listening on http://%s\n", srv.Addr)
			fmt.Println(app.WaitExitSignal())

			return srv.Shutdown(ctx)
		},
	}

}

func isServerClosed(err error) bool {
	return err.Error() == "http: Server closed"
}
