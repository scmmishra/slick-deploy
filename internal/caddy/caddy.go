package caddy

import (
	"fmt"
	"strings"

	"github.com/scmmishra/slick-deploy/internal/config"
)

func buildGlobalOptions(globalCfg config.GlobalOptions, port int) string {
	var globalOptionsBuilder strings.Builder

	if globalCfg.Email != "" || globalCfg.OnDemandTls.Ask != "" {
		globalOptionsBuilder.WriteString("{\n")
		if globalCfg.Email != "" {
			globalOptionsBuilder.WriteString(fmt.Sprintf("  email %s\n", globalCfg.Email))
		}

		if globalCfg.OnDemandTls.Ask != "" {
			var onDemandTlsBuilder strings.Builder

			onDemandTlsBuilder.WriteString("  on_demand_tls {\n")
			askPath := strings.ReplaceAll(globalCfg.OnDemandTls.Ask, "{port}", fmt.Sprintf("%d", port))
			onDemandTlsBuilder.WriteString(fmt.Sprintf("    ask %s\n", askPath))

			if globalCfg.OnDemandTls.Interval != "" {
				onDemandTlsBuilder.WriteString(fmt.Sprintf("    interval %s\n", globalCfg.OnDemandTls.Interval))
			}

			if globalCfg.OnDemandTls.Burst != "0" {
				onDemandTlsBuilder.WriteString(fmt.Sprintf("    burst %s\n", globalCfg.OnDemandTls.Burst))
			}

			onDemandTlsBuilder.WriteString("  }\n")
			globalOptionsBuilder.WriteString(onDemandTlsBuilder.String())
		}

		globalOptionsBuilder.WriteString("}\n\n")
	}

	return globalOptionsBuilder.String()
}

// ConvertToCaddyfile translates the CaddyConfig struct to a Caddyfile string
func ConvertToCaddyfile(caddyCfg config.CaddyConfig, port int) string {
	var caddyfileBuilder strings.Builder

	caddyfileBuilder.WriteString(buildGlobalOptions(caddyCfg.Global, port))

	for _, rule := range caddyCfg.Rules {
		caddyfileBuilder.WriteString(rule.Match)
		caddyfileBuilder.WriteString(" {\n")

		if rule.Tls != "" {
			tlsRule := strings.ReplaceAll(rule.Tls, "{port}", fmt.Sprintf("%d", port))
			caddyfileBuilder.WriteString(fmt.Sprintf("  tls {\n    %s\n  }\n", tlsRule))
		}

		// check if rule.Handle array is not empty
		if len(rule.Handle) > 0 {
			for _, handle := range rule.Handle {
				caddyfileBuilder.WriteString(fmt.Sprintf("  handle %s {\n", handle.Path))
				for _, directive := range handle.Directives {
					caddyfileBuilder.WriteString(fmt.Sprintf("    %s\n", directive))
				}
				caddyfileBuilder.WriteString("  }\n")
			}
		}

		for _, proxy := range rule.ReverseProxy {
			toPath := strings.ReplaceAll(proxy.To, "{port}", fmt.Sprintf("%d", port))
			caddyfileBuilder.WriteString(fmt.Sprintf("  reverse_proxy %s %s\n", proxy.Path, toPath))
		}
		caddyfileBuilder.WriteString("}\n\n")
	}

	return caddyfileBuilder.String()
}

func SetupCaddy(port int, cfg config.DeploymentConfig) error {
	caddyfile := ConvertToCaddyfile(cfg.Caddy, port)
	client := NewCaddyClient(cfg.Caddy.AdminAPI)
	err := client.Load(caddyfile)

	if err != nil {
		return err
	}

	return nil
}
