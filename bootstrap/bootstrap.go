package bootstrap

import (
	"context"
	"fmt"

	"github.com/leveldorado/screenshot/store"

	"github.com/leveldorado/screenshot/api"
	"github.com/leveldorado/screenshot/capture"
	"github.com/leveldorado/screenshot/queue"
)

type runner interface {
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
}

func Build(ctx context.Context, args []string) (runner, error) {
	opt, err := parseFlags(args)
	if err != nil {
		return nil, fmt.Errorf(`failed to parse args: [args: %+v, error: %w]`, args, err)
	}
	c, err := readConfig(opt.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf(`failed to read config: [path: %s, error: %w]`, opt.ConfigPath, err)
	}
	nats, err := queue.NewNATS(opt.Queue, c.Queue.BufferSize, c.Queue.ConnectTimeout)
	if err != nil {
		return nil, fmt.Errorf(`failed to create nats: [addr: %s, error: %w]`, opt.Queue, err)
	}
	fs, ms, err := getFileAndMetadataStore(ctx, opt.Database, c)
	if err != nil {
		return nil, err
	}
	switch opt.Mode {
	case modeAPI:
		return buildAPI(c, opt, nats, fs, ms)
	case modeCapture:
		return buildCapture(ctx, c, opt, nats, fs, ms)
	case modeStandalone:
		apiH, err := buildAPI(c, opt, nats, fs, ms)
		if err != nil {
			return nil, err
		}
		captureH, err := buildCapture(ctx, c, opt, nats, fs, ms)
		return combinedRunner{parts: []runner{apiH, captureH}}, err
	default:
		return nil, fmt.Errorf(`unsupported mode %s. please use one of (standalone, api, capture)`, opt.Mode)
	}
}

func getFileAndMetadataStore(ctx context.Context, url string, c config) (*store.MongodbGridFSFileRepo, *store.MongodbMetadataRepo, error) {
	cl, err := store.BuildMongoClient(ctx, url)
	if err != nil {
		return nil, nil, fmt.Errorf(`failed to build mongo client: [url: %s, error: %w]`, url, err)
	}
	fs, err := store.NewMongodbGridFSFileRepo(ctx, cl, c.Database.Name)
	if err != nil {
		return nil, nil, fmt.Errorf(`failed create mongodb gridfs repo: [database: %s, error: %w]`, c.Database.Name, err)
	}
	ms := store.NewMongodbMetadataRepo(cl, c.Database.Name, c.Database.Collections.Metadata, c.Database.Collections.VersionCounter)
	if err = ms.EnsureIndexes(ctx); err != nil {
		return nil, nil, fmt.Errorf(`failed to ensure metadata indexes: [error: %w]`, err)
	}
	return fs, ms, nil
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

func buildCapture(ctx context.Context, c config, opt flagOptions, nats *queue.NATS, fs *store.MongodbGridFSFileRepo, ms *store.MongodbMetadataRepo) (*capture.QueueSubscriptionHandler, error) {
	sh, err := capture.NewChromeShotMaker(ctx, opt.Chrome)
	if err != nil {
		return nil, fmt.Errorf(`failed create chrome shot maker: [url: %s, error: %w]`, opt.Chrome, err)
	}
	return capture.NewQueueSubscriptionHandler(capture.NewDefaultService(sh, fs, ms, c.Screenshot.Format, c.Screenshot.Quality), nats, c.Queue.HandleMessageTimeout), nil
}

func buildAPI(c config, opt flagOptions, nats *queue.NATS, fs *store.MongodbGridFSFileRepo, ms *store.MongodbMetadataRepo) (*api.HTTPHandler, error) {
	return api.NewHTTPHandler(api.NewDefaultService(fs, ms, nats, c.Queue.WaitReplyTimeout), opt.Address), nil
}
