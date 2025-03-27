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
	"io/ioutil"
	"log"
	"reflect"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"golang.zabbix.com/sdk/zbxerr"
)

func Test_collectionStatsHandler(t *testing.T) {
	var testData map[string]any

	jsonData, err := ioutil.ReadFile("testdata/collStats.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonData, &testData)
	if err != nil {
		log.Fatal(err)
	}

	mockSession := NewMockConn()
	db := mockSession.DB("MyDatabase")
	db.(*MockMongoDatabase).RunFunc = func(dbName, cmd string) ([]byte, error) {
		if cmd == "collStats" {
			return bson.Marshal(testData)
		}

		return nil, errors.New("no such cmd: " + cmd)
	}

	type args struct {
		s      Session
		params map[string]string
	}

	tests := []struct {
		name    string
		args    args
		want    any
		wantErr error
	}{
		{
			name: "Must parse an output of \" + collStats + \"command",
			args: args{
				s:      mockSession,
				params: map[string]string{"Database": "MyDatabase", "Collection": "MyCollection"},
			},
			want:    strings.TrimSpace(string(jsonData)),
			wantErr: nil,
		},
		{
			name: "Must catch DB.Run() error",
			args: args{
				s:      mockSession,
				params: map[string]string{"Database": mustFail},
			},
			want:    nil,
			wantErr: zbxerr.ErrorCannotFetchData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CollectionStatsHandler(context.Background(), tt.args.s, tt.args.params)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("collectionStatsHandler() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("collectionStatsHandler() got = %v, want %v", got, tt.want)
			}
		})
	}
}
