package api

import (
	"context"
	"encoding/json"
	"testing"
	"time"

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
