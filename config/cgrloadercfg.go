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

func NewLoaderConfig(cfgPath *string) (lCfg *LoaderCfg, err error) {
	lCfg = NewDefaultLoaderConfig()
	c, err := goconf.ReadConfigFile(*cfgPath)
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
	DataDBEncoding  string
	StorDBType      string
	StorDBHost      string
	StorDBPort      string
	StorDBName      string
	StorDBUser      string
	StorDBPass      string
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
	//dataDB
	if hasOpt = c.HasOption("dataDB", "db_type"); hasOpt {
		ldrCfg.DataDBType, err = c.GetString("dataDB", "db_type")
	}
	if hasOpt = c.HasOption("dataDB", "db_host"); hasOpt {
		ldrCfg.DataDBHost, err = c.GetString("dataDB", "db_host")
	}
	if hasOpt = c.HasOption("dataDB", "db_port"); hasOpt {
		ldrCfg.DataDBPort, err = c.GetString("dataDB", "db_port")
	}
	if hasOpt = c.HasOption("dataDB", "db_name"); hasOpt {
		ldrCfg.DataDBName, err = c.GetString("dataDB", "db_name")
	}
	if hasOpt = c.HasOption("dataDB", "db_user"); hasOpt {
		ldrCfg.DataDBUser, err = c.GetString("dataDB", "db_user")
	}
	if hasOpt = c.HasOption("dataDB", "db_password"); hasOpt {
		ldrCfg.DataDBPass, err = c.GetString("dataDB", "db_password")
	}
	//storDB
	if hasOpt = c.HasOption("storDB", "db_type"); hasOpt {
		ldrCfg.StorDBType, err = c.GetString("storDB", "db_type")
	}
	if hasOpt = c.HasOption("storDB", "db_host"); hasOpt {
		ldrCfg.StorDBHost, err = c.GetString("storDB", "db_host")
	}
	if hasOpt = c.HasOption("storDB", "db_port"); hasOpt {
		ldrCfg.StorDBPort, err = c.GetString("storDB", "db_port")
	}
	if hasOpt = c.HasOption("storDB", "db_name"); hasOpt {
		ldrCfg.StorDBName, err = c.GetString("storDB", "db_name")
	}
	if hasOpt = c.HasOption("storDB", "db_user"); hasOpt {
		ldrCfg.StorDBUser, err = c.GetString("storDB", "db_user")
	}
	if hasOpt = c.HasOption("storDB", "db_password"); hasOpt {
		ldrCfg.StorDBPass, err = c.GetString("storDB", "db_password")
	}
	//general
	if hasOpt = c.HasOption("general", "tpid"); hasOpt {
		ldrCfg.Tpid, err = c.GetString("general", "tpid")
	}
	if hasOpt = c.HasOption("general", "dataPath"); hasOpt {
		ldrCfg.DataPath, err = c.GetString("general", "dataPath")
	}
	if hasOpt = c.HasOption("general", "rpcEncoding"); hasOpt {
		ldrCfg.RpcEncoding, err = c.GetString("general", "rpcEncoding")
	}
	if hasOpt = c.HasOption("general", "ralsAddress"); hasOpt {
		ldrCfg.RalsAddress, err = c.GetString("general", "ralsAddress")
	}
	if hasOpt = c.HasOption("general", "cdrstatsAddress"); hasOpt {
		ldrCfg.CdrstatsAddress, err = c.GetString("general", "cdrstatsAddress")
	}
	if hasOpt = c.HasOption("general", "usersAddress"); hasOpt {
		ldrCfg.UsersAddress, err = c.GetString("general", "usersAddress")
	}
	if hasOpt = c.HasOption("general", "runId"); hasOpt {
		ldrCfg.RunId, err = c.GetString("general", "runId")
	}
	if hasOpt = c.HasOption("general", "loadHistorySize"); hasOpt {
		ldrCfg.LoadHistorySize, err = c.GetInt("general", "loadHistorySize")
	}
	if hasOpt = c.HasOption("general", "timezone"); hasOpt {
		ldrCfg.Timezone, err = c.GetString("general", "timezone")
	}
	if hasOpt = c.HasOption("general", "disable_reverse"); hasOpt {
		ldrCfg.DisableReverse, err = c.GetBool("general", "disable_reverse")
	}

	return err
}
