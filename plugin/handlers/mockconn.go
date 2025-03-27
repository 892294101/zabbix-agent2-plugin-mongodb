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
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.zabbix.com/sdk/zbxerr"
)

const (
	mustFail      = "mustFail"
	mockTimestamp = 3000
)

type MockConn struct {
	dbs map[string]*MockMongoDatabase
}

func NewMockConn() *MockConn {
	return &MockConn{
		dbs: make(map[string]*MockMongoDatabase),
	}
}

func (conn *MockConn) DB(name string) Database {
	if db, ok := conn.dbs[name]; ok {
		return db
	}

	conn.dbs[name] = &MockMongoDatabase{
		name:        name,
		collections: make(map[string]*MockMongoCollection),
	}

	return conn.dbs[name]
}

// DatabaseNames returns mock DB names.
func (conn *MockConn) DatabaseNames(_ context.Context) ([]string, error) {
	names := make([]string, 0, len(conn.dbs))

	for _, db := range conn.dbs {
		if db.name == mustFail {
			return nil, zbxerr.ErrorCannotFetchData
		}

		names = append(names, db.name)
	}

	return names, nil
}

// Ping does nothing as a mock function.
func (*MockConn) Ping(_ context.Context) error {
	return nil
}

type MockSession interface {
	DB(name string) Database
	DatabaseNames(ctx context.Context) ([]string, error)
	Ping(_ context.Context) error
}

type MockMongoDatabase struct {
	name        string
	collections map[string]*MockMongoCollection
	RunFunc     func(dbName, cmd string) ([]byte, error)
}

func (d *MockMongoDatabase) C(name string) Collection {
	if col, ok := d.collections[name]; ok {
		return col
	}

	d.collections[name] = &MockMongoCollection{
		name:    name,
		queries: make(map[any]*MockMongoQuery),
	}

	return d.collections[name]
}

// CollectionNames returns all collection names.
func (d *MockMongoDatabase) CollectionNames(_ context.Context) ([]string, error) {
	names := make([]string, 0, len(d.collections))

	for _, col := range d.collections {
		if col.name == mustFail {
			return nil, errors.New("fail")
		}

		names = append(names, col.name)
	}

	return names, nil
}

// Run executed command with given mock function.
func (d *MockMongoDatabase) Run(_ context.Context, cmd, result any) error {
	if d.RunFunc == nil {
		d.RunFunc = func(dbName, _ string) ([]byte, error) {
			if dbName == mustFail {
				return nil, errors.New("fail")
			}

			return bson.Marshal(map[string]int{"ok": 1})
		}
	}

	if result == nil {
		return nil
	}

	bsonDcmd := *(cmd.(*bson.D))
	cmdName := bsonDcmd[0].Key

	data, err := d.RunFunc(d.name, cmdName)
	if err != nil {
		return err
	}

	return bson.Unmarshal(data, result)
}

type MockMongoCollection struct {
	name    string
	queries map[any]*MockMongoQuery
}

// Find retrieves documents matching query.
//
//nolint:ireturn,nolintlint
func (c *MockMongoCollection) Find(
	_ context.Context,
	query any,
	_ ...*options.FindOptions,
) (Query, error) {
	queryHash := fmt.Sprintf("%v", query)
	if q, ok := c.queries[queryHash]; ok {
		return q, nil
	}

	c.queries[queryHash] = &MockMongoQuery{
		collection: c.name,
		query:      query,
	}

	return c.queries[queryHash], nil
}

// FindOne retrieves single document matching query.
//
//nolint:ireturn,nolintlint
func (c *MockMongoCollection) FindOne(
	_ context.Context,
	query any,
	_ ...*options.FindOneOptions,
) Query {
	queryHash := fmt.Sprintf("%v", query)
	if q, ok := c.queries[queryHash]; ok {
		return q
	}

	c.queries[queryHash] = &MockMongoQuery{
		collection: c.name,
		query:      query,
	}

	return c.queries[queryHash]
}

type MockMongoQuery struct {
	collection string
	query      any
	DataFunc   func() ([]byte, error)
}

func (q *MockMongoQuery) retrieve(result any) error {
	if q.DataFunc == nil {
		return errNotFound
	}

	if result == nil {
		return nil
	}

	data, err := q.DataFunc()
	if err != nil {
		return err
	}

	return bson.Unmarshal(data, result)
}

// Count mock function, retrieves fake document count (always 1).
func (*MockMongoQuery) Count(_ context.Context) (int, error) {
	return 1, nil
}

// Get mock function, retrieves fake result.
func (q *MockMongoQuery) Get(_ context.Context, result any) error {
	return q.retrieve(result)
}

// GetSingle mock function, retrieves fake single result.
func (q *MockMongoQuery) GetSingle(result any) error {
	return q.retrieve(result)
}
