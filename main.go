package main

import (
	"flag"
	"fmt"
	"log"
	"os/user"
	"path"

	"github.com/erroneousboat/slack-term/context"
	"github.com/erroneousboat/slack-term/handlers"
	termbox "github.com/nsf/termbox-go"

	"github.com/gizak/termui"
	"os"
)

const (
	version = "v0.2.1"
	usage   = `NAME:
    slack-term - slack client for your terminal

USAGE:
    slack-term -config [path-to-config]

VERSION:
    %s

GLOBAL OPTIONS:
   --help, -h
`
)

var (
	flgConfig string
)

func init() {
	// Get home dir for config file default
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Parse flags
	flag.StringVar(
		&flgConfig,
		"config",
		path.Join(usr.HomeDir, ".config/slack-term/slack-term.json"),
		"location of config file",
	)

	flag.Usage = func() {
		fmt.Printf(usage, version)
	}

	flag.Parse()
}

func main() {

	var err error

	// Start terminal user interface
	err = termui.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer termui.Close()

	// Create context
	ctx := context.CreateAppContext(flgConfig)

	// Register handlers
	handlers.RegisterEventHandlers(ctx)

	go func() {
		for {
			ctx.EventQueue <- termbox.PollEvent()
		}
	}()

	termui.Loop()
}