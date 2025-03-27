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

	"go.mongodb.org/mongo-driver/bson"
	"golang.zabbix.com/sdk/zbxerr"
)

// ReplSetConfigHandler
// https://docs.mongodb.com/manual/reference/command/replSetGetConfig/index.html
func ReplSetConfigHandler(ctx context.Context, s Session, _ map[string]string) (any, error) {
	replSetGetConfig := &bson.M{}
	err := s.DB("admin").Run(
		ctx,
		&bson.D{
			{
				Key:   "replSetGetConfig",
				Value: 1,
			},
			{
				Key:   "commitmentStatus",
				Value: true,
			},
		},
		replSetGetConfig,
	)

	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	jsonRes, err := json.Marshal(replSetGetConfig)
	if err != nil {
		return nil, zbxerr.ErrorCannotMarshalJSON.Wrap(err)
	}

	return string(jsonRes), nil
}
