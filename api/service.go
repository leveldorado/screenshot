package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/leveldorado/screenshot/capture"

	"github.com/leveldorado/screenshot/queue"
	"github.com/leveldorado/screenshot/store"
)

type fileGetter interface {
	Get(ctx context.Context, fileID string) (io.ReadCloser, error)
}

type metadataGetter interface {
	Get(ctx context.Context, url string, version int) (store.Metadata, error)
	GetAllVersions(ctx context.Context, url string) ([]store.Metadata, error)
}

type subscriberPublisher interface {
	Subscribe(ctx context.Context, topic string) (<-chan queue.Message, error)
	Publish(ctx context.Context, topic, reply string, data interface{}) error
}

type DefaultService struct {
	fg               fileGetter
	mg               metadataGetter
	q                subscriberPublisher
	waitReplyTimeout time.Duration
}

func NewDefaultService(fg fileGetter, mg metadataGetter, q subscriberPublisher, waitReplyTimeout time.Duration) *DefaultService {
	return &DefaultService{
		fg:               fg,
		mg:               mg,
		q:                q,
		waitReplyTimeout: waitReplyTimeout,
	}
}

type ResponseItem struct {
	URL     string `json:"url"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (s *DefaultService) MakeShots(ctx context.Context, urls []string) []ResponseItem {
	responsesChan := make(chan ResponseItem, len(urls))
	for _, u := range urls {
		go s.makeShot(ctx, u, responsesChan)
	}
	var responses []ResponseItem
	for i := 0; i < len(urls); i++ {
		responses = append(responses, <-responsesChan)
	}
	return responses
}

func (s *DefaultService) makeShot(ctx context.Context, url string, respChan chan<- ResponseItem) {
	req := capture.ShotRequest{URL: url}
	reply := uuid.New().String()
	if err := s.q.Publish(ctx, capture.ShotRequestTopic, reply, req); err != nil {
		respChan <- ResponseItem{URL: url, Error: fmt.Sprintf(`failed to publish shot request: [req: %+v, error: %s]`, req, err)}
		return
	}
	ctx, cancel := context.WithTimeout(ctx, s.waitReplyTimeout)
	sub, err := s.q.Subscribe(ctx, reply)
	if err != nil {
		respChan <- ResponseItem{URL: url, Error: fmt.Sprintf(`failed to subscribe shot response: [reply: %s, error: %s]`, reply, err)}
		return
	}
	msg := <-sub
	//empty message means channel has been closed
	if len(msg.Data) == 0 {
		respChan <- ResponseItem{URL: url, Error: `failed to receive shot response`}
		return
	}
	cancel()
	var resp capture.ShotResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		respChan <- ResponseItem{URL: url, Error: fmt.Sprintf(`failed to unmarshal shot response: [data: %s, error: %s]`, msg.Data, err)}
		return
	}
	respChan <- ResponseItem{URL: url, Success: true}
}

func (s *DefaultService) GetScreenshot(ctx context.Context, url string, version int) (file io.ReadCloser, contentType string, err error) {
	var m store.Metadata
	if version == 0 {
		versions, err := s.mg.GetAllVersions(ctx, url)
		if err != nil {
			return nil, "", fmt.Errorf(`failed to get screen shot versions: [url: %s, error: %w]`, url, err)
		}
		if len(versions) == 0 {
			return nil, "", store.ErrNotFound{}
		}
		sort.Sort(store.MetadataByVersionDesc(versions))
		m = versions[0]
	} else {
		m, err = s.mg.Get(ctx, url, version)
		if err != nil {
			return nil, "", fmt.Errorf(`failed to get screenshot metadata: [url: %s, version: %d, error: %w]`, url, version, err)
		}
	}
	file, err = s.fg.Get(ctx, m.FileID)
	if err != nil {
		return nil, "", fmt.Errorf(`failed to get file: [file_id: %s, error: %w]`, m.FileID, err)
	}
	return file, m.GetContentType(), nil
}

func (s *DefaultService) GetScreenshotVersions(ctx context.Context, url string) ([]store.Metadata, error) {
	versions, err := s.mg.GetAllVersions(ctx, url)
	if err != nil {
		return nil, fmt.Errorf(`failed to get screen shot versions: [url: %s, error: %w]`, url, err)
	}
	sort.Sort(store.MetadataByVersionDesc(versions))
	return versions, nil
}
