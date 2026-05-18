// Copyright 2022 Nicky Gerritsen
package powerdns

import (
	"errors"

	powerdns "github.com/vapronva/libdns-powerdns"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

type Provider struct{ *powerdns.Provider }

func init() {
	caddy.RegisterModule(Provider{})
}

func (Provider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dns.providers.powerdns",
		New: func() caddy.Module { return &Provider{new(powerdns.Provider)} },
	}
}

func (p *Provider) Provision(_ caddy.Context) error {
	repl := caddy.NewReplacer()
	p.ServerURL = repl.ReplaceAll(p.ServerURL, "")
	p.APIToken = repl.ReplaceAll(p.APIToken, "")
	p.ServerID = repl.ReplaceAll(p.ServerID, "")
	p.Debug = repl.ReplaceAll(p.Debug, "")
	if p.ServerURL == "" {
		return errors.New("server URL is required")
	}
	if p.APIToken == "" {
		return errors.New("API token is required")
	}
	return nil
}

func (p *Provider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if !d.Next() {
		return nil
	}
	if d.NextArg() {
		p.ServerURL = d.Val()
	}
	if d.NextArg() {
		p.APIToken = d.Val()
	}
	if d.NextArg() {
		p.ServerID = d.Val()
	}
	if d.NextArg() {
		return d.ArgErr()
	}
	fields := map[string]*string{
		"server_url": &p.ServerURL,
		"api_token":  &p.APIToken,
		"server_id":  &p.ServerID,
		"debug":      &p.Debug,
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		key := d.Val()
		dst := fields[key]
		if dst == nil {
			return d.Errf("unrecognized subdirective '%s'", key)
		}
		if *dst != "" {
			return d.Errf("%s already set", key)
		}
		if !d.NextArg() {
			return d.ArgErr()
		}
		*dst = d.Val()
		if d.NextArg() {
			return d.ArgErr()
		}
	}
	return nil
}

var (
	_ caddyfile.Unmarshaler = (*Provider)(nil)
	_ caddy.Provisioner     = (*Provider)(nil)
)
