package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	globalApiTemplateUrl string = "https://api.cloud.zilliz.com/v2"
	cnApiTemplateUrl     string = "https://api.cloud.zilliz.com.cn/v2"
)

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	apiKey     string
	RegionId   string
	baseUrl    string
	userAgent  string
	HttpClient HttpClient

	logger         *LoggerWrapper
	traceId        string
	logHttpTraffic bool
}

var (
	errApiKeyRequired   = fmt.Errorf("ApiKey is required")
	errRegionIdRequired = fmt.Errorf("RegionId is required")
)

func checkCloudRegionId(c *Client) func() error {
	return func() error {
		if c.RegionId == "" {
			return errRegionIdRequired
		}
		return nil
	}
}

func checkApiKey(c *Client) func() error {
	return func() error {
		if c.apiKey == "" {
			return errApiKeyRequired
		}
		return nil
	}
}

func (client *Client) Clone(opts ...Option) (*Client, error) {
	clone := func(c Client) *Client {
		return &c
	}

	c := clone(*client)

	for _, opt := range opts {
		opt(c)
	}

	applyOverride(c)

	if err := validate(c); err != nil {
		return nil, err
	}
	c.traceId = generateShortID()
	c.logger.logger.SetPrefix(fmt.Sprintf("[%s] ", c.traceId))

	return c, nil
}

// Cluster always use the caller's function name as the traceId.
func (client *Client) cluster(connectAddress string) (*Client, error) {
	c, err := client.Clone()
	if err != nil {
		return nil, err
	}
	c.baseUrl = connectAddress
	c.traceId = generateShortID()
	fn := getFrame(2)
	name := fn.Function
	if strings.Contains(name, "(") {
		name = strings.Split(name, "(")[1]
	}
	name = "(" + name
	c.logger.logger.SetPrefix(fmt.Sprintf("[%s] ", name))
	// TODO another validate

	return c, nil
}

