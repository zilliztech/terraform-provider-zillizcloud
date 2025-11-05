package client

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

var (
	apiKey       string
	update       bool
	pollInterval int
	env          string
)

func init() {
	flag.StringVar(&apiKey, "key", "", "Your TEST secret key for the zilliz cloud API. If present, integration tests will be run using this key.")
	flag.BoolVar(&update, "update", false, "Set this flag to update the responses used in local tests. This requires that the key flag is set so that we can interact with the zilliz cloud API.")
	flag.StringVar(&env, "env", "prod", "The environment to run the tests in. Can be 'uat' or 'prod'.")
	// flag.IntVar(&pollInterval, "poll-interval", 60, "The interval in seconds to poll the zilliz cloud API for the status of a resource.")
}

func zillizClient[T any](t *testing.T) (*Client, func()) {
	tearDowns := make([]func(), 0)
	options := make([]Option, 0)
	key := "gibberish_key"

	// when apiKey is empty, we need to create a test server, read the response from the file, and return the response
	if apiKey == "" {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			data := readResponse[T](t)
			w.WriteHeader(data.StatusCode)
			_, err := w.Write(data.Body)
			if err != nil {
				t.Fatalf("failed to write response body: %v", err)
			}
		})
		server := httptest.NewServer(mux)

		tearDowns = append(tearDowns, server.Close)

		options = append(options, WithBaseUrl(server.URL))
	} else {
		options = append(options, WithHostAddress(
			func() string {
				if env == "uat" {
					return "https://api.cloud-uat3.zilliz.com/v2"
				}
				return "https://api.cloud.zilliz.com/v2"
			}(),
		))
		key = apiKey
	}

	client := &recorderClient{t: t}
	options = append(options,
		WithApiKey(key),
		WithCloudRegionId("gcp-us-west1"),
		WithHTTPClient(client))
	c, err := NewClient(
		options...,
	)

	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if apiKey != "" && update {
		tearDowns = append(tearDowns, func() {
			recordResponse(t, client.responses)
		})
	}

	return c, func() {
		for _, fn := range tearDowns {
			fn()
		}
	}
}

func responsePath(t *testing.T) string {
	return filepath.Join("testdata", filepath.FromSlash(fmt.Sprintf("%s.json", t.Name())))
}
func recordResponse(t *testing.T, resp any) {
	path := responsePath(t)
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		t.Fatalf("failed to create the response dir: %s. err = %v", filepath.Dir(path), err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create the response file: %s. err = %v", path, err)
	}
	defer f.Close()
	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal JSON for response file: %s. err = %v", path, err)
	}
	_, err = f.Write(jsonBytes)
	if err != nil {
		t.Fatalf("failed to write json bytes for response file: %s. err = %v", path, err)
	}
}

func readResponse[T any](t *testing.T) response {
	var resp response
	path := responsePath(t)
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open the response file: %s. err = %v", path, err)
	}
	defer f.Close()
	jsonBytes, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("failed to read the response file: %s. err = %v", path, err)
	}
	err = json.Unmarshal(jsonBytes, &resp)
	if err != nil {
		t.Fatalf("failed to json unmarshal the response file: %s. err = %v", path, err)
	}
	return resp
}

type response struct {
	StatusCode int    `json:"status_code"`
	Body       []byte `json:"body"`
}

type recorderClient struct {
	t         *testing.T
	responses response
}

func (rc *recorderClient) Do(req *http.Request) (*http.Response, error) {
	httpClient := &http.Client{}
	res, err := httpClient.Do(req)
	if err != nil {
		rc.t.Fatalf("http request failed. err = %v", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		rc.t.Fatalf("failed to read the response body. err = %v", err)
	}
	rc.responses = response{
		StatusCode: res.StatusCode,
		Body:       body,
	}
	res.Body = io.NopCloser(bytes.NewReader(body))
	return res, err
}
