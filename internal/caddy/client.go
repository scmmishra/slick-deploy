package caddy

import (
	"bytes"
	"fmt"
	"net/http"
)

type CaddyClientInterface interface {
	Load(caddyfile string) error
}

type CaddyClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

var NewCaddyClient = func(baseURL string) CaddyClientInterface {
	return &CaddyClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

// https://caddyserver.com/docs/api#post-load
//
//	curl "http://localhost:2019/load" \
//		-H "Content-Type: text/caddyfile" \
//		--data-binary @Caddyfile
func (cl *CaddyClient) Load(caddyfile string) error {
	req, err := http.NewRequest("POST", cl.BaseURL+"/load", bytes.NewBuffer([]byte(caddyfile)))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "text/caddyfile")
	resp, err := cl.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request to Caddy: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response from Caddy: %s", resp.Status)
	}

	return nil
}
