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

package config

import (
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type MigratorCgrCfg struct {
	OutDataDBType               string
	OutDataDBHost               string
	OutDataDBPort               string
	OutDataDBName               string
	OutDataDBUser               string
	OutDataDBPassword           string
	OutDataDBEncoding           string
	OutDataDBRedisSentinel      string
	OutDataDBRedisCluster       bool
	OutDataDBClusterSync        time.Duration
	OutDataDBClusterOndownDelay time.Duration
	OutStorDBType               string
	OutStorDBHost               string
	OutStorDBPort               string
	OutStorDBName               string
	OutStorDBUser               string
	OutStorDBPassword           string
	UsersFilters                []string
}

func (mg *MigratorCgrCfg) loadFromJsonCfg(jsnCfg *MigratorCfgJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Out_dataDB_type != nil {
		mg.OutDataDBType = strings.TrimPrefix(*jsnCfg.Out_dataDB_type, "*")
	}
	if jsnCfg.Out_dataDB_host != nil {
		mg.OutDataDBHost = *jsnCfg.Out_dataDB_host
	}
	if jsnCfg.Out_dataDB_port != nil {
		mg.OutDataDBPort = *jsnCfg.Out_dataDB_port
	}
	if jsnCfg.Out_dataDB_name != nil {
		mg.OutDataDBName = *jsnCfg.Out_dataDB_name
	}
	if jsnCfg.Out_dataDB_user != nil {
		mg.OutDataDBUser = *jsnCfg.Out_dataDB_user
	}
	if jsnCfg.Out_dataDB_password != nil {
		mg.OutDataDBPassword = *jsnCfg.Out_dataDB_password
	}
	if jsnCfg.Out_dataDB_encoding != nil {
		mg.OutDataDBEncoding = strings.TrimPrefix(*jsnCfg.Out_dataDB_encoding, "*")
	}
	if jsnCfg.Out_dataDB_redis_sentinel != nil {
		mg.OutDataDBRedisSentinel = *jsnCfg.Out_dataDB_redis_sentinel
	}
	if jsnCfg.Out_storDB_type != nil {
		mg.OutStorDBType = *jsnCfg.Out_storDB_type
	}
	if jsnCfg.Out_storDB_host != nil {
		mg.OutStorDBHost = *jsnCfg.Out_storDB_host
	}
	if jsnCfg.Out_storDB_port != nil {
		mg.OutStorDBPort = *jsnCfg.Out_storDB_port
	}
	if jsnCfg.Out_storDB_name != nil {
		mg.OutStorDBName = *jsnCfg.Out_storDB_name
	}
	if jsnCfg.Out_storDB_user != nil {
		mg.OutStorDBUser = *jsnCfg.Out_storDB_user
	}
	if jsnCfg.Out_storDB_password != nil {
		mg.OutStorDBPassword = *jsnCfg.Out_storDB_password
	}
	if jsnCfg.Users_filters != nil && len(*jsnCfg.Users_filters) != 0 {
		mg.UsersFilters = make([]string, len(*jsnCfg.Users_filters))
		for i, v := range *jsnCfg.Users_filters {
			mg.UsersFilters[i] = v
		}
	}

	if jsnCfg.Out_dataDB_redis_cluster != nil {
		mg.OutDataDBRedisCluster = *jsnCfg.Out_dataDB_redis_cluster
	}
	if jsnCfg.Out_dataDB_cluster_ondown_delay != nil {
		if mg.OutDataDBClusterOndownDelay, err = utils.ParseDurationWithNanosecs(*jsnCfg.Out_dataDB_cluster_ondown_delay); err != nil {
			return err
		}
	}
	if jsnCfg.Out_dataDB_cluster_sync != nil {
		if mg.OutDataDBClusterSync, err = utils.ParseDurationWithNanosecs(*jsnCfg.Out_dataDB_cluster_sync); err != nil {
			return err
		}
	}
	return nil
}

func (mg *MigratorCgrCfg) AsMapInterface() map[string]interface{} {
	return map[string]interface{}{
		utils.OutDataDBTypeCfg:          mg.OutDataDBType,
		utils.OutDataDBHostCfg:          mg.OutDataDBHost,
		utils.OutDataDBPortCfg:          mg.OutDataDBPort,
		utils.OutDataDBNameCfg:          mg.OutDataDBName,
		utils.OutDataDBUserCfg:          mg.OutDataDBUser,
		utils.OutDataDBPasswordCfg:      mg.OutDataDBPassword,
		utils.OutDataDBEncodingCfg:      mg.OutDataDBEncoding,
		utils.OutDataDBRedisSentinelCfg: mg.OutDataDBRedisSentinel,
		utils.OutStorDBTypeCfg:          mg.OutStorDBType,
		utils.OutStorDBHostCfg:          mg.OutStorDBHost,
		utils.OutStorDBPortCfg:          mg.OutStorDBPort,
		utils.OutStorDBNameCfg:          mg.OutStorDBName,
		utils.OutStorDBUserCfg:          mg.OutStorDBUser,
		utils.OutStorDBPasswordCfg:      mg.OutStorDBPassword,
		utils.UsersFiltersCfg:           mg.UsersFilters,
	}

}
