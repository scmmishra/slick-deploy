package health

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/scmmishra/slick-deploy/internal/config"
)

func CheckHealth(host string, cfg *config.HealthCheck) error {
	return CheckHealthWithClock(host, cfg, clockwork.NewRealClock())
}

func CheckHealthWithClock(host string, cfg *config.HealthCheck, clock clockwork.Clock) error {
	if cfg.Endpoint == "" || host == "" {
		return nil
	}

	// if cfg.Endpoint starts with /, remove it
	if cfg.Endpoint[0] == '/' {
		cfg.Endpoint = cfg.Endpoint[1:]
	}

	endpoint := fmt.Sprintf("%s/%s", host, cfg.Endpoint)
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second

	client := &http.Client{
		Timeout: timeout,
	}

	maxRetries := cfg.MaxRetries
	delay := time.Duration(cfg.IntervalSeconds) * time.Second

	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(endpoint)
		if err != nil {
			clock.Sleep(delay)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			return nil
		}

		fmt.Println("  Retrying...")
		clock.Sleep(delay)
	}

	return fmt.Errorf("unable to reach endpoint %s after %d attempts", endpoint, maxRetries)
}
