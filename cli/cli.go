package cli

import (
	"gopkg.in/urfave/cli.v2"
)

// New returns a new CLI runtime for execution.
func New() *cli.App {
	app := &cli.App{
		Name:    "kubernetes-update-manager",
		Version: Version,
		Commands: []*cli.Command{
			ServerCommand(),
			UpdateCommand(),
		},
	}
	return app
}
