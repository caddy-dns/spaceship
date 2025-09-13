package spaceship

import (
    "os"
    "strings"
    "testing"
    "time"

    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
    libdnsspaceship "github.com/libdns/spaceship"
)

// tokenize is a very small tokenizer sufficient for the test snippets.
// It splits on whitespace and treats '{' and '}' as standalone tokens.
func tokenize(file, input string) []caddyfile.Token {
	var tokens []caddyfile.Token
	line := 1
	flush := func(buf *strings.Builder) {
		if buf.Len() == 0 {
			return
		}
		tokens = append(tokens, caddyfile.Token{File: file, Line: line, Text: buf.String()})
		buf.Reset()
	}
	var buf strings.Builder
	for _, r := range input {
		switch r {
		case ' ', '\t', '\r':
			flush(&buf)
		case '\n':
			flush(&buf)
			line++
		case '{', '}':
			flush(&buf)
			tokens = append(tokens, caddyfile.Token{File: file, Line: line, Text: string(r)})
		default:
			buf.WriteRune(r)
		}
	}
	flush(&buf)
	return tokens
}

// helper to run UnmarshalCaddyfile on a snippet
func parseProvider(t *testing.T, input string) *Provider {
    t.Helper()
    p := &Provider{new(libdnsspaceship.Provider)}
    toks := tokenize("test.caddy", input)
    d := caddyfile.NewDispenser(toks)
    if err := p.UnmarshalCaddyfile(d); err != nil {
        t.Fatalf("UnmarshalCaddyfile failed: %v", err)
    }
    return p
}

func TestUnmarshalCaddyfile_Inline(t *testing.T) {
    p := parseProvider(t, `spaceship key123 secret456`)
    if p.APIKey != "key123" || p.APISecret != "secret456" {
        t.Fatalf("unexpected credentials: %s / %s", p.APIKey, p.APISecret)
    }
}

func TestUnmarshalCaddyfile_Block(t *testing.T) {
	input := `spaceship {
		api_key key123
		api_secret secret456
		api_pagesize 250
		api_timeout 15
		api_url https://example.test/api
	}`
	p := parseProvider(t, input)
	if p.APIKey != "key123" || p.APISecret != "secret456" {
		t.Fatalf("credentials mismatch")
	}
	if p.PageSize != 250 {
		t.Fatalf("expected pagesize 250 got %d", p.PageSize)
	}
	if p.HTTPClient == nil || p.HTTPClient.Timeout != 15*time.Second {
		t.Fatalf("expected timeout 15s got %#v", p.HTTPClient)
	}
	if p.BaseURL != "https://example.test/api" {
		t.Fatalf("unexpected base url: %s", p.BaseURL)
	}
}

func TestUnmarshalCaddyfile_DuplicateKey(t *testing.T) {
	input := `spaceship {
		api_key a
		api_key b
		api_secret s
	}`
	p := &Provider{new(libdnsspaceship.Provider)}
	toks := tokenize("test.caddy", input)
	d := caddyfile.NewDispenser(toks)
	if err := p.UnmarshalCaddyfile(d); err == nil {
		t.Fatalf("expected error for duplicate api_key")
	}
}

func TestProvision_EnvFallbackAndPlaceholders(t *testing.T) {
	os.Setenv("LIBDNS_SPACESHIP_APIKEY", "envkey")
	os.Setenv("LIBDNS_SPACESHIP_APISECRET", "envsecret")
	defer func() { os.Unsetenv("LIBDNS_SPACESHIP_APIKEY"); os.Unsetenv("LIBDNS_SPACESHIP_APISECRET") }()
	p := &Provider{new(libdnsspaceship.Provider)}
	// Use placeholders referencing env
	p.APIKey = "{env.LIBDNS_SPACESHIP_APIKEY}"
	p.APISecret = "{env.LIBDNS_SPACESHIP_APISECRET}"
	var ctx caddy.Context // zero value is fine for our Provision since we don't use it
	if err := p.Provision(ctx); err != nil {
		t.Fatalf("Provision failed: %v", err)
	}
	if p.APIKey != "envkey" || p.APISecret != "envsecret" {
		t.Fatalf("placeholder expansion failed: %s / %s", p.APIKey, p.APISecret)
	}
}
