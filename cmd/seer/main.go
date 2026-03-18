package main

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/alexisbouchez/hyperseer/internal/exit"
)

func main() {
	app := &cli.Command{
		Name:  "seer",
		Usage: "Hyperseer CLI",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return errors.New("no command yet")
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		exit.WithError(err)
	}
}
