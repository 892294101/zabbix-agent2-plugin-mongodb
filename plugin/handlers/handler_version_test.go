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
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/bson"
)

func TestVersionHandler(t *testing.T) {
	t.Parallel()

	sampleResp := bson.M{
		"version":           "4.4.0",
		"debug":             false,
		"gitVersion":        "90c65f9cc8fc4e6664a5848230abaa9b3f3b02f7",
		"javascriptEngine":  "mozjs",
		"maxBsonObjectSize": 16777216,
		"ok":                1,
	}

	type db struct {
		resp bson.M
		err  error
	}

	tests := []struct {
		name    string
		db      db
		want    any
		wantErr bool
	}{
		{
			"+valid",
			db{sampleResp, nil},
			any("4.4.0"),
			false,
		},
		{
			"-commandErr",
			db{sampleResp, errors.New("fail")},
			nil,
			true,
		},
		{
			"-missingVersion",
			db{
				bson.M{
					"debug":             false,
					"gitVersion":        "90c65f9cc8fc4e6664a5848230abaa9b3f3b02f7",
					"javascriptEngine":  "mozjs",
					"maxBsonObjectSize": 16777216,
					"ok":                1,
				},
				nil,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSess := &MockConn{
				dbs: map[string]*MockMongoDatabase{
					"admin": {
						RunFunc: func(_, _ string) ([]byte, error) {
							if tt.db.err != nil {
								return nil, tt.db.err
							}

							b, err := bson.Marshal(tt.db.resp)
							if err != nil {
								t.Fatalf("failed to marshal response: %v", err)
							}

							return b, nil
						},
					},
				},
			}

			got, err := VersionHandler(context.Background(), mockSess, nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"VersionHandler() error = %v, wantErr %v", err, tt.wantErr,
				)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("VersionHandler() = %s", diff)
			}
		})
	}
}
