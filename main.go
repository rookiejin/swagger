package main
import (
	"github.com/swaggo/swag/gen"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Version = "v1.0.0"
	app.Usage = "Automatically generate RESTful API documentation with Swagger 2.0 for Go."

	searchDir := "./"
	mainApiFile := "./main.go"
	gen.New().Build(searchDir, mainApiFile)
}
