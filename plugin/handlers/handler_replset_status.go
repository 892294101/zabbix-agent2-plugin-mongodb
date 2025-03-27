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
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.zabbix.com/sdk/zbxerr"
)

const (
	statePrimary   = 1
	stateSecondary = 2
)

const nodeHealthy = 1

type Member struct {
	health int
	optime int
	state  int
	name   string
	ptr    any
}

type rawMember = map[string]any

var errUnknownStructure = errors.New("failed to parse the members structure")

func parseMembers(raw []any) ([]Member, error) {
	var (
		members     []Member
		primaryNode Member
	)

	for _, m := range raw {
		member, err := parseMember(m)
		if err != nil {
			return nil, err
		}

		if member.state == statePrimary {
			primaryNode = member
		} else {
			members = append(members, member)
		}
	}

	result := append([]Member{primaryNode}, members...)
	if len(result) == 0 {
		return nil, errUnknownStructure
	}

	return result, nil
}

func parseMember(m any) (Member, error) {
	var (
		member       Member
		extractedNum int
	)

	if v, ok := m.(rawMember)["name"].(string); ok {
		member.name = v
		extractedNum++
	}

	if v, ok := m.(rawMember)["health"].(float64); ok {
		member.health = int(v)
		extractedNum++
	}

	if v, ok := m.(rawMember)["optime"].(map[string]any); ok {
		if pa, tsOk := v["ts"].(primitive.Timestamp); tsOk {
			member.optime = int(time.Unix(int64(pa.T), 0).Unix())
		} else {
			member.optime = int(int64(v["ts"].(float64)))
		}

		extractedNum++
	}

	if v, ok := m.(rawMember)["state"].(int32); ok {
		member.state = int(v)
		extractedNum++
	}

	if extractedNum == 0 {
		return member, errUnknownStructure
	}

	member.ptr = m

	return member, nil
}

func injectExtendedMembersStats(raw []any) error {
	members, err := parseMembers(raw)
	if err != nil {
		return err
	}

	unhealthyNodes := []string{}
	unhealthyCount := 0
	primary := members[0]

	for _, node := range members {
		if ptr, ok := node.ptr.(rawMember); ok {
			ptr["lag"] = primary.optime - node.optime
			node.ptr = ptr
		}

		if node.state == stateSecondary && node.health != nodeHealthy {
			unhealthyNodes = append(unhealthyNodes, node.name)
			unhealthyCount++
		}
	}

	if ptr, ok := primary.ptr.(rawMember); ok {
		ptr["unhealthyNodes"] = unhealthyNodes
		ptr["unhealthyCount"] = unhealthyCount
		ptr["totalNodes"] = len(members) - 1
		primary.ptr = ptr
	}

	return nil
}

// ReplSetStatusHandler
// https://docs.mongodb.com/manual/reference/command/replSetGetStatus/index.html
func ReplSetStatusHandler(ctx context.Context, s Session, _ map[string]string) (any, error) {
	var replSetGetStatus map[string]any

	err := s.DB("admin").Run(
		ctx,
		&bson.D{
			{
				Key:   "replSetGetStatus",
				Value: 1,
			},
		},
		&replSetGetStatus,
	)

	if err != nil {
		if strings.Contains(err.Error(), "not running with --replSet") {
			return "{}", nil
		}

		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	if pa, ok := replSetGetStatus["members"].(primitive.A); ok {
		Logger.Debugf("members got as primitive A")

		i := []any(pa)

		Logger.Debugf("value:%v\n type: %T\n", i, i)

		err = injectExtendedMembersStats(i)
		if err != nil {
			return nil, zbxerr.ErrorCannotParseResult.Wrap(err)
		}
	}

	jsonRes, err := json.Marshal(replSetGetStatus)
	if err != nil {
		return nil, zbxerr.ErrorCannotMarshalJSON.Wrap(err)
	}

	return string(jsonRes), nil
}
