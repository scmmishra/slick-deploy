package health

import (
	"testing"

	"github.com/scmmishra/slick-deploy/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestCheckHealth(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		endpoint string
		wantErr  bool
	}{
		{
			name:     "Healthy host",
			host:     "http://shivam.dev",
			endpoint: "/health",
			wantErr:  false,
		},
		{
			name:     "Unhealthy host",
			host:     "http://unhealthy.shivam.dev",
			endpoint: "/health",
			wantErr:  true,
		},
		{
			name:     "Empty host",
			host:     "",
			endpoint: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckHealth(tt.host, &config.HealthCheck{
				Endpoint:       tt.endpoint,
				TimeoutSeconds: 5,
			})

			if tt.wantErr {
				assert.Error(t, err, "CheckHealth() should return an error")
			} else {
				assert.NoError(t, err, "CheckHealth() should not return an error")
			}
		})
	}
}
