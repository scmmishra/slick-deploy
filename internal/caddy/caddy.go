package caddy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/scmmishra/slick-deploy/internal/config"
)

type CaddyConfig struct {
	Apps Apps `json:"apps"`
}

type Apps struct {
	HTTP HTTP `json:"http"`
}

type HTTP struct {
	Servers map[string]Server `json:"servers"`
}

type Server struct {
	Listen   []string `json:"listen"`
	Routes   []Route  `json:"routes"`
	Terminal bool     `json:"terminal"`
}

type Route struct {
	Handle   []Handle `json:"handle"`
	Match    []Match  `json:"match"`
	Terminal bool     `json:"terminal"`
}

type Handle struct {
	Handler   string     `json:"handler"`
	Routes    []Route    `json:"routes,omitempty"`
	Upstreams []Upstream `json:"upstreams,omitempty"` // Adding this line
}

type Match struct {
	Host []string `json:"host,omitempty"`
	Path []string `json:"path,omitempty"`
}

type Upstream struct {
	Dial string `json:"dial"`
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

func DefaultConfig() CaddyConfig {
	return CaddyConfig{
		Apps: Apps{
			HTTP: HTTP{
				Servers: map[string]Server{
					"slick-server": {},
				},
			},
		},
	}
}

func (c *CaddyConfig) AddReverseProxy() {
}

func (c *CaddyClient) GetCurrentConfig() (*CaddyConfig, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/config/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bodyStr := string(body)

	if bodyStr == "null\n" {
		return nil, nil
	}

	var config CaddyConfig
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
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

	if caddyConfig == nil {
		fmt.Println("Empty Caddy config, bootstrapping...")
		err = client.BootStrapCaddy()
		if err != nil {
			return err
		}

		// fetch the config again
		caddyConfig, _ = client.GetCurrentConfig()
	}

	// add the reverse proxy

	fmt.Println("Current Caddy config")
	fmt.Println(caddyConfig.Apps.HTTP)

	return nil
}
