package main

import (
	"os"

	"github.com/swaggo/swag/gen"
	"github.com/urfave/cli"
	"fmt"
)

func main() {
	app := cli.NewApp()
	app.Version = "v1.0.0"
	app.Usage = "Automatically generate RESTful API documentation with Swagger 2.0 for Go."

	dir := "C:\\gopath\\src\\app\\"

	fmt.Print( dir + "main.go")
	os.Exit(0)

	app.Commands = []cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "create docs.go",
			Action: func(c *cli.Context) error {
				searchDir := dir
				mainApiFile := dir + "main.go"
				gen.New().Build(searchDir, mainApiFile)
				return nil
			},
		},
	}

	app.Run(os.Args)
}
