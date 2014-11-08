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

package engine

import (
	"fmt"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"path"

	_ "github.com/bmizerany/pq"
	"github.com/jinzhu/gorm"
)

type PostgresStorage struct {
	*SQLStorage
}

func NewPostgresStorage(host, port, name, user, password string) (Storage, error) {
	connectString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", host, port, name, user, password)
	db, err := gorm.Open("postgres", connectString)
	if err != nil {
		return nil, err
	}
	err = db.DB().Ping()
	if err != nil {
		return nil, err
	}
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	//db.LogMode(true)

	return &PostgresStorage{&SQLStorage{Db: db.DB(), db: db}}, nil
}

func (self *PostgresStorage) Flush() (err error) {
	cfg := config.CgrConfig()
	for _, scriptName := range []string{CREATE_CDRS_TABLES_SQL, CREATE_TARIFFPLAN_TABLES_SQL} {
		if err := self.CreateTablesFromScript(path.Join(cfg.DataFolderPath, "storage", utils.POSTGRES, scriptName)); err != nil {
			return err
		}
	}
	for _, tbl := range []string{utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA} {
		if _, err := self.Db.Query(fmt.Sprintf("SELECT 1 FROM %s", tbl)); err != nil {
			return err
		}
	}
	return nil
}
