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
	VERSION = "v0.2.1"
	USAGE   = `NAME:
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
		fmt.Printf(USAGE, VERSION)
	}

	flag.Parse()
}

func main() {

	//err := OpenLogfile()

	// Start terminal user interface
	err := termui.Init()
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
func OpenLogfile() error {
	f, err := os.OpenFile("filename", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	//defer to close when you're done with it, not because you think it's idiomatic!
	defer f.Close()
	//set output of logs to f
	log.SetOutput(f)
	return err
}
