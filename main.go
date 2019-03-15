package main

import (
	"os"

	"kubernetes-update-manager/cli"
)

func main() {
	app := cli.New()

	app.Run(os.Args)
}
