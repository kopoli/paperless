package paperless

import (
	"fmt"
	"runtime"

	"github.com/jawher/mow.cli"
	util "github.com/kopoli/go-util"
)

func Cli(opts util.Options, args []string) error {
	progName := opts.Get("program-name", "paperless")
	progVersion := opts.Get("program-version", "undefined")
	app := cli.App(progName, "Paperless office")

	app.Version("version v", fmt.Sprintf("%s: %s\nBuilt with: %s/%s on %s/%s",
		progName, progVersion, runtime.Compiler, runtime.Version(),
		runtime.GOOS, runtime.GOARCH))

	optDbFile := app.StringOpt("d dbfile", "data/paperless.sqlite3",
		"The database file.")
	optImageDir := app.StringOpt("i image-directory", "data/images",
		"Directory to save the images in")
	optListenAddr := app.StringOpt("a address", ":8078", "Listen address and port")

	optPrintRoutes := app.BoolOpt("print-routes", false,
		"A debug option to print the REST API")

	app.Action = func() {
		opts.Set("database-file", *optDbFile)
		opts.Set("image-directory", *optImageDir)
		opts.Set("listen-address", *optListenAddr)

		if *optPrintRoutes {
			opts.Set("print-routes", "t")
		}
	}

	return app.Run(args)
}
