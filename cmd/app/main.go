package main

import (
	"github.com/antonydenyer/block-builder-mempool/cmd/app/chain"
	"github.com/antonydenyer/block-builder-mempool/cmd/app/db"
	"github.com/antonydenyer/block-builder-mempool/cmd/app/pool"
	"github.com/antonydenyer/block-builder-mempool/cmd/app/web"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name: "bun",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "env",
				Value:   "dev",
				Usage:   "environment",
				EnvVars: []string{"ENV"}},
		},
		Commands: []*cli.Command{
			chain.Command(),
			pool.Command(),
			web.Command(),
			db.Command(),
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
