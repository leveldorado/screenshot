package api

import (
	"context"
	"io"

	"github.com/leveldorado/screenshot/store"
)

type fileGetter interface {
	Get(ctx context.Context, fileID string) (io.Reader, error)
}

type screenshotMetadataGetter interface {
	Get(ctx context.Context, url string, version int) (store.ScreenshotMetadata, error)
	GetAllVersions(ctx context.Context, url string) ([]store.ScreenshotMetadata, error)
}

type DefaultService struct {
}
