package template

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	libdnsspaceship "github.com/libdns/spaceship"
)

// Provider lets Caddy read and manipulate DNS records hosted by the Spaceship DNS provider.
// It adapts the libdns spaceship provider for use in Caddy.
type Provider struct{ *libdnsspaceship.Provider }

func init() {
	caddy.RegisterModule(Provider{})
}

// CaddyModule returns the Caddy module information.
func (Provider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dns.providers.spaceship",
		New: func() caddy.Module { return &Provider{new(libdnsspaceship.Provider)} },
	}
}

// TODO: This is just an example. Useful to allow env variable placeholders; update accordingly.
// Provision sets up the module. Implements caddy.Provisioner.
func (p *Provider) Provision(ctx caddy.Context) error {
	// Single replacer instance for all placeholder expansions
	repl := caddy.NewReplacer()
	// Expand placeholders for fields that may already be set from Caddyfile
	p.Provider.APIKey = repl.ReplaceAll(p.Provider.APIKey, "")
	p.Provider.APISecret = repl.ReplaceAll(p.Provider.APISecret, "")
	p.Provider.BaseURL = repl.ReplaceAll(p.Provider.BaseURL, "")

	// Populate any still-empty fields from environment variables (libdns helper)
	p.Provider.PopulateFromEnv()

	// Re-expand in case env values contain placeholders (unlikely but cheap)
	p.Provider.APIKey = repl.ReplaceAll(p.Provider.APIKey, "")
	p.Provider.APISecret = repl.ReplaceAll(p.Provider.APISecret, "")
	p.Provider.BaseURL = repl.ReplaceAll(p.Provider.BaseURL, "")

	// Ensure HTTP client has a timeout
	if p.Provider.HTTPClient != nil && p.Provider.HTTPClient.Timeout == 0 {
		p.Provider.HTTPClient.Timeout = 30 * time.Second
	}

	if p.Provider.APIKey == "" || p.Provider.APISecret == "" {
		return fmt.Errorf("spaceship: api_key and api_secret are required")
	}
	return nil
}

// TODO: This is just an example. Update accordingly.
// UnmarshalCaddyfile sets up the DNS provider from Caddyfile tokens. Syntax:
//
// providername [<api_token>] {
//     api_token <api_token>
// }
//
// **THIS IS JUST AN EXAMPLE AND NEEDS TO BE CUSTOMIZED.**
func (p *Provider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		// Optional inline args: api_key [api_secret]
		if d.NextArg() {
			p.Provider.APIKey = d.Val()
		}
		if d.NextArg() {
			p.Provider.APISecret = d.Val()
		}
		if d.NextArg() { // too many inline args
			return d.ArgErr()
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "api_key":
				if p.Provider.APIKey != "" {
					return d.Err("api_key already set")
				}
				if !d.NextArg() {
					return d.Err("api_key requires value")
				}
				p.Provider.APIKey = d.Val()
				if d.NextArg() { // no extra args
					return d.ArgErr()
				}
			case "api_secret":
				if p.Provider.APISecret != "" {
					return d.Err("api_secret already set")
				}
				if !d.NextArg() {
					return d.Err("api_secret requires value")
				}
				p.Provider.APISecret = d.Val()
				if d.NextArg() {
					return d.ArgErr()
				}
			case "api_url":
				if p.Provider.BaseURL != "" {
					return d.Err("api_url already set")
				}
				if !d.NextArg() {
					return d.Err("api_url requires value")
				}
				p.Provider.BaseURL = d.Val()
				if d.NextArg() { return d.ArgErr() }
			case "api_pagesize":
				if !d.NextArg() { return d.Err("api_pagesize requires value") }
				ps, err := strconv.Atoi(d.Val())
				if err != nil || ps <= 0 { return d.Err("api_pagesize must be a positive integer") }
				p.Provider.PageSize = ps
				if d.NextArg() { return d.ArgErr() }
			case "api_timeout":
				if !d.NextArg() { return d.Err("api_timeout requires value (seconds)") }
				secs, err := strconv.Atoi(d.Val())
				if err != nil || secs <= 0 { return d.Err("api_timeout must be a positive integer (seconds)") }
				if p.Provider.HTTPClient == nil {
					p.Provider.HTTPClient = &http.Client{Timeout: time.Duration(secs) * time.Second}
				} else {
					p.Provider.HTTPClient.Timeout = time.Duration(secs) * time.Second
				}
				if d.NextArg() { return d.ArgErr() }
			default:
				return d.Errf("unrecognized subdirective '%s'", d.Val())
			}
		}
	}
	if p.Provider.APIKey == "" || p.Provider.APISecret == "" {
		return d.Err("missing api_key or api_secret")
	}
	return nil
}

// Interface guards
var (
	_ caddyfile.Unmarshaler = (*Provider)(nil)
	_ caddy.Provisioner     = (*Provider)(nil)
)
