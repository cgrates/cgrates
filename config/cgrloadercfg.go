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
	"errors"
	"fmt"

	"github.com/cgrates/cgrates/utils"
	"github.com/dlintw/goconf"
)

func NewLoaderConfig(cfgPath string) (lCfg *LoaderCfg, err error) {
	lCfg = NewDefaultLoaderConfig()
	c, err := goconf.ReadConfigFile(cfgPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not open the configuration file: %s", err))
	}
	if err = lCfg.loadConfig(c); err != nil {
		return lCfg, err
	}
	return
}

func NewDefaultLoaderConfig() (lCfg *LoaderCfg) {
	lCfg = new(LoaderCfg)
	lCfg.setDefaults()
	return lCfg
}

type LoaderCfg struct {
	DataDBType      string
	DataDBHost      string
	DataDBPort      string
	DataDBName      string
	DataDBUser      string
	DataDBPass      string
	StorDBType      string
	StorDBHost      string
	StorDBPort      string
	StorDBName      string
	StorDBUser      string
	StorDBPass      string
	DBDataEncoding  string
	Flush           bool
	Tpid            string
	DataPath        string
	Version         bool
	Verbose         bool
	DryRun          bool
	Validate        bool
	Stats           bool
	FromStorDB      bool
	ToStorDB        bool
	RpcEncoding     string
	RalsAddress     string
	CdrstatsAddress string
	UsersAddress    string
	RunId           string
	LoadHistorySize int
	Timezone        string
	DisableReverse  bool
	FlushStorDB     bool
	Remove          bool
}

func (lCfg *LoaderCfg) setDefaults() {
	lCfg.DataDBType = cgrCfg.DataDbType
	lCfg.DataDBHost = utils.MetaDynamic
	lCfg.DataDBPort = utils.MetaDynamic
	lCfg.DataDBName = utils.MetaDynamic
	lCfg.DataDBUser = utils.MetaDynamic
	lCfg.DataDBPass = utils.MetaDynamic
	lCfg.StorDBType = cgrCfg.StorDBType
	lCfg.StorDBHost = utils.MetaDynamic
	lCfg.StorDBPort = utils.MetaDynamic
	lCfg.StorDBName = utils.MetaDynamic
	lCfg.StorDBUser = utils.MetaDynamic
	lCfg.StorDBPass = utils.MetaDynamic
	lCfg.DBDataEncoding = cgrCfg.DBDataEncoding
	lCfg.Flush = false
	lCfg.Tpid = ""
	lCfg.DataPath = "./"
	lCfg.Version = false
	lCfg.Verbose = false
	lCfg.DryRun = false
	lCfg.Validate = false
	lCfg.Stats = false
	lCfg.FromStorDB = false
	lCfg.ToStorDB = false
	lCfg.RpcEncoding = "json"
	lCfg.RalsAddress = cgrCfg.RPCJSONListen
	lCfg.CdrstatsAddress = cgrCfg.RPCJSONListen
	lCfg.UsersAddress = cgrCfg.RPCJSONListen
	lCfg.RunId = ""
	lCfg.LoadHistorySize = cgrCfg.LoadHistorySize
	lCfg.Timezone = cgrCfg.DefaultTimezone
	lCfg.DisableReverse = false
	lCfg.FlushStorDB = false
	lCfg.Remove = false
}

func (ldrCfg *LoaderCfg) loadConfig(c *goconf.ConfigFile) (err error) {
	var hasOpt bool
	//data_db
	if hasOpt = c.HasOption("data_db", "db_type"); hasOpt {
		ldrCfg.DataDBType, err = c.GetString("data_db", "db_type")
	}
	if hasOpt = c.HasOption("data_db", "db_host"); hasOpt {
		ldrCfg.DataDBHost, err = c.GetString("data_db", "db_host")
	}
	if hasOpt = c.HasOption("data_db", "db_port"); hasOpt {
		ldrCfg.DataDBPort, err = c.GetString("data_db", "db_port")
	}
	if hasOpt = c.HasOption("data_db", "db_name"); hasOpt {
		ldrCfg.DataDBName, err = c.GetString("data_db", "db_name")
	}
	if hasOpt = c.HasOption("data_db", "db_user"); hasOpt {
		ldrCfg.DataDBUser, err = c.GetString("data_db", "db_user")
	}
	if hasOpt = c.HasOption("data_db", "db_password"); hasOpt {
		ldrCfg.DataDBPass, err = c.GetString("data_db", "db_password")
	}
	//stor_db
	if hasOpt = c.HasOption("stor_db", "db_type"); hasOpt {
		ldrCfg.StorDBType, err = c.GetString("stor_db", "db_type")
	}
	if hasOpt = c.HasOption("stor_db", "db_host"); hasOpt {
		ldrCfg.StorDBHost, err = c.GetString("stor_db", "db_host")
	}
	if hasOpt = c.HasOption("stor_db", "db_port"); hasOpt {
		ldrCfg.StorDBPort, err = c.GetString("stor_db", "db_port")
	}
	if hasOpt = c.HasOption("stor_db", "db_name"); hasOpt {
		ldrCfg.StorDBName, err = c.GetString("stor_db", "db_name")
	}
	if hasOpt = c.HasOption("stor_db", "db_user"); hasOpt {
		ldrCfg.StorDBUser, err = c.GetString("stor_db", "db_user")
	}
	if hasOpt = c.HasOption("stor_db", "db_password"); hasOpt {
		ldrCfg.StorDBPass, err = c.GetString("stor_db", "db_password")
	}
	//general
	if hasOpt = c.HasOption("general", "dbdata_encoding"); hasOpt {
		ldrCfg.DBDataEncoding, err = c.GetString("general", "dbdata_encoding")
	}
	if hasOpt = c.HasOption("general", "tpid"); hasOpt {
		ldrCfg.Tpid, err = c.GetString("general", "tpid")
	}
	if hasOpt = c.HasOption("general", "data_path"); hasOpt {
		ldrCfg.DataPath, err = c.GetString("general", "data_path")
	}
	if hasOpt = c.HasOption("general", "rpc_encoding"); hasOpt {
		ldrCfg.RpcEncoding, err = c.GetString("general", "rpc_encoding")
	}
	if hasOpt = c.HasOption("general", "rals_address"); hasOpt {
		ldrCfg.RalsAddress, err = c.GetString("general", "rals_address")
	}
	if hasOpt = c.HasOption("general", "runid"); hasOpt {
		ldrCfg.RunId, err = c.GetString("general", "runid")
	}
	if hasOpt = c.HasOption("general", "load_history_size"); hasOpt {
		ldrCfg.LoadHistorySize, err = c.GetInt("general", "load_history_size")
	}
	if hasOpt = c.HasOption("general", "timezone"); hasOpt {
		ldrCfg.Timezone, err = c.GetString("general", "timezone")
	}
	if hasOpt = c.HasOption("general", "disable_reverse"); hasOpt {
		ldrCfg.DisableReverse, err = c.GetBool("general", "disable_reverse")
	}

	return err
}
