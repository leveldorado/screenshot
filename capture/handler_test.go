package capture

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/google/uuid"

	"github.com/leveldorado/screenshot/queue"

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

type mockSubscriberReplier struct {
	mock.Mock
}

func (m *mockSubscriberReplier) GroupSubscribe(ctx context.Context, topic, group string) (<-chan queue.Message, error) {
	args := m.Called(ctx, topic, group)
	return args.Get(0).(chan queue.Message), args.Error(1)
}

func (m *mockSubscriberReplier) Reply(ctx context.Context, reply string, data interface{}) error {
	return m.Called(ctx, reply, data).Error(0)
}

func TestQueueSubscriptionHandlerMakeShotAndSave(t *testing.T) {
	s := &mockService{}
	url := uuid.New().String()
	metadata := store.Metadata{ID: uuid.New().String(), Url: url, Format: "jpeg"}
	s.On("MakeShotAndSave", mock.Anything, url).Return(metadata, nil)

	resp := ShotResponse{Success: true, Metadata: metadata}
	req := ShotRequest{URL: url}
	reqData, err := json.Marshal(req)
	require.NoError(t, err)
	msg := queue.Message{Data: reqData, Reply: uuid.New().String()}
	msgChan := make(chan queue.Message)
	q := &mockSubscriberReplier{}
	q.On("GroupSubscribe", mock.Anything, ShotRequestTopic, subscriptionGroupCapture).Return(msgChan, nil)
	q.On("Reply", mock.Anything, msg.Reply, resp).Return(nil)
	h := NewQueueSubscriptionHandler(s, q, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, h.Run(ctx))
	msgChan <- msg
	<-time.After(time.Millisecond)
	s.AssertExpectations(t)
	q.AssertExpectations(t)
}
