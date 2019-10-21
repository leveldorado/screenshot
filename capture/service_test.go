package capture

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
)

type mockShotMaker struct {
	mock.Mock
}

func (m *mockShotMaker) MakeShot(ctx context.Context, url, format string, quality int) (io.Reader, error) {
	args := m.Called(ctx, url, format, quality)
	return args.Get(0).(io.Reader), args.Error(1)
}

func TestDefaultService_MakeShot(t *testing.T) {
	fmt.Println(len(strings.Split("", ";")))
}
