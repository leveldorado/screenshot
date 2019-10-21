package screenshotctl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"

	"github.com/leveldorado/screenshot/api"
)

type FlagOptions struct {
	Backend string `short:"b"  long:"backend" description:"address and port of screenshot backend api" default:"localhost:9000" env:"SCREENSHOT_BACKEND"`
	URLs    string `short:"u" long:"urls" description:"list of urls for screenshoting separated by ;"`
	File    string `short:"f" long:"file" description:"path to file with list of urls"`
}

const (
	urlSeparatorSign = ";"
)

func (f FlagOptions) ExtractURLs() ([]string, error) {
	if f.File != "" {
		data, err := ioutil.ReadFile(f.File)
		if err != nil {
			return nil, fmt.Errorf(`failed to read urls file: [path: %s, error: %w]`, f.File, err)
		}
		f.URLs = string(data)
	}
	if f.URLs == "" {
		return nil, fmt.Errorf(`please specify urls via --urls or file with urls -f`)
	}
	parts := strings.Split(f.URLs, urlSeparatorSign)
	var res []string
	for _, el := range parts {
		el = strings.TrimSpace(el)
		if el == "" {
			continue
		}
		res = append(res, el)
	}
	if len(res) == 0 {
		return nil, fmt.Errorf(`no urls specified: [urls: %v]`, f.URLs)
	}
	return res, nil
}

func ParseFlags(args []string) (FlagOptions, error) {
	var f FlagOptions
	_, err := flags.ParseArgs(&f, args)
	return f, err
}

type Command struct {
	cl         *http.Client
	serverAddr string
}

const (
	defaultRequestTimeout = 30 * time.Second
)

func NewCommand(serverAddr string) *Command {
	return &Command{cl: &http.Client{Timeout: defaultRequestTimeout}, serverAddr: serverAddr}
}

func (c *Command) MakeScreenShotsAndPrintResult(urls []string) error {
	buff := &bytes.Buffer{}
	if err := json.NewEncoder(buff).Encode(api.MakeShotsRequest{URLs: urls}); err != nil {
		return fmt.Errorf(`failed to encode request: [urls: %+v, errror: %w]`, urls, err)
	}
	req, err := http.NewRequest(http.MethodPost, api.ScreenshotPath, buff)
	if err != nil {
		return fmt.Errorf(`failed to request request: [path: %s, method: %s, error: %w]`, api.ScreenshotVersionsPath, http.MethodPost, err)
	}
	req.Header.Set("Content-type", "application/json")
	resp, err := c.cl.Do(req)
	if err != nil {
		return fmt.Errorf(`failed to do request: [url: %s, error: %w]`, req.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf(`received not successfull response code: [code: %s, response: %s]`, resp.Status, data)
	}
	var list []api.ResponseItem
	if err = json.NewDecoder(resp.Body).Decode(&resp); err != nil {
		return fmt.Errorf(`failed to decode response: [error: %w]`, err)
	}
	for _, el := range list {
		msg := fmt.Sprintf(`screenshot done for %s. you may fetch image by %s%s?url=%s`,
			el.URL, c.serverAddr, api.ScreenshotPath, el.URL)
		if !el.Success {
			msg = fmt.Sprintf(`screenshot failed for %s with error %s`, el.URL, el.Error)
		}
		log.Println(msg)
	}
	return nil
}
