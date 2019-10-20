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

type HTTPHandler struct {
	server  *echo.Echo
	s       service
	address string
}

func (h *HTTPHandler) Run(ctx context.Context) error {
	h.server = echo.New()
	h.server.Use(middleware.Recover())
	h.registerEndpoints(h.server)
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

func (h *HTTPHandler) registerEndpoints(e *echo.Echo) {
	e.POST("/api/v1/screenshot", h.makeShots)
	e.GET("/api/v1/screenshot", h.getScreenshot)
	e.GET("/api/v1/screenshot/versions", h.getScreenshotVersions)
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
	return ctx.JSON(http.StatusOK, resp)
}
