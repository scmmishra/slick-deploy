package caddy

import (
	"errors"
	"testing"

	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConvertToCaddyfile(t *testing.T) {
	caddyCfg := config.CaddyConfig{
		Global: config.GlobalOptions{
			Email: "test@example.com",
			OnDemandTls: config.OnDemandTlsConfig{
				Ask:      "https://acme.example.com/directory",
				Interval: "3600",
				Burst:    "13",
			},
		},
		Rules: []config.Rule{
			{
				Match: "localhost",
				Tls:   "internal",
				Handle: []config.Handle{
					{
						Path: "/",
						Directives: []string{
							"root * /usr/share/caddy",
						},
					},
					{
						Path: "/healthz",
						Directives: []string{
							"respond \"OK\" 200",
						},
					},
				},
				ReverseProxy: []config.ReverseProxy{
					{
						Path: "/",
						To:   "http://localhost:{port}",
						HeaderUp: []config.HeaderUp{
							{
								Name:  "X-Real-IP",
								Value: "{http.request.remote.host}",
							},
						},
					},
				},
			},
		},
	}

	caddyfile := ConvertToCaddyfile(caddyCfg, 8080)

	expectedCaddyfile := `{
  email test@example.com
  on_demand_tls {
    ask https://acme.example.com/directory
    interval 3600
    burst 13
  }
}

localhost {
  tls {
    internal
  }
  handle / {
    root * /usr/share/caddy
  }
  handle /healthz {
    respond "OK" 200
  }
  reverse_proxy / http://localhost:8080 {
    header_up X-Real-IP {http.request.remote.host}
  }
}

`
	assert.Equal(t, expectedCaddyfile, caddyfile)
}

func TestConvertToCaddyfile_EmptyRules(t *testing.T) {
	caddyCfg := config.CaddyConfig{
		Rules: []config.Rule{},
	}
	caddyfile := ConvertToCaddyfile(caddyCfg, 8080)
	assert.Equal(t, "", caddyfile)
}

type MockCaddyClient struct {
	mock.Mock
}

func (m *MockCaddyClient) Load(caddyfile string) error {
	args := m.Called(caddyfile)
	return args.Error(0)
}

func TestSetupCaddy(t *testing.T) {
	mockClient := new(MockCaddyClient)
	mockClient.On("Load", mock.Anything).Return(nil)

	// replace NewCaddyClient with a function that returns the mock client
	oldNewCaddyClient := NewCaddyClient
	NewCaddyClient = func(baseURL string) CaddyClientInterface { return mockClient }
	defer func() { NewCaddyClient = oldNewCaddyClient }() // restore original function after the test

	cfg := config.DeploymentConfig{
		Caddy: config.CaddyConfig{
			AdminAPI: "http://localhost:2019",
			Rules:    []config.Rule{},
		},
	}
	err := SetupCaddy(8080, cfg)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestSetupCaddy_Error(t *testing.T) {
	mockClient := new(MockCaddyClient)
	mockClient.On("Load", mock.Anything).Return(errors.New("mock error"))

	// replace NewCaddyClient with a function that returns the mock client
	oldNewCaddyClient := NewCaddyClient
	NewCaddyClient = func(baseURL string) CaddyClientInterface { return mockClient }
	defer func() { NewCaddyClient = oldNewCaddyClient }() // restore original function after the test

	cfg := config.DeploymentConfig{
		Caddy: config.CaddyConfig{
			AdminAPI: "http://localhost:2019",
			Rules:    []config.Rule{},
		},
	}
	err := SetupCaddy(8080, cfg)
	assert.Error(t, err)

	mockClient.AssertExpectations(t)
}
