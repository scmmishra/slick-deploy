package caddy

import (
	"fmt"
	"strings"

	"github.com/scmmishra/slick-deploy/internal/config"
)

// Rule defines a routing rule with a match pattern and reverse proxy settings.
type Rule struct {
	Match        string         `yaml:"match"`
	ReverseProxy []ReverseProxy `yaml:"reverse_proxy"`
}

// ReverseProxy specifies the reverse proxy configuration.
type ReverseProxy struct {
	Path string `yaml:"path"`
	To   string `yaml:"to"`
}

// buildGlobalOptions formats global options for the Caddyfile.
func buildGlobalOptions(globalCfg config.GlobalOptions, port int) string {
	var builder strings.Builder

	hasEmail := globalCfg.Email != ""
	hasOnDemandTls := globalCfg.OnDemandTls.Ask != ""

	if hasEmail || hasOnDemandTls {
		builder.WriteString("{\n")
		if hasEmail {
			builder.WriteString(fmt.Sprintf("  email %s\n", globalCfg.Email))
		}
		if hasOnDemandTls {
			appendOnDemandTlsConfig(&builder, globalCfg.OnDemandTls, port)
		}
		builder.WriteString("}\n\n")
	}

	return builder.String()
}

// appendOnDemandTlsConfig adds the on_demand_tls configuration.
func appendOnDemandTlsConfig(builder *strings.Builder, tlsConfig config.OnDemandTlsConfig, port int) {
	builder.WriteString("  on_demand_tls {\n")
	builder.WriteString(fmt.Sprintf("    ask %s\n", strings.ReplaceAll(tlsConfig.Ask, "{port}", fmt.Sprintf("%d", port))))

	if tlsConfig.Interval != "" {
		builder.WriteString(fmt.Sprintf("    interval %s\n", tlsConfig.Interval))
	}
	if tlsConfig.Burst != "" {
		builder.WriteString(fmt.Sprintf("    burst %s\n", tlsConfig.Burst))
	}
	builder.WriteString("  }\n")
}

// ConvertToCaddyfile converts configuration into a Caddyfile representation.
func ConvertToCaddyfile(caddyCfg config.CaddyConfig, port int) string {
	var builder strings.Builder

	builder.WriteString(buildGlobalOptions(caddyCfg.Global, port))
	for _, rule := range caddyCfg.Rules {
		appendRule(&builder, rule, port)
	}

	return builder.String()
}

// appendRule adds a server block with its configuration.
func appendRule(builder *strings.Builder, rule config.Rule, port int) {
	builder.WriteString(rule.Match + " {\n")
	if rule.Tls != "" {
		builder.WriteString(fmt.Sprintf("  tls %s\n", strings.ReplaceAll(rule.Tls, "{port}", fmt.Sprintf("%d", port))))
	}

	for _, handle := range rule.Handle {
		builder.WriteString(fmt.Sprintf("  handle %s {\n", handle.Path))
		for _, directive := range handle.Directives {
			builder.WriteString(fmt.Sprintf("    %s\n", directive))
		}
		builder.WriteString("  }\n")
	}

	for _, proxy := range rule.ReverseProxy {
		toPath := strings.ReplaceAll(proxy.To, "{port}", fmt.Sprintf("%d", port))
		builder.WriteString(fmt.Sprintf("  reverse_proxy %s %s\n", proxy.Path, toPath))
	}
	builder.WriteString("}\n\n")
}

// SetupCaddy loads the Caddyfile configuration into Caddy.
func SetupCaddy(port int, cfg config.DeploymentConfig) error {
	caddyfile := ConvertToCaddyfile(cfg.Caddy, port)
	client := NewCaddyClient(cfg.Caddy.AdminAPI)
	return client.Load(caddyfile)
}
