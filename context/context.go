package context

import (
	"log"

	termbox "github.com/nsf/termbox-go"

	"github.com/erroneousboat/slack-term/config"
	"github.com/erroneousboat/slack-term/service"
	"github.com/erroneousboat/slack-term/views"
)

const (
	// CommandMode sets the app into command mode
	CommandMode = "command"
	// InsertMode sets the app into insert mode
	InsertMode  = "insert"
)

// AppContext contains all app services and views
type AppContext struct {
	EventQueue chan termbox.Event
	Service    *service.SlackService
	View       *views.View
	Config     *config.Config
	Mode       string
}

// CreateAppContext creates an application context which can be passed
// and referenced througout the application
func CreateAppContext(flgConfig string) *AppContext {
	// Load appConfig
	appConfig, err := config.NewConfig(flgConfig)
	if err != nil {
		log.Fatalf("ERROR: not able to load appConfig file (%s): %s", flgConfig, err)
	}

	// Create Service
	svc := service.CreateSlackService(appConfig.SlackTokens)

	// Create ChatView
	view := views.CreateUIComponents(appConfig, svc)

	return &AppContext{
		EventQueue: make(chan termbox.Event, 20),
		Service:    svc,
		View:       view,
		Config:     appConfig,
		Mode:       CommandMode,
	}
}