func validate(c *Client) error {
	checkFns := []func() error{
		// checkCloudRegionId(c),
		checkApiKey(c),
	}
	for _, fn := range checkFns {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

func applyOverride(c *Client) {
	overrides := []Option{
		OverrideBaseUrl(),
	}
	for _, opt := range overrides {
		opt(c)
	}
}

func applyDefaults(c *Client) {
	defaultOptions := []Option{
		WithDefaultClient(),
		WithDefaultBaseUrl(),
		WithDefaultUserAgent(),
		WithDefaultTraceID(),
		WithDefaultHttpTrafficLogging(),
		WithDefaultLogger(),
	}
	for _, opt := range defaultOptions {
		opt(c)
	}
}

func NewClient(opts ...Option) (*Client, error) {

	// create a new client with options
	c := new(Client)
	for _, opt := range opts {
		opt(c)
	}

	applyDefaults(c)

	if err := validate(c); err != nil {
		return nil, err
	}

	return c, nil
}

type Option func(*Client)

func WithHTTPClient(client HttpClient) Option {
	return func(c *Client) {
		c.HttpClient = client
	}
}

func WithBaseUrl(baseUrl string) Option {
	return func(c *Client) {
		c.baseUrl = baseUrl
	}
}

func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

func WithCloudRegionId(cloudRegionId string) Option {
	return func(c *Client) {
		c.RegionId = cloudRegionId
	}
}

func WithApiKey(apiKey string) Option {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

func WithHostAddress(address string) Option {
	return func(c *Client) {
		c.baseUrl = address
	}
}

func WithLogger(logger *log.Logger) Option {
	return func(c *Client) {
		c.logger = NewLoggerWrapper(logger)
	}
}

func WithHttpTrafficLogging(enabled bool) Option {
	return func(c *Client) {
		c.logHttpTraffic = enabled
	}
}

func WithDefaultBaseUrl() Option {
	return func(c *Client) {
		if c.baseUrl == "" {
			c.baseUrl = globalApiTemplateUrl
		}
	}
}

func OverrideBaseUrl() Option {
	return func(c *Client) {
		if c.RegionId != "" {
			c.baseUrl = BaseUrlFrom(c.RegionId)
		}
	}

}

func WithDefaultClient() Option {
	return func(c *Client) {
		if c.HttpClient == nil {
			c.HttpClient = &http.Client{}
		}
	}
}

func WithDefaultUserAgent() Option {
	return func(c *Client) {
		if c.userAgent == "" {
			c.userAgent = "zilliztech/terraform-provider-zillizcloud"
		}
	}
}

func WithDefaultLogger() Option {
	return func(c *Client) {
		if c.logger == nil {
			if c.logHttpTraffic {
				f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					// fallback to stderr
					c.logger = NewLoggerWrapper(log.New(os.Stderr, fmt.Sprintf("[%s] ", c.traceId), log.LstdFlags))
				} else {
					c.logger = NewLoggerWrapper(log.New(f, fmt.Sprintf("[%s] ", c.traceId), log.LstdFlags))
				}
			} else {
				c.logger = NewLoggerWrapper(log.New(io.Discard, "", 0))
			}
		}
	}
}

func WithDefaultTraceID() Option {
	return func(c *Client) {
		if c.traceId == "" {
			c.traceId = generateShortID()
		}
	}
}

func WithDefaultHttpTrafficLogging() Option {
	return func(c *Client) {
		c.logHttpTraffic = os.Getenv("ZILLIZCLOUD_DEBUG") == "true"
	}
}

type zillizResponse[T any] struct {
	Error
	Data T `json:"data"`
}

type zillizPage struct {
	Count       int `json:"count"`
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
}

func (c *Client) do(method string, path string, body interface{}, result interface{}) error {

	u, err := c.url(path)
	if err != nil {
		return err
	}

	req, err := c.newRequest(method, u, body)
	if err != nil {
		return err
	}
	return c.doRequest(req, result)
}

func (c *Client) newRequest(method string, u *url.URL, body interface{}) (*http.Request, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	// req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer ${TOKEN}")
	command, _ := RequestToCurl(req)
	c.logger.Debugf("%v", command)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	return req, nil
}

func (c *Client) doRequest(req *http.Request, v any) error {
	// Log the request if HTTP traffic logging is enabled
	if c.logHttpTraffic {
		c.logger.LogRequest(req)
	}

	res, err := c.HttpClient.Do(req)
	if err != nil {
		if c.logHttpTraffic {
			c.logger.Errorf("HTTP request failed: %v", err)
		}
		return err
	}

	defer res.Body.Close()

	// Read the response body first so we can log it
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		if c.logHttpTraffic {
			c.logger.Errorf("Failed to read response body: %v", err)
		}
		return err
	}

	// Log the response if HTTP traffic logging is enabled
	if c.logHttpTraffic {
		c.logger.LogResponse(res, bodyBytes)
	}
	requestId := res.Header.Get("requestid")

	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("http status code: %d, error: %w, requestId: %s", res.StatusCode, parseError(bytes.NewReader(bodyBytes)), requestId)
	}

	return c.decodeResponse(bytes.NewReader(bodyBytes), requestId, v)
}

func parseError(body io.Reader) error {

	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	var e Error
	err = json.Unmarshal(b, &e)
	if err != nil {
		return err
	}

	return e
}

func (c *Client) decodeResponse(body io.Reader, requestId string, v any) error {
	if v == nil {
		return nil
	}
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	var apierr Error
	err = json.Unmarshal(b, &apierr)
	// if the error code is 0 or 200, it means the request is successful
	// otherwise, it means the request is failed
	if err == nil && apierr.Code != 200 && apierr.Code != 0 {
		apierr.RequestId = requestId
		return &apierr
	}
	err = json.Unmarshal(b, v)
	return err
}

func (c *Client) url(path string) (*url.URL, error) {
	return url.Parse(fmt.Sprintf("%s/%s", c.baseUrl, path))
}

func (c *Client) Log() *LoggerWrapper {
	return c.logger
}
