package caddy

import "net/http"

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
