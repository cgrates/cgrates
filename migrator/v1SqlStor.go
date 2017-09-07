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

// import (
// 	"fmt"
// 	"time"

// 	_ "github.com/go-sql-driver/mysql"
// 	"github.com/cgrates/cgrates/utils"
// 	"github.com/jinzhu/gorm"
// 	_ "github.com/lib/pq"
// )

// type v1SQLStorage struct {
// 	Db *sql.DB
// 	db *gorm.DB
// 	Storsql
// }

// func NewPostgresStorage(host, port, name, user, password string, maxConn, maxIdleConn, connMaxLifetime int) (*v1SQLStorage, error) {
// 	connectString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", host, port, name, user, password)
// 	db, err := gorm.Open("postgres", connectString)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = db.DB().Ping()
// 	if err != nil {
// 		return nil, err
// 	}
// 	db.DB().SetMaxIdleConns(maxIdleConn)
// 	db.DB().SetMaxOpenConns(maxConn)
// 	db.DB().SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
// 	//db.LogMode(true)
// 	postgressStorage := new(PostgresStorage)
// 	postgressStorage.db = db
// 	postgressStorage.Db = db.DB()
// 	return &SQLStorage{db.DB(), db, postgressStorage, postgressStorage}, nil
// }

// func NewMySQLStorage(host, port, name, user, password string, maxConn, maxIdleConn, connMaxLifetime int) (*v1SQLStorage, error) {
// 	connectString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true", user, password, host, port, name)
// 	db, err := gorm.Open("mysql", connectString)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err = db.DB().Ping(); err != nil {
// 		return nil, err
// 	}
// 	db.DB().SetMaxIdleConns(maxIdleConn)
// 	db.DB().SetMaxOpenConns(maxConn)
// 	db.DB().SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
// 	//db.LogMode(true)
// 	mySQLStorage := new(MySQLStorage)
// 	mySQLStorage.db = db
// 	mySQLStorage.Db = db.DB()
// 	return &SQLStorage{db.DB(), db, mySQLStorage, mySQLStorage}, nil
// }

// getV1CallCost() (v1CC *v1CallCost, err error){
// 	//echivalentu la ce am facut la mongo doar ca cu rows

// 	var storSQL *sql.DB
// 	switch m.storDBType {
// 	case utils.MYSQL:
// 		storSQL = m.storDB.(*engine.SQLStorage).Db
// 	case utils.POSTGRES:
// 		storSQL = m.storDB.(*engine.PostgresStorage).Db
// 	default:
// 		return utils.NewCGRError(utils.Migrator,
// 			utils.MandatoryIEMissingCaps,
// 			utils.UnsupportedDB,
// 			fmt.Sprintf("unsupported database type: <%s>", m.storDBType))
// 	}
// 	rows, err := storSQL.Query("SELECT id, tor, direction, tenant, category, account, subject, destination, cost, cost_details FROM cdrs")
// 	if err != nil {
// 		return utils.NewCGRError(utils.Migrator,
// 			utils.ServerErrorCaps,
// 			err.Error(),
// 			fmt.Sprintf("error: <%s> when querying storDB for cdrs", err.Error()))
// 	}

// 	defer rows.Close()

// 	for cnt := 0; rows.Next(); cnt++ {
// 		var id int64
// 		var ccDirection, ccCategory, ccTenant, ccSubject, ccAccount, ccDestination, ccTor sql.NullString
// 		var ccCost sql.NullFloat64
// 		var tts []byte

// 		if err := rows.Scan(&id, &ccTor, &ccDirection, &ccTenant, &ccCategory, &ccAccount, &ccSubject, &ccDestination, &ccCost, &tts); err != nil {
// 			return utils.NewCGRError(utils.Migrator,
// 				utils.ServerErrorCaps,
// 				err.Error(),
// 				fmt.Sprintf("error: <%s> when scanning at count: <%d>", err.Error(), cnt))
// 		}
// 		var v1tmsps v1TimeSpans
// 		if err := json.Unmarshal(tts, &v1tmsps); err != nil {
// 			utils.Logger.Warning(
// 				fmt.Sprintf("<Migrator> Unmarshalling timespans at CDR with id: <%d>, error: <%s>", id, err.Error()))
// 			continue
// 		}
// 	return	v1CC := &v1CallCost{Direction: ccDirection.String, Category: ccCategory.String, Tenant: ccTenant.String,
// 			Subject: ccSubject.String, Account: ccAccount.String, Destination: ccDestination.String, TOR: ccTor.String,
// 			Cost: ccCost.Float64, Timespans: v1tmsps}

// 	}
// }
