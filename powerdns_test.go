package powerdns_test

import (
	"strings"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	libdnspowerdns "github.com/vapronva/libdns-powerdns"

	caddydns "github.com/vapronva/caddydns-powerdns"
)

func unmarshal(t *testing.T, input string) (*caddydns.Provider, error) {
	t.Helper()
	p := &caddydns.Provider{Provider: new(libdnspowerdns.Provider)}
	return p, p.UnmarshalCaddyfile(caddyfile.NewTestDispenser(input))
}

func TestUnmarshalCaddyfileSegmentScopedDispenserDoesNotConsumeSiblingDirective(t *testing.T) {
	d := caddyfile.NewTestDispenser(`dns powerdns http://localhost secret
other_directive other_arg`)
	if !d.Next() {
		t.Fatal("expected to consume dns directive")
	}
	if !d.Next() {
		t.Fatal("expected to consume provider name")
	}
	if d.Val() != "powerdns" {
		t.Fatalf("expected provider name powerdns, got %q", d.Val())
	}
	segment := d.NewFromNextSegment()
	p := &caddydns.Provider{Provider: new(libdnspowerdns.Provider)}
	if err := p.UnmarshalCaddyfile(segment); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ServerURL != "http://localhost" || p.APIToken != "secret" {
		t.Fatalf("unexpected provider values: server_url=%q api_token=%q", p.ServerURL, p.APIToken)
	}
	if !d.Next() {
		t.Fatal("expected other_directive to remain for outer parser")
	}
	if d.Val() != "other_directive" {
		t.Fatalf("expected remaining token to be other_directive, got %q", d.Val())
	}
}

func TestUnmarshalCaddyfileBlockForm(t *testing.T) {
	p, err := unmarshal(t, `powerdns {
		server_url http://pdns:8081
		api_token  secret
		server_id  primary
		debug      stderr
	}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ServerURL != "http://pdns:8081" || p.APIToken != "secret" ||
		p.ServerID != "primary" || p.Debug != "stderr" {
		t.Fatalf("unexpected values: %+v", p.Provider)
	}
}

func TestUnmarshalCaddyfileHybridPositionalAndBlock(t *testing.T) {
	p, err := unmarshal(t, `powerdns http://pdns:8081 secret {
		server_id primary
		debug     stderr
	}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ServerURL != "http://pdns:8081" || p.APIToken != "secret" ||
		p.ServerID != "primary" || p.Debug != "stderr" {
		t.Fatalf("unexpected values: %+v", p.Provider)
	}
}

func TestUnmarshalCaddyfileErrors(t *testing.T) {
	cases := map[string]string{
		"unrecognized subdirective": `powerdns {
			bogus value
		}`,
		"server_url already set": `powerdns http://a secret {
			server_url http://b
		}`,
		"missing arg": `powerdns {
			api_token
		}`,
		"too many args": `powerdns {
			server_id a b
		}`,
		"too many positional args": `powerdns http://a secret id extra`,
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := unmarshal(t, input); err == nil {
				t.Fatalf("expected error for %q, got nil", name)
			}
		})
	}
}

func TestUnmarshalCaddyfileMissingRequiredFieldReportsFileLine(t *testing.T) {
	cases := map[string]string{
		"missing api_token": `powerdns {
			server_url http://pdns:8081
		}`,
		"missing server_url": `powerdns {
			api_token secret
		}`,
		"missing both": `powerdns`,
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := unmarshal(t, input)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", name)
			}
			if !strings.Contains(err.Error(), "Testfile:") {
				t.Fatalf("expected Caddyfile file/line context in error, got: %v", err)
			}
		})
	}
}

func TestProvisionExpandsPlaceholdersAndValidates(t *testing.T) {
	t.Setenv("POWERDNS_TEST_URL", "http://pdns:8081")
	t.Setenv("POWERDNS_TEST_TOKEN", "secret")
	p := &caddydns.Provider{Provider: new(libdnspowerdns.Provider)}
	p.ServerURL = "{env.POWERDNS_TEST_URL}"
	p.APIToken = "{env.POWERDNS_TEST_TOKEN}"
	if err := p.Provision(caddy.Context{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ServerURL != "http://pdns:8081" || p.APIToken != "secret" {
		t.Fatalf("placeholders not expanded: %+v", p.Provider)
	}
	if p.ServerID != "" {
		t.Fatalf("server_id should stay empty, got %q", p.ServerID)
	}
}

func TestProvisionRequiresServerURLAndToken(t *testing.T) {
	if err := (&caddydns.Provider{Provider: &libdnspowerdns.Provider{APIToken: "x"}}).
		Provision(caddy.Context{}); err == nil {
		t.Fatal("expected error when server URL is missing")
	}
	if err := (&caddydns.Provider{Provider: &libdnspowerdns.Provider{ServerURL: "http://a"}}).
		Provision(caddy.Context{}); err == nil {
		t.Fatal("expected error when API token is missing")
	}
}
