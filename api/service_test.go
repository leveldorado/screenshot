package api

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/leveldorado/screenshot/store"

	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
	"github.com/leveldorado/screenshot/capture"

	"github.com/leveldorado/screenshot/queue"
	"github.com/stretchr/testify/mock"
)

type mockSubscriberPublisher struct {
	mock.Mock
}

func (m *mockSubscriberPublisher) Subscribe(ctx context.Context, topic string) (<-chan queue.Message, error) {
	args := m.Called(ctx, topic)
	return args.Get(0).(chan queue.Message), args.Error(1)
}

func (m *mockSubscriberPublisher) Publish(ctx context.Context, topic, reply string, data interface{}) error {
	return m.Called(ctx, topic, reply, data).Error(0)
}

type mockMetadataGetter struct {
	mock.Mock
}

func (m *mockMetadataGetter) Get(ctx context.Context, url string, version int) (store.Metadata, error) {
	args := m.Called(ctx, url, version)
	return args.Get(0).(store.Metadata), args.Error(1)
}
func (m *mockMetadataGetter) GetAllVersions(ctx context.Context, url string) ([]store.Metadata, error) {
	args := m.Called(ctx, url)
	return args.Get(0).([]store.Metadata), args.Error(1)
}

type mockFileGetter struct {
	mock.Mock
}

func (m *mockFileGetter) Get(ctx context.Context, fileID string) (io.ReadCloser, error) {
	args := m.Called(ctx, fileID)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func TestDefaultService_GetScreenshot(t *testing.T) {
	mg := &mockMetadataGetter{}
	latest := store.Metadata{FileID: uuid.New().String(), Format: "jpeg", Version: 2}
	list := []store.Metadata{{FileID: uuid.New().String(), Format: "png", Version: 1}, latest}
	url := uuid.New().String()
	mg.On("GetAllVersions", mock.Anything, url).Return(list, nil)
	fg := &mockFileGetter{}
	file := ioutil.NopCloser(strings.NewReader(uuid.New().String()))
	fg.On("Get", mock.Anything, latest.FileID).Return(file, nil)
	s := NewDefaultService(fg, mg, nil, 0)
	respFile, contentType, err := s.GetScreenshot(context.Background(), url, 0)
	require.NoError(t, err)
	require.Equal(t, file, respFile)
	require.Equal(t, "image/jpeg", contentType)
	mg.AssertExpectations(t)
	fg.AssertExpectations(t)
}

func TestDefaultService_GetScreenshotVersions(t *testing.T) {
	mg := &mockMetadataGetter{}
	list := []store.Metadata{{FileID: uuid.New().String(), Format: "jpeg", Version: 2}, {FileID: uuid.New().String(), Format: "jpeg", Version: 1}}
	url := uuid.New().String()
	mg.On("GetAllVersions", mock.Anything, url).Return(list, nil)
	s := NewDefaultService(nil, mg, nil, 0)
	resp, err := s.GetScreenshotVersions(context.Background(), url)
	require.NoError(t, err)
	require.Equal(t, list, resp)
	mg.AssertExpectations(t)
}

func TestDefaultService_MakeShots(t *testing.T) {
	q := &mockSubscriberPublisher{}
	req := capture.ShotRequest{URL: uuid.New().String()}
	q.On("Publish", mock.Anything, capture.ShotRequestTopic, mock.Anything, req).Return(nil)
	msgChan := make(chan queue.Message, 1)
	captureResp := capture.ShotResponse{Success: true}
	data, err := json.Marshal(captureResp)
	require.NoError(t, err)
	msgChan <- queue.Message{Data: data}
	q.On("Subscribe", mock.Anything, mock.Anything).Return(msgChan, nil)
	s := NewDefaultService(nil, nil, q, time.Second)
	resp := s.MakeShots(context.Background(), []string{req.URL})
	require.Equal(t, []ResponseItem{{URL: req.URL, Success: true}}, resp)
	q.AssertExpectations(t)
}
