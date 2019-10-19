package capture

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDefaultService_MakeShot(t *testing.T) {
	s := &DefaultService{}
	require.NoError(t, s.MakeShot(context.Background(), "https://www.google.com/search?q=spread&oq=spread+&aqs=chrome..69i57j0l7.20017j0j9&sourceid=chrome&ie=UTF-8"))
}
