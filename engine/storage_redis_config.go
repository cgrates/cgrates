// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func (rs *RedisStorage) GetSection(ctx *context.Context, section string, val any) (err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, utils.ConfigPrefix+section); err != nil || len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, val)
	return
}

func (rs *RedisStorage) SetSection(_ *context.Context, section string, jsn any) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(jsn); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, utils.ConfigPrefix+section, string(result))
}

// Only intended for InternalDB
func (rs *RedisStorage) DumpConfigDB() (err error) {
	return utils.ErrNotImplemented
}

// Only intended for InternalDB
func (rs *RedisStorage) RewriteConfigDB() (err error) {
	return utils.ErrNotImplemented
}

// Only intended for InternalDB
func (rs *RedisStorage) BackupConfigDB(backupFolderPath string, zip bool) (err error) {
	return utils.ErrNotImplemented
}
