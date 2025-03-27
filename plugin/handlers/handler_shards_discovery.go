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
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.zabbix.com/sdk/zbxerr"
)

type lldShEntity struct {
	ID        string `json:"{#ID}"`
	Hostname  string `json:"{#HOSTNAME}"`
	MongodURI string `json:"{#MONGOD_URI}"`
	State     string `json:"{#STATE}"`
}

type shEntry struct {
	ID    string      `bson:"_id"`
	Host  string      `bson:"host"`
	State json.Number `bson:"state"`
}

// ShardsDiscoveryHandler
// https://docs.mongodb.com/manual/reference/method/sh.status/#sh.status
func ShardsDiscoveryHandler(ctx context.Context, s Session, _ map[string]string) (any, error) {
	var shards []shEntry

	opts := options.Find()
	opts.SetSort(bson.D{{Key: sortNatural, Value: 1}})

	q, err := s.DB("config").C("shards").Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	err = q.Get(ctx, &shards)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	lld := make([]lldShEntity, 0)

	for _, sh := range shards {
		lld, err = handlerShards(sh, lld)
		if err != nil {
			return nil, zbxerr.ErrorCannotParseResult.Wrap(err)
		}
	}

	jsonLLD, err := json.Marshal(lld)
	if err != nil {
		return nil, zbxerr.ErrorCannotMarshalJSON.Wrap(err)
	}

	return string(jsonLLD), nil
}

func handlerShards(sh shEntry, lld []lldShEntity) ([]lldShEntity, error) {
	hosts := sh.Host

	h := strings.SplitN(sh.Host, "/", splitCount)
	if len(h) > 1 {
		hosts = h[1]
	}

	for _, hostport := range strings.Split(hosts, ",") {
		host, _, err := net.SplitHostPort(hostport)
		if err != nil {
			return nil, err
		}

		lld = append(lld, lldShEntity{
			ID:        sh.ID,
			Hostname:  host,
			MongodURI: fmt.Sprintf("%s://%s", UriDefaults.Scheme, hostport),
			State:     sh.State.String(),
		})
	}

	return lld, nil
}
