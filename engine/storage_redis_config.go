/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"github.com/cgrates/birpc/context"
)

const (
	configPrefix = "cfg_"
)

func (rs *RedisStorage) GetSection(ctx *context.Context, section string, val interface{}) (err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+section); err != nil || len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, val)
	return
}

func (rs *RedisStorage) SetSection(_ *context.Context, section string, jsn interface{}) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(jsn); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, configPrefix+section, string(result))
}
