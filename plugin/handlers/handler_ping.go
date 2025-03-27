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

import "context"

// PingHandler executes 'ping' command and returns pingOk if a connection is alive or pingFailed otherwise.
// https://docs.mongodb.com/manual/reference/command/ping/index.html
func PingHandler(ctx context.Context, s Session, _ map[string]string) (any, error) {
	if err := s.Ping(ctx); err != nil {
		Logger.Debugf("ping failed, %s", err.Error())

		return PingFailed, nil
	}

	return PingOk, nil
}
