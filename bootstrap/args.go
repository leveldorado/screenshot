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

type flagOptions struct {
	ConfigPath string `short:"c" long:"config" description:"Path to config file" default:"config.yml"`
	Queue      string `short:"q" long:"queue" description:"queue connect url" default:"nats://localhost:4222"`
	Database   string `short:"d" long:"database" description:"database connect url" default:"mongodb://localhost:27017"`
	Chrome     string `long:"chrome" description:"headless chrome url" default:"localhost:9222"`
	Mode       string `short:"m" long:"mode" description:"Supported modes: capture (run only application part which capture screenshots), api (run only application part which receive http requests), standalone: (run both services)" default:"standalone"`
	Address    string `long:"address" description:"address (host and port) on which api listen http request" default:":9000"`
}

func parseFlags(args []string) (flagOptions, error) {
	opt := flagOptions{}
	if _, err := flags.ParseArgs(&opt, args); err != nil {
		return flagOptions{}, fmt.Errorf(`failed to parse args: [args: %+v, error: %w]`, args, err)
	}
	return opt, nil
}
