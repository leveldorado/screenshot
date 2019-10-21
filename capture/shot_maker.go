package capture

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
)

type ChromeShotMaker struct {
	addr string
}

func NewChromeShotMaker(addr string) *ChromeShotMaker {
	return &ChromeShotMaker{addr: addr}
}

func (c *ChromeShotMaker) buildClient(ctx context.Context) (*cdp.Client, func(), error) {
	devt := devtool.New(c.addr)
	pt, err := devt.Create(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf(`failed to create page target: [chrome_address: %s, error: %w]`, c.addr, err)
	}
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		devt.Close(ctx, pt)
		return nil, nil, fmt.Errorf(`failed to dial web socket debugger url: [url: %s, error: %w]`, pt.WebSocketDebuggerURL, err)
	}
	return cdp.NewClient(conn), func() {
		conn.Close()
		devt.Close(ctx, pt)
	}, nil
}

func navigateToPage(ctx context.Context, cl *cdp.Client, url string) error {
	frameStopedEventClient, err := cl.Page.FrameStoppedLoading(ctx)
	if err != nil {
		return fmt.Errorf(`failed to create frame stopped event client: [error: %w]`, err)
	}
	if err = cl.Page.Enable(ctx); err != nil {
		return fmt.Errorf(`failed to enable page domain notification: [error: %w]`, err)
	}
	_, err = cl.Page.Navigate(ctx, page.NewNavigateArgs(url))
	if err != nil {
		return fmt.Errorf(`failed navigate site url: [url: %s, error: %w]`, url, err)
	}
	_, err = frameStopedEventClient.Recv()
	if err != nil {
		return fmt.Errorf(`failed to receive frame stopped event: [error: %w]`, err)
	}
	return nil
}

func (c *ChromeShotMaker) MakeShot(ctx context.Context, url, format string, quality int) (io.Reader, error) {
	cl, close, err := c.buildClient(ctx)
	if err != nil {
		return nil, fmt.Errorf(`failed to build client: [error: %w]`, err)
	}
	defer close()
	if err = navigateToPage(ctx, cl, url); err != nil {
		return nil, fmt.Errorf(`failed to navigate to page: [error: %s]`, err)
	}
	screenshot, err := cl.Page.CaptureScreenshot(ctx, page.NewCaptureScreenshotArgs().
		SetFormat(format).
		SetQuality(quality))
	if err != nil {
		return nil, fmt.Errorf(`failed to capture screenshot [url: %s, format: %s, quality: %d, error: %w]`,
			url, format, quality, err)
	}
	buff := bytes.NewBuffer(screenshot.Data)
	return buff, nil
}
