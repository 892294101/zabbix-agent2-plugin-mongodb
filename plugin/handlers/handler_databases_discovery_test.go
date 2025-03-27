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
	"reflect"
	"testing"

	"golang.zabbix.com/sdk/zbxerr"
)

func Test_databasesDiscoveryHandler(t *testing.T) {
	type args struct {
		s   Session
		dbs []string
	}

	tests := []struct {
		name    string
		args    args
		want    any
		wantErr error
	}{
		{
			name: "Must return a list of databases",
			args: args{
				s:   NewMockConn(),
				dbs: []string{"testdb", "local", "config"},
			},
			want:    "[{\"{#DBNAME}\":\"config\"},{\"{#DBNAME}\":\"local\"},{\"{#DBNAME}\":\"testdb\"}]",
			wantErr: nil,
		},
		{
			name: "Must catch DB.DatabaseNames() error",
			args: args{
				s:   NewMockConn(),
				dbs: []string{mustFail},
			},
			want:    nil,
			wantErr: zbxerr.ErrorCannotFetchData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, db := range tt.args.dbs {
				tt.args.s.DB(db)
			}

			got, err := DatabasesDiscoveryHandler(context.Background(), tt.args.s, nil)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("databasesDiscoveryHandler() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("databasesDiscoveryHandler() got = %v, want %v", got, tt.want)
			}
		})
	}
}
