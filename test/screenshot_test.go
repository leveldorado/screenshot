package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/leveldorado/screenshot/store"

	"github.com/labstack/echo"
	"github.com/leveldorado/screenshot/api"
	"github.com/stretchr/testify/require"
)

const (
	testAPIAddress = "SCREENSHOT_TEST_API"
)

func TestScreenshotAPI(t *testing.T) {
	address := os.Getenv(testAPIAddress)
	form := api.MakeShotsRequest{URLs: []string{"https://stackoverflow.com", "https://github.com"}}
	buf := &bytes.Buffer{}
	require.NoError(t, json.NewEncoder(buf).Encode(form))
	cl := http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(`%s%s`, address, api.ScreenshotPath), buf)
	require.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp, err := cl.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var response []api.ResponseItem
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	resp.Body.Close()
	for _, url := range form.URLs {
		require.Contains(t, response, api.ResponseItem{URL: url, Success: true})

		req, err = http.NewRequest(http.MethodGet, fmt.Sprintf(`%s%s?url=%s`, address, api.ScreenshotPath, url), nil)
		require.NoError(t, err)
		resp, err = cl.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		require.NoError(t, err)
		// TODO design how to test image
		require.NotEmpty(t, data)

		req, err = http.NewRequest(http.MethodGet, fmt.Sprintf(`%s%s?url=%s`, address, api.ScreenshotVersionsPath, url), nil)
		require.NoError(t, err)
		resp, err = cl.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var versions []store.Metadata
		require.GreaterOrEqual(t, 1, len(versions))
	}

}
