package controller

import (
	"os"
	"pgoperator/cmd/controller/app"
)

func main() {
	command := app.NewControllerManagerCommand()
	if command == nil {
		os.Exit(1)
	}
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
