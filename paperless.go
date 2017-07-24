package main

import (
	"fmt"
	"os"
	"strings"

	util "github.com/kopoli/go-util"
	"github.com/kopoliitti/paperless/lib"
)

func printErr(err error, message string, arg ...string) {
	msg := ""
	if err != nil {
		msg = fmt.Sprintf(" (error: %s)", err)
	}
	fmt.Fprintf(os.Stderr, "Error: %s%s.%s\n", message, strings.Join(arg, " "), msg)
}

func fault(err error, message string, arg ...string) {
	printErr(err, message, arg...)
	os.Exit(1)
}

func main() {
	opts := util.NewOptions()

	err := paperless.Cli(opts, os.Args)
	if err != nil {
		fault(err, "Command line parsing failed")
	}

	err = paperless.StartWeb(opts)
	if err != nil {
		fault(err, "Starting paperless web server failed")
	}
}

// Local Variables:
// outline-regexp: "^////*\\|^func\\|^import"
// End:
