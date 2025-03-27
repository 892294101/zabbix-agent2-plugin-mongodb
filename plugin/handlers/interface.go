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

	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.zabbix.com/sdk/log"
	"golang.zabbix.com/sdk/uri"
)

const (
	PingFailed = 0
	PingOk     = 1

	splitCount = 2

	sortNatural = "$natural"

	timestampBSONName = "ts"
)

var UriDefaults = &uri.Defaults{Scheme: "tcp", Port: "27017"}

var Logger log.Logger

var errNotFound = errors.New("not found")

// Session is an interface to access to the session struct.
type Session interface {
	DB(name string) Database
	DatabaseNames(ctx context.Context) (names []string, err error)
	Ping(ctx context.Context) error
}

type Database interface {
	C(name string) Collection
	CollectionNames(ctx context.Context) (names []string, err error)
	Run(ctx context.Context, cmd, result any) error
}

type Collection interface {
	Find(ctx context.Context, query any, opts ...*options.FindOptions) (q Query, err error)
	FindOne(ctx context.Context, query any, opts ...*options.FindOneOptions) Query
}

type Query interface {
	Count(ctx context.Context) (n int, err error)
	Get(ctx context.Context, result any) error
	GetSingle(result any) error
}
