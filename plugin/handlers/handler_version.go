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

	"go.mongodb.org/mongo-driver/bson"
	"golang.zabbix.com/sdk/zbxerr"
)

// VersionHandler executes 'buildInfo' command extracting and returning version
// info from the response.
func VersionHandler(ctx context.Context, s Session, _ map[string]string) (any, error) {
	buildInfo := bson.M{}

	err := s.DB("admin").Run(ctx, &bson.D{{Key: "buildInfo", Value: 1}}, &buildInfo)
	if err != nil {
		return nil, zbxerr.New("failed to run buildInfo command").Wrap(err)
	}

	version, ok := buildInfo["version"]
	if !ok {
		return nil, zbxerr.New("version not found in buildInfo")
	}

	return version, nil
}
