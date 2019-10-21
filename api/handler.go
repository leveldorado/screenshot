package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/middleware"

	"github.com/labstack/echo"

	"github.com/leveldorado/screenshot/store"
)

type service interface {
	MakeShots(ctx context.Context, urls []string) []ResponseItem
	GetScreenshot(ctx context.Context, url string, version int) (file io.ReadCloser, contentType string, err error)
	GetScreenshotVersions(ctx context.Context, url string) ([]store.Metadata, error)
}

type httpServer interface {
	Start(addr string) error
	Shutdown(ctx context.Context) error
}

type HTTPHandler struct {
	server  *echo.Echo
	s       service
	address string
}

func NewHTTPHandler(s service, addr string) *HTTPHandler {
	e := echo.New()
	e.Use(middleware.Recover())
	h := &HTTPHandler{s: s, address: addr, server: e}
	h.registerEndpoints()
	return h
}

func (h *HTTPHandler) Run(ctx context.Context) error {
	go func() {
		if err := h.server.Start(h.address); err != nil {
			log.Println(fmt.Sprintf(`failed to start http server on address %s with error: %s`, h.address, err))
		}
		log.Println("http server shutdown")
	}()
	return nil
}

func (h *HTTPHandler) Stop(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

const (
	ScreenshotPath         = "/api/v1/screenshot"
	ScreenshotVersionsPath = "/api/v1/screenshot/versions"
)

func (h *HTTPHandler) registerEndpoints() {
	h.server.POST(ScreenshotPath, h.makeShots)
	h.server.GET(ScreenshotPath, h.getScreenshot)
	h.server.GET(ScreenshotVersionsPath, h.getScreenshotVersions)
}

type MakeShotsRequest struct {
	URLs []string `json:"urls"`
}

type ErrorResponse struct {
	Message string
}

func (h HTTPHandler) makeShots(ctx echo.Context) error {
	req := MakeShotsRequest{}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
	}
	response := h.s.MakeShots(ctx.Request().Context(), req.URLs)
	return ctx.JSON(http.StatusOK, response)
}

func (h HTTPHandler) getScreenshot(ctx echo.Context) error {
	url := ctx.QueryParam("url")
	if url == "" {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "missing required query parameter url"})
	}
	var version int
	versionParam := ctx.QueryParam("version")
	if versionParam != "" {
		v, err := strconv.ParseInt(versionParam, 10, 64)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: fmt.Sprintf(`invalid parameter version: [version: %s, error: %s]`, versionParam, err)})
		}
		version = int(v)
	}
	file, contentType, err := h.s.GetScreenshot(ctx.Request().Context(), url, version)
	if errors.As(err, &store.ErrNotFound{}) {
		return ctx.JSON(http.StatusNotFound, ErrorResponse{Message: "screenshot not found"})
	}
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	defer file.Close()
	return ctx.Stream(http.StatusOK, contentType, file)
}

func (h HTTPHandler) getScreenshotVersions(ctx echo.Context) error {
	url := ctx.QueryParam("url")
	if url == "" {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "missing required query parameter url"})
	}
	resp, err := h.s.GetScreenshotVersions(ctx.Request().Context(), url)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	return ctx.JSONPretty(http.StatusOK, resp, " ")
}
