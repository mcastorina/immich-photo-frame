package immich

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
)

// Client provides a raw HTTP client for accessing the immich API. All requests
// will get rewritten to the API endpoint with authorization, so only the path
// is required for requests.
//
// Example:
//
// ```
// client := NewClientFromEnv()
// resp, err := client.Get("/users/me")
// ```
type Client struct {
	*http.Client
}

// Config holds configuration values for configuring the immich client.
//
// It is organized to take advantage of TOML parsing, however this package does
// not handle parsing and has no expectation on how it will be initialized.
type Config struct {
	// ImmichAPIEndpoint is the URL for accessing the immich API.
	ImmichAPIEndpoint string
	// ImmichAPIKey should ideally not be written to disk un-encrypted,
	// however, for ease of "deployment" I'm going to allow it.
	ImmichAPIKey string
}

// HydrateFromEnv overwrites any values in Config with their associated
// environment variable value. Environment variables take precedence.
func (c *Config) HydrateFromEnv() {
	if v, ok := os.LookupEnv("IMMICH_API_ENDPOINT"); ok {
		c.ImmichAPIEndpoint = v
	}
	if v, ok := os.LookupEnv("IMMICH_API_KEY"); ok {
		c.ImmichAPIKey = v
	}
}

// immichTransport is a custom http.Transport that rewrites the http.Request
// via transformF.
type immichTransport struct {
	transformF func(*http.Request)
}

func (i immichTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	i.transformF(req)
	return http.DefaultTransport.RoundTrip(req)
}

// NewClientFromEnv initializes a Client using the IMMICH_API_ENDPOINT and
// IMMICH_API_KEY environment variables.
func NewClientFromEnv() Client {
	conf := Config{}
	conf.HydrateFromEnv()
	return NewClient(conf)
}

// NewClient initializes a Client with the provided API endpoint and API key.
// Use [IsConnected] to check if the Client was properly configured.
func NewClient(conf Config) Client {
	// Canonicalize apiEndpoint.
	apiEndpointURI, _ := url.Parse(conf.ImmichAPIEndpoint)
	if apiEndpointURI.Path != "/api" {
		apiEndpointURI.Path = "/api"
	}

	// Build a custom http.Transport to set the API credentials and host.
	transport := immichTransport{
		transformF: func(r *http.Request) {
			// Add the API header credentials.
			r.Header.Add("X-API-Key", conf.ImmichAPIEndpoint)
			// Prefix the API endpoint in the new URL.
			immichAPI := *apiEndpointURI
			immichAPI.Path = path.Join(immichAPI.Path, r.URL.Path)
			r.URL = &immichAPI
		},
	}
	return Client{&http.Client{Transport: transport}}
}

// IsConnected performs a sanity check API request to /users/me to verify the
// Client is configured correctly and the immich server is responsive.
func (c Client) IsConnected() error {
	resp, err := c.Get("/users/me")
	if err != nil && err.Error() == `Get "/users/me": unsupported protocol scheme ""` {
		return errors.New("misconfigured client: missing immich endpoint")
	} else if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Check the response code.
	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("misconfigured client: invalid immich token")
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	// Check it's a JSON response.
	var m map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return err
	}
	return nil
}
