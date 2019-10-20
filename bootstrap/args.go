package bootstrap

import (
	"fmt"

	"github.com/jessevdk/go-flags"
)

const (
	modeStandalone = "standalone"
	modeCapture    = "capture"
	modeAPI        = "api"
)

type options struct {
	ConfigPath string `short:"c" long:"config" description:"Path to config file" default:"./config/config.yml"`
	Queue      string `short:"q" long:"queue" description:"queue connect url" default:"nats://localhost:4222"`
	Database   string `short:"d" long:"database" description:"database connect url" default:"localhost:27017"`
	Chrome     string `long:"chrome" description:"headless chrome url" default:"localhost:9222"`
	Mode       string `short:"m" long:"mode" description:"Supported modes: capture (run only application part which capture screenshots), api (run only application part which receive http requests), standalone: (run both services)" default:"standalone"`
}

func parseFlags(args []string) (options, error) {
	opt := options{}
	if _, err := flags.ParseArgs(&opt, args); err != nil {
		return options{}, fmt.Errorf(`failed to parse args: [args: %+v, error: %w]`, args, err)
	}
	return opt, nil
}
