/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package rater

import (
	"strconv"
	"errors"
)

// Various helpers to deal with database

func ConfigureDatabase(db_type, host, port, name, user, pass string) (db DataStorage, err error) {
	switch db_type {
	case REDIS:
		var db_nb int
		db_nb, err = strconv.Atoi(name)
		if err != nil {
			Logger.Crit("Redis db name must be an integer!")
			return nil, err
		}
		if port != "" {
			host += ":" + port
		}
		db, err = NewRedisStorage(host, db_nb, pass)
	case MONGO:
		db, err = NewMongoStorage(host, port, name, user, pass)
	case POSTGRES:
		db, err = NewPostgresStorage(host, port, name, user, pass)
	case MYSQL:
		db, err = NewMySQLStorage(host, port, name, user, pass)
	default:
		err = errors.New("unknown db")
	}
	if err != nil {
		return nil, err
	}
	return db, nil
}

