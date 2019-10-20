package capture

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/leveldorado/screenshot/queue"
	"github.com/leveldorado/screenshot/store"
)

type ShotResponse struct {
	Success  bool           `json:"success"`
	Metadata store.Metadata `json:"metadata"`
	Error    string         `json:"error"`
}

type ShotRequest struct {
	URL string `json:"url"`
}

const ShotRequestTopic = "shot_request"

type service interface {
	MakeShotAndSave(ctx context.Context, url string) (store.Metadata, error)
}

type subscriberReplier interface {
	GroupSubscribe(ctx context.Context, topic, group string) (<-chan queue.Message, error)
	Reply(ctx context.Context, reply string, data interface{}) error
}

type QueueSubscriptionHandler struct {
	s              service
	q              subscriberReplier
	requestTimeout time.Duration
}

const subscriptionGroupCapture = "capture"

func (h *QueueSubscriptionHandler) SubscribeTopics(ctx context.Context) error {
	if err := h.subscribeTopic(ctx, ShotRequestTopic, h.makeShotAndSave); err != nil {
		return fmt.Errorf(`failed to subscribe shot request topic: [error: %w]`, err)
	}
	return nil
}

type messageHandler func(ctx context.Context, msg []byte) interface{}

func (h *QueueSubscriptionHandler) subscribeTopic(ctx context.Context, topic string, mh messageHandler) error {
	sub, err := h.q.GroupSubscribe(ctx, topic, subscriptionGroupCapture)
	if err != nil {
		return fmt.Errorf(`failed to subscribe topic: [topic: %s, error: %w]`, topic, err)
	}
	go func() {
		for msg := range sub {
			//TODO handle context timeout
			msgCtx, _ := context.WithTimeout(context.Background(), h.requestTimeout)
			resp := mh(msgCtx, msg.Data)
			if err := h.q.Reply(msgCtx, msg.Reply, resp); err != nil {
				log.Println(fmt.Sprintf(`failed to publish reply: [topic: %s, reply: %s, resp: %+v, error: %s]`, topic, msg.Reply, resp, err))
			}
		}
	}()
	return nil
}

func (h *QueueSubscriptionHandler) makeShotAndSave(ctx context.Context, msg []byte) interface{} {
	var req ShotRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		return ShotResponse{Error: fmt.Sprintf(`failed to unmarshal shot request: [msg: %s, error: %s]`, msg, err)}
	}
	metadata, err := h.s.MakeShotAndSave(ctx, req.URL)
	if err != nil {
		return ShotResponse{Error: fmt.Sprintf(`failed to make shot and save: [url: %s, error: %s]`, req.URL, err)}
	}
	return ShotResponse{Success: true, Metadata: metadata}
}
