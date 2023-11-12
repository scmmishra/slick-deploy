package caddy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/scmmishra/slick-deploy/internal/config"
)

type Config struct {
	// Add fields here according to the Caddy config you want to set up.
	// This is a basic example.
	Apps map[string]interface{} `json:"apps"`
}

// DefaultConfig returns a basic Caddy configuration.
func DefaultConfig() Config {
	return Config{
		Apps: map[string]interface{}{
			"http": map[string]interface{}{
				"servers": map[string]interface{}{
					"srv0": map[string]interface{}{
						"listen": []string{":80"},
						"routes": []map[string]interface{}{
							{
								"handle": []map[string]interface{}{
									{
										"handler": "static_response",
										"body":    "Hello, world!",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

type CaddyClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewCaddyClient(baseURL string) *CaddyClient {
	return &CaddyClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

func (c *CaddyClient) GetCurrentConfig() (string, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/config/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyStr := string(body)

	if bodyStr == "null\n" {
		return "", nil
	}

	return bodyStr, nil
}

func (c *CaddyClient) BootStrapCaddy() error {
	basicConfig := DefaultConfig()
	configJSON, err := json.Marshal(basicConfig)

	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	// Send request to Caddy to update config
	req, err := http.NewRequest("POST", c.BaseURL+"/load", bytes.NewBuffer(configJSON))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request to Caddy: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response from Caddy: %s", resp.Status)
	}

	return nil
}

func SetupCaddy(port int, cfg config.DeploymentConfig) error {
	client := NewCaddyClient(cfg.Caddy.AdminAPI)

	caddyConfig, err := client.GetCurrentConfig()
	if err != nil {
		return err
	}

	if caddyConfig == "" {
		fmt.Println("Empty Caddy config, bootstrapping...")
		err = client.BootStrapCaddy()
		if err != nil {
			return err
		}
	}

	return nil
}
