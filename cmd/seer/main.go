package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
