package bitmarksdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
)

// Nework - it indicates what network we will connect
type Network string

// Config - struct that initialize the API client connection
type Config struct {
	Network    Network
	HTTPClient *http.Client
	APIToken   string
}

var (
	config    *Config
	apiClient *BackendImplementation
)

const (
	// Livenet - Defines the network type live
	Livenet = Network("livenet")
	// Testnet - Defines the network type test
	Testnet = Network("testnet")
)

// APIError - struct that holds the API errors
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
}

func (ae *APIError) Error() string {
	return fmt.Sprintf("[%d] message: %s reason: %s", ae.Code, ae.Message, ae.Reason)
}

// Init - SDK initialization
func Init(cfg *Config) {
	config = cfg
	switch cfg.Network {
	case Livenet:
		apiClient = &BackendImplementation{
			HTTPClient:   cfg.HTTPClient,
			URLAuthority: "https://api.bitmark.com",
			APIToken:     cfg.APIToken,
		}
	case Testnet:
		apiClient = &BackendImplementation{
			HTTPClient:   cfg.HTTPClient,
			URLAuthority: "https://api.test.bitmark.com",
			APIToken:     cfg.APIToken,
		}
	}
}

// GetNetwork - returns the network used
func GetNetwork() Network {
	return config.Network
}

// BackendImplementation - structure used by the API client
type BackendImplementation struct {
	HTTPClient        *http.Client
	URLAuthority      string
	APIToken          string
	MaxNetworkRetries int
}

// GetAPIClient - returns the API client
func GetAPIClient() *BackendImplementation {
	return apiClient
}

// NewRequest - returns a new Request given a method, URL and body
func (s *BackendImplementation) NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	url := s.URLAuthority + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("api-token", s.APIToken)
	// TODO: workaround for gateway proxy
	req.Header.Add("Accept-Encoding", "*")
	req.Header.Add("User-Agent", fmt.Sprintf("%s, %s, %s", "bitmark-sdk-go", runtime.GOOS, runtime.Version()))

	return req, nil
}

// Do - sends an HTTP request
func (s *BackendImplementation) Do(req *http.Request, v interface{}) error {
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var aerr APIError
		if err := json.NewDecoder(resp.Body).Decode(&aerr); err != nil {
			return errors.New("unexpected api response")
		}

		return &aerr
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}
