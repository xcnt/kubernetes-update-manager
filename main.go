package main

import (
	"os"

	"kubernetes-update-manager/cli"

	"github.com/gookit/color"
)

func main() {
	app := cli.New()

	err := app.Run(os.Args)
	if err != nil {
		color.Error.Println(err)
	}
}
