package caddy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/scmmishra/slick-deploy/internal/config"
)

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

type Rule struct {
	Match        string         `yaml:"match"`
	ReverseProxy []ReverseProxy `yaml:"reverse_proxy"`
}

type ReverseProxy struct {
	Path string `yaml:"path"`
	To   string `yaml:"to"`
}

type DataWithPort struct {
	Port int
}

// ConvertToCaddyfile translates the CaddyConfig struct to a Caddyfile string
func ConvertToCaddyfile(caddyCfg config.CaddyConfig, port int) (string, error) {
	var caddyfileBuilder strings.Builder

	for _, rule := range caddyCfg.Rules {
		caddyfileBuilder.WriteString(rule.Match)
		caddyfileBuilder.WriteString(" {\n")
		for _, proxy := range rule.ReverseProxy {
			toPath := strings.ReplaceAll(proxy.To, "{port}", fmt.Sprintf("%d", port))

			caddyfileBuilder.WriteString(fmt.Sprintf("  reverse_proxy %s %s\n", proxy.Path, toPath))
		}
		caddyfileBuilder.WriteString("}\n\n")
	}

	return caddyfileBuilder.String(), nil
}

func SetupCaddy(port int, cfg config.DeploymentConfig) error {
	caddyfile, err := ConvertToCaddyfile(cfg.Caddy, port)
	if err != nil {
		return err
	}
	fmt.Println("New Caddy config")
	fmt.Println(caddyfile)

	return nil
}
