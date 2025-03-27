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
)

func Test_replSetConfigHandler(t *testing.T) {
	var testData map[string]any

	jsonData, err := ioutil.ReadFile("testdata/replSetGetConfig.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonData, &testData)
	if err != nil {
		log.Fatal(err)
	}

	mockSession := NewMockConn()
	db := mockSession.DB("admin")
	db.(*MockMongoDatabase).RunFunc = func(dbName, cmd string) ([]byte, error) {
		if cmd == "replSetGetConfig" {
			return bson.Marshal(testData)
		}

		return nil, errors.New("no such cmd: " + cmd)
	}

	type args struct {
		s Session
	}

	tests := []struct {
		name    string
		args    args
		want    any
		wantErr error
	}{
		{
			name: "Must parse an output of \" + replSetGetConfig + \"command",
			args: args{
				s: mockSession,
			},
			want:    strings.TrimSpace(string(jsonData)),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplSetConfigHandler(context.Background(), tt.args.s, nil)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("replSetConfigHandler() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("replSetConfigHandler() got = %v, want %v", got, tt.want)
			}
		})
	}
}
