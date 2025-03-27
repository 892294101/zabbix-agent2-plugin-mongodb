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

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"golang.zabbix.com/sdk/zbxerr"
)

type lldCfgEntity struct {
	ReplicaSet string `json:"{#REPLICASET}"`
	Hostname   string `json:"{#HOSTNAME}"`
	MongodURI  string `json:"{#MONGOD_URI}"`
}

type shardMap struct {
	Map map[string]string
}

// ConfigDiscoveryHandler
// https://docs.mongodb.com/manual/reference/command/getShardMap/#dbcmd.getShardMap
func ConfigDiscoveryHandler(ctx context.Context, s Session, _ map[string]string) (any, error) {
	var cfgServers shardMap
	err := s.DB("admin").Run(
		ctx,
		&bson.D{
			{Key: "getShardMap", Value: 1},
		},
		&cfgServers,
	)

	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	lld := make([]lldCfgEntity, 0)

	if servers, ok := cfgServers.Map["config"]; ok {
		lld, err = handlerServer(servers, lld)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, zbxerr.ErrorCannotParseResult
	}

	jsonRes, err := json.Marshal(lld)
	if err != nil {
		return nil, zbxerr.ErrorCannotMarshalJSON.Wrap(err)
	}

	return string(jsonRes), nil
}

func handlerServer(servers string, lld []lldCfgEntity) ([]lldCfgEntity, error) {
	var rs string

	hosts := servers

	h := strings.SplitN(hosts, "/", splitCount)
	if len(h) > 1 {
		rs = h[0]
		hosts = h[1]
	}

	for _, hostport := range strings.Split(hosts, ",") {
		host, _, err := net.SplitHostPort(hostport)
		if err != nil {
			return nil, zbxerr.ErrorCannotParseResult.Wrap(err)
		}

		lld = append(lld, lldCfgEntity{
			Hostname:   host,
			MongodURI:  fmt.Sprintf("%s://%s", UriDefaults.Scheme, hostport),
			ReplicaSet: rs,
		})
	}

	return lld, nil
}
