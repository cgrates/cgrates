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

	"github.com/cgrates/cgrates/utils"
)

// MigratorCgrCfg the migrator config section
type MigratorCgrCfg struct {
	OutDataDBType     string
	OutDataDBHost     string
	OutDataDBPort     string
	OutDataDBName     string
	OutDataDBUser     string
	OutDataDBPassword string
	OutDataDBEncoding string
	OutStorDBType     string
	OutStorDBHost     string
	OutStorDBPort     string
	OutStorDBName     string
	OutStorDBUser     string
	OutStorDBPassword string
	UsersFilters      []string
	OutDataDBOpts     map[string]interface{}
	OutStorDBOpts     map[string]interface{}
}

func (mg *MigratorCgrCfg) loadFromJSONCfg(jsnCfg *MigratorCfgJson) (err error) {
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
		mg.UsersFilters = utils.CloneStringSlice(*jsnCfg.Users_filters)
	}

	if jsnCfg.Out_dataDB_opts != nil {
		for k, v := range jsnCfg.Out_dataDB_opts {
			mg.OutDataDBOpts[k] = v
		}
	}
	if jsnCfg.Out_storDB_opts != nil {
		for k, v := range jsnCfg.Out_storDB_opts {
			mg.OutStorDBOpts[k] = v
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (mg *MigratorCgrCfg) AsMapInterface() (initialMP map[string]interface{}) {
	outDataDBOpts := make(map[string]interface{})
	for k, v := range mg.OutDataDBOpts {
		outDataDBOpts[k] = v
	}
	outStorDBOpts := make(map[string]interface{})
	for k, v := range mg.OutStorDBOpts {
		outStorDBOpts[k] = v
	}
	return map[string]interface{}{
		utils.OutDataDBTypeCfg:     mg.OutDataDBType,
		utils.OutDataDBHostCfg:     mg.OutDataDBHost,
		utils.OutDataDBPortCfg:     mg.OutDataDBPort,
		utils.OutDataDBNameCfg:     mg.OutDataDBName,
		utils.OutDataDBUserCfg:     mg.OutDataDBUser,
		utils.OutDataDBPasswordCfg: mg.OutDataDBPassword,
		utils.OutDataDBEncodingCfg: mg.OutDataDBEncoding,
		utils.OutStorDBTypeCfg:     mg.OutStorDBType,
		utils.OutStorDBHostCfg:     mg.OutStorDBHost,
		utils.OutStorDBPortCfg:     mg.OutStorDBPort,
		utils.OutStorDBNameCfg:     mg.OutStorDBName,
		utils.OutStorDBUserCfg:     mg.OutStorDBUser,
		utils.OutStorDBPasswordCfg: mg.OutStorDBPassword,
		utils.OutDataDBOptsCfg:     outDataDBOpts,
		utils.OutStorDBOptsCfg:     outStorDBOpts,
		utils.UsersFiltersCfg:      utils.CloneStringSlice(mg.UsersFilters),
	}
}

// Clone returns a deep copy of MigratorCgrCfg
func (mg MigratorCgrCfg) Clone() (cln *MigratorCgrCfg) {
	cln = &MigratorCgrCfg{
		OutDataDBType:     mg.OutDataDBType,
		OutDataDBHost:     mg.OutDataDBHost,
		OutDataDBPort:     mg.OutDataDBPort,
		OutDataDBName:     mg.OutDataDBName,
		OutDataDBUser:     mg.OutDataDBUser,
		OutDataDBPassword: mg.OutDataDBPassword,
		OutDataDBEncoding: mg.OutDataDBEncoding,
		OutStorDBType:     mg.OutStorDBType,
		OutStorDBHost:     mg.OutStorDBHost,
		OutStorDBPort:     mg.OutStorDBPort,
		OutStorDBName:     mg.OutStorDBName,
		OutStorDBUser:     mg.OutStorDBUser,
		OutStorDBPassword: mg.OutStorDBPassword,
		OutDataDBOpts:     make(map[string]interface{}),
		OutStorDBOpts:     make(map[string]interface{}),
	}
	if mg.UsersFilters != nil {
		cln.UsersFilters = utils.CloneStringSlice(mg.UsersFilters)
	}
	for k, v := range mg.OutDataDBOpts {
		cln.OutDataDBOpts[k] = v
	}
	for k, v := range mg.OutStorDBOpts {
		cln.OutStorDBOpts[k] = v
	}
	return
}

type MigratorCfgJson struct {
	Out_dataDB_type     *string
	Out_dataDB_host     *string
	Out_dataDB_port     *string
	Out_dataDB_name     *string
	Out_dataDB_user     *string
	Out_dataDB_password *string
	Out_dataDB_encoding *string
	Out_storDB_type     *string
	Out_storDB_host     *string
	Out_storDB_port     *string
	Out_storDB_name     *string
	Out_storDB_user     *string
	Out_storDB_password *string
	Users_filters       *[]string
	Out_dataDB_opts     map[string]interface{}
	Out_storDB_opts     map[string]interface{}
}

func diffMigratorCfgJson(d *MigratorCfgJson, v1, v2 *MigratorCgrCfg) *MigratorCfgJson {
	if d == nil {
		d = new(MigratorCfgJson)
	}
	if v1.OutDataDBType != v2.OutDataDBType {
		d.Out_dataDB_type = utils.StringPointer(v2.OutDataDBType)
	}
	if v1.OutDataDBHost != v2.OutDataDBHost {
		d.Out_dataDB_host = utils.StringPointer(v2.OutDataDBHost)
	}
	if v1.OutDataDBPort != v2.OutDataDBPort {
		d.Out_dataDB_port = utils.StringPointer(v2.OutDataDBPort)
	}
	if v1.OutDataDBName != v2.OutDataDBName {
		d.Out_dataDB_name = utils.StringPointer(v2.OutDataDBName)
	}
	if v1.OutDataDBUser != v2.OutDataDBUser {
		d.Out_dataDB_user = utils.StringPointer(v2.OutDataDBUser)
	}
	if v1.OutDataDBPassword != v2.OutDataDBPassword {
		d.Out_dataDB_password = utils.StringPointer(v2.OutDataDBPassword)
	}
	if v1.OutDataDBEncoding != v2.OutDataDBEncoding {
		d.Out_dataDB_encoding = utils.StringPointer(v2.OutDataDBEncoding)
	}
	if v1.OutStorDBType != v2.OutStorDBType {
		d.Out_storDB_type = utils.StringPointer(v2.OutStorDBType)
	}
	if v1.OutStorDBHost != v2.OutStorDBHost {
		d.Out_storDB_host = utils.StringPointer(v2.OutStorDBHost)
	}
	if v1.OutStorDBPort != v2.OutStorDBPort {
		d.Out_storDB_port = utils.StringPointer(v2.OutStorDBPort)
	}
	if v1.OutStorDBName != v2.OutStorDBName {
		d.Out_storDB_name = utils.StringPointer(v2.OutStorDBName)
	}
	if v1.OutStorDBUser != v2.OutStorDBUser {
		d.Out_storDB_user = utils.StringPointer(v2.OutStorDBUser)
	}
	if v1.OutStorDBPassword != v2.OutStorDBPassword {
		d.Out_storDB_password = utils.StringPointer(v2.OutStorDBPassword)
	}
	if !utils.SliceStringEqual(v1.UsersFilters, v2.UsersFilters) {
		d.Users_filters = utils.SliceStringPointer(utils.CloneStringSlice(v2.UsersFilters))
	}
	d.Out_dataDB_opts = diffMap(d.Out_dataDB_opts, v1.OutDataDBOpts, v2.OutDataDBOpts)
	d.Out_storDB_opts = diffMap(d.Out_storDB_opts, v1.OutStorDBOpts, v2.OutStorDBOpts)
	return d
}
