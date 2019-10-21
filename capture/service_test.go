package capture

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/google/uuid"

	"github.com/leveldorado/screenshot/store"

	"github.com/stretchr/testify/mock"
)

type mockShotMaker struct {
	mock.Mock
}

func (m *mockShotMaker) MakeShot(ctx context.Context, url, format string, quality int) (io.Reader, error) {
	args := m.Called(ctx, url, format, quality)
	return args.Get(0).(io.Reader), args.Error(1)
}

type mockFileSaver struct {
	mock.Mock
}

func (m *mockFileSaver) Save(ctx context.Context, file io.Reader, fileID, filename string) error {
	return m.Called(ctx, file, fileID, filename).Error(0)
}

type mockMetadataSaver struct {
	mock.Mock
}

func (m *mockMetadataSaver) Save(ctx context.Context, doc *store.Metadata) error {
	return m.Called(ctx, doc).Error(0)
}

func TestDefaultService_MakeShot(t *testing.T) {
	sm := &mockShotMaker{}
	url := uuid.New().String()
	format := "jpeg"
	quality := 80
	file := strings.NewReader(uuid.New().String())
	sm.On("MakeShot", mock.Anything, url, format, quality).Return(file, nil)
	fs := &mockFileSaver{}
	fs.On("Save", mock.Anything, file, mock.Anything, url).Return(nil)
	ms := &mockMetadataSaver{}

	var savedMetadata *store.Metadata
	version := 1
	ms.On("Save", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		savedMetadata = args.Get(1).(*store.Metadata)
		savedMetadata.Version = 1
	}).Return(nil)

	s := NewDefaultService(sm, fs, ms, format, quality)
	resp, err := s.MakeShotAndSave(context.Background(), url)
	require.NoError(t, err)
	require.Equal(t, resp, *savedMetadata)
	require.Equal(t, url, resp.Url)
	require.Equal(t, format, resp.Format)
	require.Equal(t, quality, resp.Quality)
	require.Equal(t, version, resp.Version)
	sm.AssertExpectations(t)
	fs.AssertExpectations(t)
	ms.AssertExpectations(t)
}
