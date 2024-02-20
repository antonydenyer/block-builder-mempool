package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/uptrace/bun/dialect/pgdialect"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bunrouter"
	"github.com/urfave/cli/v2"
)

type appCtxKey struct{}

func AppFromContext(ctx context.Context) *App {
	return ctx.Value(appCtxKey{}).(*App)
}

func ContextWithApp(ctx context.Context, app *App) context.Context {
	ctx = context.WithValue(ctx, appCtxKey{}, app)
	return ctx
}

type App struct {
	ctx context.Context
	cfg *Config

	stopping uint32
	stopCh   chan struct{}

	onStop      appHooks
	onAfterStop appHooks

	router    *bunrouter.Router
	apiRouter *bunrouter.Group

	// lazy init
	dbOnce sync.Once
	db     *bun.DB
}

func New(ctx context.Context, cfg *Config) *App {
	app := &App{
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
	app.ctx = ContextWithApp(ctx, app)
	app.initRouter()
	return app
}

func StartCLI(c *cli.Context) (context.Context, *App, error) {
	return Start(c.Context, c.Command.Name, c.String("env"))
}

func Start(ctx context.Context, service, envName string) (context.Context, *App, error) {
	fmt.Println(envName)
	cfg, err := ReadConfig(FS(), service, envName)
	if err != nil {
		return nil, nil, err
	}
	return StartConfig(ctx, cfg)
}

func StartConfig(ctx context.Context, cfg *Config) (context.Context, *App, error) {
	rand.Seed(time.Now().UnixNano())

	app := New(ctx, cfg)
	if err := onStart.Run(ctx, app); err != nil {
		return nil, nil, err
	}
	return app.ctx, app, nil
}

func (app *App) Stop() {
	_ = app.onStop.Run(app.ctx, app)
	_ = app.onAfterStop.Run(app.ctx, app)
}

func (app *App) OnStop(name string, fn HookFunc) {
	app.onStop.Add(newHook(name, fn))
}

func (app *App) OnAfterStop(name string, fn HookFunc) {
	app.onAfterStop.Add(newHook(name, fn))
}

func (app *App) Context() context.Context {
	return app.ctx
}

func (app *App) Config() *Config {
	return app.cfg
}

func (app *App) Running() bool {
	return !app.Stopping()
}

func (app *App) Stopping() bool {
	return atomic.LoadUint32(&app.stopping) == 1
}

func (app *App) IsDebug() bool {
	return app.cfg.Debug
}

func (app *App) Router() *bunrouter.Router {
	return app.router
}

func (app *App) APIRouter() *bunrouter.Group {
	return app.apiRouter
}

func (app *App) DB() *bun.DB {
	app.dbOnce.Do(func() {
		var db *bun.DB
		var pg *sql.DB
		if app.cfg.DB.Username != "" {
			pg = sql.OpenDB(pgdriver.NewConnector(
				pgdriver.WithDSN(app.cfg.DB.DSN),
				pgdriver.WithUser(app.cfg.DB.Username),
				pgdriver.WithPassword(app.cfg.DB.Password),
			))
		} else {
			pg = sql.OpenDB(pgdriver.NewConnector(
				pgdriver.WithDSN(app.cfg.DB.DSN),
			))
		}
		db = bun.NewDB(pg, pgdialect.New(), bun.WithDiscardUnknownColumns())

		_, err := db.Exec("SELECT 1")
		if err != nil {
			log.Fatal("Failed to start db")
		}

		app.OnStop("db.Close", func(_ context.Context, _ *App) error {
			return db.Close()
		})

		app.db = db
	})
	return app.db
}

//------------------------------------------------------------------------------

func (app *App) WaitExitSignal() os.Signal {
	ch := make(chan os.Signal, 3)
	signal.Notify(
		ch,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	return <-ch
}
