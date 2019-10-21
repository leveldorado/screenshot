package capture

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const ScreenshotTestChromeAddressEnvVariable = "SCREENSHOT_TEST_CHROME"

func TestChromeShotMaker_MakeShot(t *testing.T) {
	address := os.Getenv(ScreenshotTestChromeAddressEnvVariable)
	sm := NewChromeShotMaker(address)
	go func() {
		image, err := sm.MakeShot(context.Background(), "http://facebook.com", "jpeg", 80)
		require.NoError(t, err)
		// do not know how to automatically test screenshot generation
		require.NotNil(t, image)

		data, err := ioutil.ReadAll(image)
		require.NoError(t, err)
		require.NoError(t, ioutil.WriteFile("image.jpeg", data, os.ModePerm))
	}()

	go func() {
		image, err := sm.MakeShot(context.Background(), "http://google.com", "jpeg", 80)
		require.NoError(t, err)
		// do not know how to automatically test screenshot generation
		require.NotNil(t, image)

		data, err := ioutil.ReadAll(image)
		require.NoError(t, err)
		require.NoError(t, ioutil.WriteFile("image2.jpeg", data, os.ModePerm))
	}()
	<-time.After(10 * time.Second)
}
