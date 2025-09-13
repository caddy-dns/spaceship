Spaceship DNS provider module for Caddy
=======================================

[![CI](https://github.com/caddy-dns/spaceship/actions/workflows/ci.yml/badge.svg)](https://github.com/caddy-dns/spaceship/actions/workflows/ci.yml)

This package contains a DNS provider module for [Caddy](https://github.com/caddyserver/caddy). It manages DNS records using the Spaceship DNS API through the libdns adapter.

## Caddy module name

```
dns.providers.spaceship
```

## Config examples

To use this module for the ACME DNS challenge, [configure the ACME issuer in your Caddy JSON](https://caddyserver.com/docs/json/apps/tls/automation/policies/issuer/acme/) like so:

```json
{
	"module": "acme",
	"challenges": {
		"dns": {
			"provider": {
				"name": "spaceship",
				"api_token": "YOUR_SPACESHIP_API_TOKEN",
				"api_secret": "YOUR_SPACESHIP_API_SECRET",
			}
		}
	}
}
```

### Caddyfile usage

You can configure the provider globally or per-site. Supported directives inside the provider block:

* `api_key <value>` – (required) Spaceship API key
* `api_secret <value>` (alias: `apisecret`) – (required) Spaceship API secret
* `api_url <value>` (aliases: `api_baseurl`, `base_url`) – override base API URL (default `https://spaceship.dev/api`)
* `api_pagesize <n>` (aliases: `pagesize`, `page_size`) – page size for listing operations (default 100)
* `api_timeout <seconds>` (alias: `timeout`) – HTTP client timeout in seconds (default 30)

Inline form (key + secret):
```
tls {
	dns spaceship YOUR_API_KEY YOUR_API_SECRET
}
```

Block form:
```
tls {
	dns spaceship {
		api_key    {env.LIBDNS_SPACESHIP_APIKEY}
		api_secret {env.LIBDNS_SPACESHIP_APISECRET}
		api_timeout 20
		api_pagesize 200
	}
}
```

Global example:
```
{
	acme_dns spaceship {
		api_key    {env.LIBDNS_SPACESHIP_APIKEY}
		api_secret {env.LIBDNS_SPACESHIP_APISECRET}
	}
}
```

Environment variable fallbacks recognized by the underlying provider: `LIBDNS_SPACESHIP_APIKEY`, `LIBDNS_SPACESHIP_APISECRET`, `LIBDNS_SPACESHIP_BASEURL`, `LIBDNS_SPACESHIP_PAGESIZE`, `LIBDNS_SPACESHIP_TIMEOUT`.

## Compatibility

This module requires Caddy v2.10.0 or newer (tested with v2.10.2). Older Caddy versions (≤ v2.4.x) use an earlier `libdns` API and will not compile with this provider.

Use Caddy's placeholder `{env.VAR}` to reference environment variables in the Caddyfile.
