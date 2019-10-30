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
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// DataDbCfg Database config
type DataDbCfg struct {
	DataDbType         string
	DataDbHost         string // The host to connect to. Values that start with / are for UNIX domain sockets.
	DataDbPort         string // The port to bind to.
	DataDbName         string // The name of the database to connect to.
	DataDbUser         string // The user to sign in as.
	DataDbPass         string // The user's password.
	DataDbSentinelName string
	QueryTimeout       time.Duration
	RmtDataDBCfgs      []*DataDbCfg // Remote DataDB  configurations
	RplDataDBCfgs      []*DataDbCfg // Replication DataDB configurations
}

//loadFromJsonCfg loads Database config from JsonCfg
func (dbcfg *DataDbCfg) loadFromJsonCfg(jsnDbCfg *DbJsonCfg) (err error) {
	if jsnDbCfg == nil {
		return nil
	}
	if jsnDbCfg.Db_type != nil {
		dbcfg.DataDbType = strings.TrimPrefix(*jsnDbCfg.Db_type, "*")
	}
	if jsnDbCfg.Db_host != nil {
		dbcfg.DataDbHost = *jsnDbCfg.Db_host
	}
	if jsnDbCfg.Db_port != nil {
		port := strconv.Itoa(*jsnDbCfg.Db_port)
		if port == "-1" {
			port = utils.MetaDynamic
		}
		dbcfg.DataDbPort = NewDbDefaults().DBPort(dbcfg.DataDbType, port)
	}
	if jsnDbCfg.Db_name != nil {
		dbcfg.DataDbName = *jsnDbCfg.Db_name
	}
	if jsnDbCfg.Db_user != nil {
		dbcfg.DataDbUser = *jsnDbCfg.Db_user
	}
	if jsnDbCfg.Db_password != nil {
		dbcfg.DataDbPass = *jsnDbCfg.Db_password
	}
	if jsnDbCfg.Redis_sentinel != nil {
		dbcfg.DataDbSentinelName = *jsnDbCfg.Redis_sentinel
	}
	if jsnDbCfg.Query_timeout != nil {
		if dbcfg.QueryTimeout, err = utils.ParseDurationWithNanosecs(*jsnDbCfg.Query_timeout); err != nil {
			return err
		}
	}
	if jsnDbCfg.Remote_db_urls != nil {
		dbcfg.RmtDataDBCfgs = make([]*DataDbCfg, len(*jsnDbCfg.Remote_db_urls))
		for i, url := range *jsnDbCfg.Remote_db_urls {
			db, err := newDataDBCfgFromUrl(url)
			if err != nil {
				return err
			}
			dbcfg.RmtDataDBCfgs[i] = db
		}
	}
	if jsnDbCfg.Replicate_db_urls != nil {
		dbcfg.RplDataDBCfgs = make([]*DataDbCfg, len(*jsnDbCfg.Replicate_db_urls))
		for i, url := range *jsnDbCfg.Replicate_db_urls {
			db, err := newDataDBCfgFromUrl(url)
			if err != nil {
				return err
			}
			dbcfg.RplDataDBCfgs[i] = db
		}
	}
	return nil
}

// Clone returns the cloned object
func (dbcfg *DataDbCfg) Clone() *DataDbCfg {
	return &DataDbCfg{
		DataDbType:         dbcfg.DataDbType,
		DataDbHost:         dbcfg.DataDbHost,
		DataDbPort:         dbcfg.DataDbPort,
		DataDbName:         dbcfg.DataDbName,
		DataDbUser:         dbcfg.DataDbUser,
		DataDbPass:         dbcfg.DataDbPass,
		DataDbSentinelName: dbcfg.DataDbSentinelName,
		QueryTimeout:       dbcfg.QueryTimeout,
	}
}

//newDataDBCfgFromUrl will create a DataDB configuration out of url
//Format: host:port/?type=valOfType&name=valOFName&etc...
//Sample: 127.0.0.1:6379
func newDataDBCfgFromUrl(pUrl string) (newDbCfg *DataDbCfg, err error) {
	newDbCfg = new(DataDbCfg)
	if pUrl == utils.EmptyString {
		return nil, utils.ErrMandatoryIeMissing
	}
	// populate with default dataDBCfg and overwrite in case we found arguments in url
	dfltCfg, _ := NewDefaultCGRConfig()
	*newDbCfg = *dfltCfg.dataDbCfg
	hostPortSls := strings.Split(strings.Split(pUrl, utils.Slash)[0], utils.InInFieldSep)
	newDbCfg.DataDbHost = hostPortSls[0]
	newDbCfg.DataDbPort = hostPortSls[1]
	arg := utils.GetUrlRawArguments(pUrl)
	if val, has := arg[utils.TypeLow]; has {
		newDbCfg.DataDbType = strings.TrimPrefix(val, "*")
	}
	if val, has := arg[utils.UserLow]; has {
		newDbCfg.DataDbUser = val
	}
	if val, has := arg[utils.PassLow]; has {
		newDbCfg.DataDbPass = val
	}
	if val, has := arg[utils.SentinelLow]; has {
		newDbCfg.DataDbSentinelName = val
	}
	if val, has := arg[utils.QueryLow]; has {
		dur, err := utils.ParseDurationWithNanosecs(val)
		if err != nil {
			return nil, err
		}
		newDbCfg.QueryTimeout = dur
	}
	return
}
