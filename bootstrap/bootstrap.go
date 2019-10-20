package bootstrap

import (
	"context"
	"fmt"

	"github.com/leveldorado/screenshot/api"

	"github.com/leveldorado/screenshot/capture"
)

type runner interface {
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
}

func Build(args []string) (runner, error) {
	opt, err := parseFlags(args)
	if err != nil {
		return nil, fmt.Errorf(`failed to parse args: [args: %+v, error: %w]`, args, err)
	}
	c, err := readConfig(opt.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf(`failed to read config: [path: %s, error: %w]`, opt.ConfigPath, err)
	}
	switch opt.Mode {
	case modeAPI:
		return buildAPI(c, opt)
	case modeCapture:
		return buildCapture(c, opt)
	case modeStandalone:
		apiH, err := buildAPI(c, opt)
		if err != nil {
			return nil, err
		}
		captureH, err := buildCapture(c, opt)
		return combinedRunner{parts: []runner{apiH, captureH}}, err
	default:
		return nil, fmt.Errorf(`unsupported mode %s. please use one of (standalone, api, capture)`, opt.Mode)
	}
}

type combinedRunner struct {
	parts []runner
}

func (cr combinedRunner) Run(ctx context.Context) error {
	for _, p := range cr.parts {
		if err := p.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (cr combinedRunner) Stop(ctx context.Context) error {
	for _, p := range cr.parts {
		if err := p.Stop(ctx); err != nil {
			return err
		}
	}
	return nil
}

func buildCapture(c config, opt options) (*capture.QueueSubscriptionHandler, error) {
	return &capture.QueueSubscriptionHandler{}, nil
}

func buildAPI(c config, opt options) (*api.HTTPHandler, error) {
	return &api.HTTPHandler{}, nil
}
