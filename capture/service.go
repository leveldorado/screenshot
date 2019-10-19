package capture

import (
	"context"
	"fmt"
	"io"

	"go.mongodb.org/mongo-driver/mongo"
)

type shotMaker interface {
	MakeShot(ctx context.Context, url string) (io.Reader, error)
}

type fileStore interface {
	StoreFile(ctx context.Context, file io.Reader, fileID, filename string) (error)
}

type screenshotMetadata interface {

}

type DefaultService struct {
	sm   shotMaker
	repo *mongo.Client
}

func NewDefaultService(sm shotMaker) *DefaultService {
	return &DefaultService{sm: sm}
}

func (s *DefaultService) MakeShotAndSave(ctx context.Context, url string) error {
	shot, err := s.sm.MakeShot(ctx, url)
	if err != nil {
		return fmt.Errorf(`failed to make shot: [url: %s, error: %w]`, url, err)
	}
	s.repo.Database("test").
}
