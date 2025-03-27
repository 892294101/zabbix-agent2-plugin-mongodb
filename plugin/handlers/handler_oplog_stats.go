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
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.zabbix.com/sdk/zbxerr"
)

// OplogStatsHandler
// https://docs.mongodb.com/manual/reference/method/db.getReplicationInfo/index.html
func OplogStatsHandler(ctx context.Context, s Session, _ map[string]string) (any, error) {
	var (
		err             error
		firstTs, lastTs int
	)

	localDb := s.DB("local")
	findOptions := options.FindOne()

	for _, collection := range []string{
		"oplog.rs",    // the capped collection that holds the oplog for Replica Set Members
		"oplog.$main", // oplog for the master-slave configuration
	} {
		firstTs, lastTs, err = getTS(ctx, collection, localDb, findOptions)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
			}

			continue
		}

		break
	}

	jsonRes, err := json.Marshal(
		struct {
			TimeDiff int `json:"timediff"` // in seconds
		}{
			TimeDiff: firstTs - lastTs,
		},
	)
	if err != nil {
		return nil, zbxerr.ErrorCannotMarshalJSON.Wrap(err)
	}

	return string(jsonRes), nil
}

func getTS(
	ctx context.Context,
	collection string,
	localDb Database,
	findOptions *options.FindOneOptions,
) (int, int, error) {
	findOptions.SetSort(bson.D{{Key: sortNatural, Value: -1}})

	firstTs, err := getOplogStats(ctx, localDb, collection, findOptions)
	if err != nil {
		return 0, 0, err
	}

	findOptions.SetSort(bson.D{{Key: sortNatural, Value: 1}})

	lastTs, err := getOplogStats(ctx, localDb, collection, findOptions)
	if err != nil {
		return 0, 0, err
	}

	return firstTs, lastTs, nil
}

func getOplogStats(
	ctx context.Context,
	db Database,
	collection string,
	opt *options.FindOneOptions,
) (int, error) {
	var result primitive.D

	err := db.C(collection).FindOne(ctx, bson.M{"ts": bson.M{"$exists": true}}, opt).
		GetSingle(&result)
	if err != nil {
		return 0, err
	}

	var out int

	for _, op := range result {
		if op.Key == timestampBSONName {
			if pt, ok := op.Value.(primitive.Timestamp); ok {
				out = int(time.Unix(int64(pt.T), 0).Unix())
			}
		}
	}

	return out, nil
}
