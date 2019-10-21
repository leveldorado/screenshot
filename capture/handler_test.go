package capture

import (
	"context"

	"github.com/leveldorado/screenshot/store"
	"github.com/stretchr/testify/mock"
)

type mockService struct {
	mock.Mock
}

func (m *mockService) MakeShotAndSave(ctx context.Context, url string) (store.Metadata, error) {
	args := m.Called(ctx, url)
	return args.Get(0).(store.Metadata), args.Error(1)
}
