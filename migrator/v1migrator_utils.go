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

package migrator

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/cgrates/cgrates/utils"
)

func NewMigratorDataDB(db_type, host, port, name, user, pass, marshaler string) {

}

func ConfigureV1DataStorage() (db MigratorDataDB, err error) {
	var d MigratorDataDB
	d.DataManger().(engine.RedisStoarage)
	switch db_type {
	case utils.REDIS:
		var db_nb int
		db_nb, err = strconv.Atoi(name)
		if err != nil {
			utils.Logger.Crit("Redis db name must be an integer!")
			return nil, err
		}
		if port != "" {
			host += ":" + port
		}
		d, err = newv1RedisStorage(host, db_nb, pass, marshaler)
	case utils.MONGO:
		d, err = newv1MongoStorage(host, port, name, user, pass, utils.DataDB, nil)
		db = d.(MigratorDataDB)
	default:
		err = errors.New(fmt.Sprintf("Unknown db '%s' valid options are '%s' or '%s'",
			db_type, utils.REDIS, utils.MONGO))
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

func ConfigureV1StorDB(db_type, host, port, name, user, pass string) (db MigratorStorDB, err error) {
	var d MigratorStorDB
	switch db_type {
	case utils.MONGO:
		d, err = newv1MongoStorage(host, port, name, user, pass, utils.StorDB, nil)
		db = d.(MigratorStorDB)
	case utils.MYSQL:
		d, err = newSqlStorage(host, port, name, user, pass)
		db = d.(MigratorStorDB)
	default:
		err = errors.New(fmt.Sprintf("Unknown db '%s' valid options are '%s'",
			db_type, utils.MONGO))
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}
