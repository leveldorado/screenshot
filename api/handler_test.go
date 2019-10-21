package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/leveldorado/screenshot/store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockService struct {
	mock.Mock
}

func (m *mockService) MakeShots(ctx context.Context, urls []string) []ResponseItem {
	return m.Called(ctx, urls).Get(0).([]ResponseItem)
}
func (m *mockService) GetScreenshot(ctx context.Context, url string, version int) (file io.ReadCloser, contentType string, err error) {
	args := m.Called(ctx, url, version)
	return args.Get(0).(io.ReadCloser), args.Get(1).(string), args.Error(2)
}
func (m *mockService) GetScreenshotVersions(ctx context.Context, url string) ([]store.Metadata, error) {
	args := m.Called(ctx, url)
	return args.Get(0).([]store.Metadata), args.Error(1)
}

func TestHTTPHandlerGetScreenshotVersions(t *testing.T) {
	s := &mockService{}
	url := uuid.New().String()
	response := []store.Metadata{{ID: uuid.New().String(), Url: url, Format: "jpeg", Version: 13}}
	s.On("GetScreenshotVersions", mock.Anything, url).Return(response, nil)
	h := NewHTTPHandler(s, "address")
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf(`%s?url=%s`, ScreenshotVersionsPath, url), nil)
	resp := httptest.NewRecorder()
	ctx := h.server.NewContext(req, resp)
	require.NoError(t, h.getScreenshotVersions(ctx))
	require.Equal(t, http.StatusOK, resp.Code)
	var actualResponse []store.Metadata
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &actualResponse))
	require.Equal(t, response, actualResponse)
	s.AssertExpectations(t)
}

func TestHTTPHandlerGetScreenshot(t *testing.T) {
	s := &mockService{}
	url := uuid.New().String()
	version := 2
	data := uuid.New().String()
	file := ioutil.NopCloser(strings.NewReader(data))
	contentType := "image/jpeg"
	s.On("GetScreenshot", mock.Anything, url, version).Return(file, contentType, nil)
	h := NewHTTPHandler(s, "address")
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf(`%s?url=%s&version=%d`, ScreenshotPath, url, version), nil)
	resp := httptest.NewRecorder()
	ctx := h.server.NewContext(req, resp)
	require.NoError(t, h.getScreenshot(ctx))
	require.Equal(t, http.StatusOK, resp.Code)
	require.Equal(t, contentType, resp.Header().Get(echo.HeaderContentType))
	require.Equal(t, data, resp.Body.String())
	s.AssertExpectations(t)
}

func TestHTTPHandlerMakeShots(t *testing.T) {
	s := &mockService{}
	urls := []string{uuid.New().String(), uuid.New().String()}
	response := []ResponseItem{{URL: urls[0], Success: true}, {URL: urls[1], Error: "some error"}}
	s.On("MakeShots", mock.Anything, urls).Return(response)

	h := NewHTTPHandler(s, "address")
	data, err := json.Marshal(MakeShotsRequest{URLs: urls})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, ScreenshotPath, bytes.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp := httptest.NewRecorder()
	ctx := h.server.NewContext(req, resp)
	require.NoError(t, h.makeShots(ctx))
	require.Equal(t, http.StatusOK, resp.Code)
	var actualResponse []ResponseItem
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &actualResponse))
	require.Equal(t, response, actualResponse)
	s.AssertExpectations(t)
}
