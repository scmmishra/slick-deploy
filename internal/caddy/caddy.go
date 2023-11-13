package caddy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/scmmishra/slick-deploy/internal/config"
)

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
