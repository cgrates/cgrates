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

// StorDbCfg StroreDb config
type StorDbCfg struct {
	Type                string // Should reflect the database type used to store logs
	Host                string // The host to connect to. Values that start with / are for UNIX domain sockets.
	Port                string // Th e port to bind to.
	Name                string // The name of the database to connect to.
	User                string // The user to sign in as.
	Password            string // The user's password.
	MaxOpenConns        int    // Maximum database connections opened
	MaxIdleConns        int    // Maximum idle connections to keep opened
	ConnMaxLifetime     int
	StringIndexedFields []string
	PrefixIndexedFields []string
	QueryTimeout        time.Duration
	SSLMode             string // for PostgresDB used to change default sslmode
}

// loadFromJsonCfg loads StoreDb config from JsonCfg
func (dbcfg *StorDbCfg) loadFromJsonCfg(jsnDbCfg *DbJsonCfg) (err error) {
	if jsnDbCfg == nil {
		return nil
	}
	if jsnDbCfg.Db_type != nil {
		dbcfg.Type = strings.TrimPrefix(*jsnDbCfg.Db_type, "*")
	}
	if jsnDbCfg.Db_host != nil {
		dbcfg.Host = *jsnDbCfg.Db_host
	}
	if jsnDbCfg.Db_port != nil {
		port := strconv.Itoa(*jsnDbCfg.Db_port)
		if port == "-1" {
			port = utils.MetaDynamic
		}
		dbcfg.Port = dbDefaultsCfg.dbPort(dbcfg.Type, port)
	}
	if jsnDbCfg.Db_name != nil {
		dbcfg.Name = *jsnDbCfg.Db_name
	}
	if jsnDbCfg.Db_user != nil {
		dbcfg.User = *jsnDbCfg.Db_user
	}
	if jsnDbCfg.Db_password != nil {
		dbcfg.Password = *jsnDbCfg.Db_password
	}
	if jsnDbCfg.Max_open_conns != nil {
		dbcfg.MaxOpenConns = *jsnDbCfg.Max_open_conns
	}
	if jsnDbCfg.Max_idle_conns != nil {
		dbcfg.MaxIdleConns = *jsnDbCfg.Max_idle_conns
	}
	if jsnDbCfg.Conn_max_lifetime != nil {
		dbcfg.ConnMaxLifetime = *jsnDbCfg.Conn_max_lifetime
	}
	if jsnDbCfg.String_indexed_fields != nil {
		dbcfg.StringIndexedFields = *jsnDbCfg.String_indexed_fields
	}
	if jsnDbCfg.Prefix_indexed_fields != nil {
		dbcfg.PrefixIndexedFields = *jsnDbCfg.Prefix_indexed_fields
	}
	if jsnDbCfg.Query_timeout != nil {
		if dbcfg.QueryTimeout, err = utils.ParseDurationWithNanosecs(*jsnDbCfg.Query_timeout); err != nil {
			return err
		}
	}
	if jsnDbCfg.Sslmode != nil {
		dbcfg.SSLMode = *jsnDbCfg.Sslmode
	}
	return nil
}

// Clone returns the cloned object
func (dbcfg *StorDbCfg) Clone() *StorDbCfg {
	return &StorDbCfg{
		Type:                dbcfg.Type,
		Host:                dbcfg.Host,
		Port:                dbcfg.Port,
		Name:                dbcfg.Name,
		User:                dbcfg.User,
		Password:            dbcfg.Password,
		MaxOpenConns:        dbcfg.MaxOpenConns,
		MaxIdleConns:        dbcfg.MaxIdleConns,
		ConnMaxLifetime:     dbcfg.ConnMaxLifetime,
		StringIndexedFields: dbcfg.StringIndexedFields,
		PrefixIndexedFields: dbcfg.PrefixIndexedFields,
		QueryTimeout:        dbcfg.QueryTimeout,
		SSLMode:             dbcfg.SSLMode,
	}
}
