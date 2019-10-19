package capture

import (
	"context"
	"fmt"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
)

type ChromeShotMaker struct {
	outputFormat  string
	outputQuality int
	conn          *rpcc.Conn
}

func NewChromeShotMaker(ctx context.Context, addr, format string, quality int) (*ChromeShotMaker, error) {
	devt := devtool.New(addr)
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(ctx)
		if err != nil {
			return nil, fmt.Errorf(`failed to create page target: [chrome_address: %s, error: %w]`, addr, err)
		}
	}
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return nil, fmt.Errorf(`failed to dial web socket debugger url: [url: %s, error: %w]`, pt.WebSocketDebuggerURL, err)
	}
	return &ChromeShotMaker{outputFormat: format, outputQuality: quality, conn: conn}, nil
}

func (c *ChromeShotMaker) MakeShot(ctx context.Context, url string) ([]byte, error) {
	cl := cdp.NewClient(c.conn)
	frameStopedEventClient, err := cl.Page.FrameStoppedLoading(ctx)
	if err != nil {
		return nil, fmt.Errorf(`failed to create frame stopped event client: [error: %w]`, err)
	}
	if err = cl.Page.Enable(ctx); err != nil {
		return nil, fmt.Errorf(`failed to enable page domain notification: [error: %w]`, err)
	}
	_, err = cl.Page.Navigate(ctx, page.NewNavigateArgs(url))
	if err != nil {
		return nil, fmt.Errorf(`failed navigate site url: [url: %s, error: %w]`, url, err)
	}
	_, err = frameStopedEventClient.Recv()
	if err != nil {
		return nil, fmt.Errorf(`failed to receive frame stopped event: [error: %w]`, err)
	}
	screenshot, err := cl.Page.CaptureScreenshot(ctx, page.NewCaptureScreenshotArgs().
		SetFormat(c.outputFormat).
		SetQuality(c.outputQuality))
	if err != nil {
		return nil, fmt.Errorf(`failed to capture screenshot [url: %s, format: %s, quality: %d, error: %w]`,
			url, c.outputFormat, c.outputQuality, err)
	}
	return screenshot.Data, nil
}
