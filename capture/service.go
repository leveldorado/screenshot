package capture

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/leveldorado/screenshot/store"
)

type shotMaker interface {
	MakeShot(ctx context.Context, url, format string, quality int) (io.Reader, error)
}

type fileSaver interface {
	Save(ctx context.Context, file io.Reader, fileID, filename string) error
}

type metadataSaver interface {
	Save(ctx context.Context, doc *store.Metadata) error
}

type DefaultService struct {
	sm          shotMaker
	fs          fileSaver
	ms          metadataSaver
	shotFormat  string
	shotQuality int
}

func NewDefaultService(sm shotMaker, fs fileSaver, ms metadataSaver, format string, quality int) *DefaultService {
	return &DefaultService{sm: sm, fs: fs, ms: ms, shotFormat: format, shotQuality: quality}
}

func (s *DefaultService) MakeShotAndSave(ctx context.Context, url string) (store.Metadata, error) {
	shot, err := s.sm.MakeShot(ctx, url, s.shotFormat, s.shotQuality)
	if err != nil {
		return store.Metadata{}, fmt.Errorf(`failed to make shot: [url: %s, error: %w]`, url, err)
	}
	fileID := uuid.New().String()
	if err = s.fs.Save(ctx, shot, fileID, url); err != nil {
		return store.Metadata{}, fmt.Errorf(`failed to store file: [id: %s, name: %s, error: %w]`, fileID, url, err)
	}
	metadata := store.Metadata{
		ID:      uuid.New().String(),
		Url:     url,
		Format:  s.shotFormat,
		Quality: s.shotQuality,
		FileID:  fileID,
	}
	if err = s.ms.Save(ctx, &metadata); err != nil {
		return store.Metadata{}, fmt.Errorf(`failed to save screen shot metadata: [doc: %+v, error: %w]`, metadata, err)
	}
	return metadata, nil
}
