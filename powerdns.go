// Copyright 2022 Nicky Gerritsen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package powerdns

import (
	powerdns "github.com/vapronva/libdns-powerdns"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

// Provider wraps the provider implementation as a Caddy module.
type Provider struct{ *powerdns.Provider }

func init() {
	caddy.RegisterModule(Provider{})
}

// CaddyModule returns the Caddy module information.
func (Provider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dns.providers.powerdns",
		New: func() caddy.Module { return &Provider{new(powerdns.Provider)} },
	}
}

// Provision implements the Provisioner interface to initialize the PowerDNS client
func (p *Provider) Provision(ctx caddy.Context) error {
	repl := caddy.NewReplacer()
	p.ServerURL = repl.ReplaceAll(p.ServerURL, "")
	p.APIToken = repl.ReplaceAll(p.APIToken, "")
	p.ServerID = repl.ReplaceAll(p.ServerID, "")

	return nil
}

// UnmarshalCaddyfile sets up the DNS provider from Caddyfile tokens. Syntax:
//
//	powerdns {
//	    server_url <string>
//	    api_token <string>
//	    server_id <string>
//	}
func (p *Provider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
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
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "server_url":
				if p.ServerURL != "" {
					return d.Err("Server URL already set")
				}
				if d.NextArg() {
					p.ServerURL = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "api_token":
				if p.APIToken != "" {
					return d.Err("API token already set")
				}
				if d.NextArg() {
					p.APIToken = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "server_id":
				if p.ServerID != "" {
					return d.Err("Server ID already set")
				}
				if d.NextArg() {
					p.ServerID = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			default:
				return d.Errf("unrecognized subdirective '%s'", d.Val())
			}
		}
	}
	if p.ServerURL == "" {
		return d.Err("missing server URL")
	}
	if p.APIToken == "" {
		return d.Err("missing API token")
	}
	return nil
}

// Interface guards
var (
	_ caddyfile.Unmarshaler = (*Provider)(nil)
	_ caddy.Provisioner     = (*Provider)(nil)
)
