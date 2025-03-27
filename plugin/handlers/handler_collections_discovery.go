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
	"sort"

	"golang.zabbix.com/sdk/zbxerr"
)

type colEntity struct {
	ColName string `json:"{#COLLECTION}"`
	DbName  string `json:"{#DBNAME}"`
}

// CollectionsDiscoveryHandler
// https://docs.mongodb.com/manual/reference/command/listDatabases/
func CollectionsDiscoveryHandler(ctx context.Context, s Session, _ map[string]string) (any, error) {
	dbs, err := s.DatabaseNames(ctx)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	sort.Strings(dbs)

	lld := make([]colEntity, 0)

	for _, db := range dbs {
		var collections []string
		collections, err = s.DB(db).CollectionNames(ctx)

		sort.Strings(collections)

		if err != nil {
			return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
		}

		for _, col := range collections {
			lld = append(lld, colEntity{
				ColName: col,
				DbName:  db,
			})
		}
	}

	jsonLLD, err := json.Marshal(lld)
	if err != nil {
		return nil, zbxerr.ErrorCannotMarshalJSON.Wrap(err)
	}

	return string(jsonLLD), nil
}
