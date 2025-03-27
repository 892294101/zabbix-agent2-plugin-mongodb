/*
** Copyright (C) 2001-2025 Zabbix SIA
**
** This program is free software: you can redistribute it and/or modify it under the terms of
** the GNU Affero General Public License as published by the Free Software Foundation, version 3.
**
** This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
** without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
** See the GNU Affero General Public License for more details.
**
** You should have received a copy of the GNU Affero General Public License along with this program.
** If not, see <https://www.gnu.org/licenses/>.
**/

package plugin

import (
	"fmt"

	"golang.zabbix.com/sdk/conf"
	"golang.zabbix.com/sdk/plugin"
)

const (
	empty   = ""
	req     = "required"
	reqCa   = "verify_ca"
	reqFull = "verify_full"
)

var validTLSOptions = []string{empty, req, reqCa, reqFull}

type Session struct {
	URI         string `conf:"name=Uri,optional"`
	Password    string `conf:"optional"`
	User        string `conf:"optional"`
	TLSConnect  string `conf:"name=TLSConnect,optional"`
	TLSCAFile   string `conf:"name=TLSCAFile,optional"`
	TLSCertFile string `conf:"name=TLSCertFile,optional"`
	TLSKeyFile  string `conf:"name=TLSKeyFile,optional"`
}

type PluginOptions struct {
	System plugin.SystemOptions `conf:"optional"` //nolint:staticcheck
	// Timeout is the amount of time to wait for a server to respond when
	// first connecting and on follow up operations in the session.
	Timeout int `conf:"optional,range=1:30"`

	// KeepAlive is a time to wait before unused connections will be closed.
	KeepAlive int `conf:"optional,range=60:900,default=60"`

	// Sessions stores pre-defined named sets of connections settings.
	Sessions map[string]Session `conf:"optional"`

	// Default stores default connection parameter values from configuration file
	Default Session `conf:"optional"`
}

// Configure implements the Configurator interface.
// Initializes configuration structures.
func (p *Plugin) Configure(global *plugin.GlobalOptions, options any) {
	if err := conf.UnmarshalStrict(options, &p.options); err != nil {
		p.Errf("cannot unmarshal configuration options: %s", err)
	}

	if p.options.Timeout == 0 {
		p.options.Timeout = global.Timeout
	}
}

// Validate implements the Configurator interface.
// Returns an error if validation of a plugin's configuration is failed.
func (*Plugin) Validate(options any) error {
	var opts PluginOptions

	err := conf.UnmarshalStrict(options, &opts)
	if err != nil {
		return err
	}

	for _, s := range opts.Sessions {
		if !contains(validTLSOptions, s.TLSConnect) {
			return fmt.Errorf("incorrect tls connection type %s", s.TLSConnect)
		}
	}

	return nil
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}

	return false
}
