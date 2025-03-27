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
	"context"
	"errors"
	"time"

	"golang.zabbix.com/plugin/mongodb/plugin/handlers"
	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/metric"
	"golang.zabbix.com/sdk/plugin"
	"golang.zabbix.com/sdk/uri"
	"golang.zabbix.com/sdk/zbxerr"
)

const (
	Name       = "MongoDB"
	hkInterval = 10
)

var Impl Plugin

// Plugin -
type Plugin struct {
	plugin.Base
	connMgr *ConnManager
	options PluginOptions
}

// Export metrics.
//
//nolint:gocyclo,cyclop
func (p *Plugin) Export(key string, rawParams []string, pluginCtx plugin.ContextProvider) (any, error) {
	params, _, hc, err := metrics[key].EvalParams(rawParams, p.options.Sessions)
	if err != nil {
		return nil, err
	}

	err = metric.SetDefaults(params, hc, p.options.Default)
	if err != nil {
		return nil, err
	}

	uri, err := uri.NewWithCreds(params["URI"], params["User"], params["Password"], handlers.UriDefaults)
	if err != nil {
		return nil, err
	}

	handleMetric := getHandlerFunc(key)
	if handleMetric == nil {
		return nil, zbxerr.ErrorUnsupportedMetric
	}

	conn, err := p.connMgr.GetConnection(*uri, params)
	if err != nil {
		// Special logic of processing connection errors should be used if mongodb.ping is requested
		// because it must return pingFailed if any error occurred.
		if key == keyPing {
			p.Debugf(err.Error())

			return handlers.PingFailed, nil
		}

		p.Errf(err.Error())

		return nil, err
	}

	p.Debugf("Params: %v", params)

	timeout := conn.getTimeout()

	if timeout < time.Second*time.Duration(pluginCtx.Timeout()) {
		timeout = time.Second * time.Duration(pluginCtx.Timeout())
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, err := handleMetric(ctx, conn, params)
	if err != nil {
		p.Errf(err.Error())

		ctxErr := ctx.Err()

		if ctxErr != nil && errors.Is(ctxErr, context.DeadlineExceeded) {
			return nil, errs.New("request execution timeout exceeded")
		}

		return nil, errs.Wrap(err, "failed to run command")
	}

	return result, err
}

// Start implements the Runner interface and performs initialization when plugin is activated.
func (p *Plugin) Start() {
	handlers.Logger = p.Logger
	p.connMgr = NewConnManager(
		time.Duration(p.options.KeepAlive)*time.Second,
		time.Duration(p.options.Timeout)*time.Second,
		hkInterval*time.Second,
		p.Logger,
	)
}

// Stop implements the Runner interface and frees resources when plugin is deactivated.
func (p *Plugin) Stop() {
	p.connMgr.Destroy()
	p.connMgr = nil
}
