package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo"

	"github.com/leveldorado/screenshot/store"
)

type service interface {
	MakeShots(ctx context.Context, urls []string) []ResponseItem
	GetScreenshot(ctx context.Context, url string, version int) (file io.ReadCloser, contentType string, err error)
	GetScreenshotVersions(ctx context.Context, url string) ([]store.Metadata, error)
}

type HTTPHandler struct {
	s service
}

type MakeShotsRequest struct {
	URLs []string `json:"urls"`
}

type ErrorResponse struct {
	Message string
}

func (h HTTPHandler) makeShots(ctx echo.Context) {
	req := MakeShotsRequest{}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}
	response := h.s.MakeShots(ctx.Request().Context(), req.URLs)
	ctx.JSON(http.StatusOK, response)
}

func (h HTTPHandler) getScreenshot(ctx echo.Context) {
	url := ctx.QueryParam("url")
	if url == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "missing required query parameter url"})
		return
	}
	var version int
	versionParam := ctx.QueryParam("version")
	if versionParam != "" {
		v, err := strconv.ParseInt(versionParam, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: fmt.Sprintf(`invalid parameter version: [version: %s, error: %s]`, versionParam, err)})
			return
		}
		version = int(v)
	}
	file, contentType, err := h.s.GetScreenshot(ctx.Request().Context(), url, version)
	if errors.As(err, &store.ErrNotFound{}) {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Message: "screenshot not found"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}
	ctx.Stream(http.StatusOK, contentType, file)
	file.Close()
}

func (h HTTPHandler) getScreenshotVersions(ctx echo.Context) {
	url := ctx.QueryParam("url")
	if url == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "missing required query parameter url"})
		return
	}
	resp, err := h.s.GetScreenshotVersions(ctx.Request().Context(), url)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
