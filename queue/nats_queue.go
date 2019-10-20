package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/go-nats"
)

type NATS struct {
	conn       *nats.Conn
	bufferSize int
}

func NewNATS(addr string, bufferSize int, timeout time.Duration) (*NATS, error) {
	conn, err := nats.Connect(addr, nats.Timeout(timeout))
	if err != nil {
		return nil, fmt.Errorf(`failed to connect to nats server: [addr: %s, timeout: %s]`, addr, timeout)
	}
	return &NATS{conn: conn, bufferSize: bufferSize}, nil
}

func (n *NATS) Publish(ctx context.Context, topic, reply string, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf(`failed to marshal data to json: [data: %+v, error: %w]`, data, err)
	}
	if err = n.conn.PublishRequest(topic, reply, bytes); err != nil {
		return fmt.Errorf(`failed to publish message: [topic: %s, reply: %s, data: %s, error: %w]`, topic, reply, bytes, err)
	}
	return nil
}

func (n *NATS) Reply(ctx context.Context, reply string, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf(`failed to marshal data to json: [data: %+v, error: %w]`, data, err)
	}
	if err = n.conn.Publish(reply, bytes); err != nil {
		return fmt.Errorf(`failed to publish reply message: [reply: %s, data: %s, error: %w]`, reply, bytes, err)
	}
	return nil
}

type Message struct {
	Data  []byte
	Reply string
}

func (n *NATS) GroupSubscribe(ctx context.Context, topic, group string) (<-chan Message, error) {
	sub, err := n.conn.QueueSubscribeSync(topic, group)
	if err != nil {
		return nil, fmt.Errorf(`failed to queue subscribe: [topic: %s, group: %s]`, topic, group)
	}
	return n.consumeSubscription(ctx, topic, sub)
}

func (n *NATS) Subscribe(ctx context.Context, topic string) (<-chan Message, error) {
	sub, err := n.conn.SubscribeSync(topic)
	if err != nil {
		return nil, fmt.Errorf(`failed to subscribe: [topic: %s]`, topic)
	}
	return n.consumeSubscription(ctx, topic, sub)
}

func (n *NATS) consumeSubscription(ctx context.Context, topic string, sub *nats.Subscription) (<-chan Message, error) {
	c := make(chan Message, n.bufferSize)
	go func() {
		for {
			msg, err := sub.NextMsgWithContext(ctx)
			if err != nil {
				log.Println(fmt.Sprintf(`failed get next message for topic %s: with error: %s`, topic, err))
				close(c)
				return
			}
			c <- Message{Data: msg.Data, Reply: msg.Reply}
		}
	}()
	return c, nil
}
