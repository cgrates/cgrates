/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/gorm"
)

func NewMySQLStorage(host, port, name, user, password string, maxConn, maxIdleConn int) (Storage, error) {
	connectString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true", user, password, host, port, name)
	db, err := gorm.Open("mysql", connectString)
	if err != nil {
		return nil, err
	}
	if err = db.DB().Ping(); err != nil {
		return nil, err
	}
	db.DB().SetMaxIdleConns(maxIdleConn)
	db.DB().SetMaxOpenConns(maxConn)
	//db.LogMode(true)

	return &MySQLStorage{&SQLStorage{Db: db.DB(), db: db}}, nil
}

type MySQLStorage struct {
	*SQLStorage
}

func (self *MySQLStorage) Flush(scriptsPath string) (err error) {
	for _, scriptName := range []string{CREATE_CDRS_TABLES_SQL, CREATE_TARIFFPLAN_TABLES_SQL} {
		if err := self.CreateTablesFromScript(path.Join(scriptsPath, scriptName)); err != nil {
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

func (self *MySQLStorage) LogCallCost(cgrid, source, runid string, cc *CallCost) (err error) {
	if cc == nil {
		return nil
	}
	tss, err := json.Marshal(cc.Timespans)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error marshalling timespans to json: %v", err))
		return err
	}
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO %s (cgrid,runid,tor,direction,tenant,category,account,subject,destination,cost,timespans,cost_source,created_at) VALUES ('%s','%s','%s','%s','%s','%s','%s','%s','%s',%f,'%s','%s','%s') ON DUPLICATE KEY UPDATE tor=values(tor),direction=values(direction),tenant=values(tenant),category=values(category),account=values(account),subject=values(subject),destination=values(destination),cost=values(cost),timespans=values(timespans),cost_source=values(cost_source),updated_at='%s'",
		utils.TBL_COST_DETAILS,
		cgrid,
		runid,
		cc.TOR,
		cc.Direction,
		cc.Tenant,
		cc.Category,
		cc.Account,
		cc.Subject,
		cc.Destination,
		cc.Cost,
		tss,
		source,
		time.Now().Format(time.RFC3339),
		time.Now().Format(time.RFC3339)))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute insert statement: %v", err))
		return err
	}
	return nil
}

func (self *MySQLStorage) SetRatedCdr(storedCdr *StoredCdr) (err error) {
	_, err = self.Db.Exec(fmt.Sprintf("INSERT INTO %s (cgrid,runid,reqtype,direction,tenant,category,account,subject,destination,setup_time,answer_time,`usage`,pdd,supplier,disconnect_cause,cost,extra_info,created_at) VALUES ('%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s',%v,%v,'%s','%s',%f,'%s','%s') ON DUPLICATE KEY UPDATE reqtype=values(reqtype),direction=values(direction),tenant=values(tenant),category=values(category),account=values(account),subject=values(subject),destination=values(destination),setup_time=values(setup_time),answer_time=values(answer_time),`usage`=values(`usage`),pdd=values(pdd),cost=values(cost),supplier=values(supplier),disconnect_cause=values(disconnect_cause),extra_info=values(extra_info), updated_at='%s'",
		utils.TBL_RATED_CDRS,
		storedCdr.CgrId,
		storedCdr.MediationRunId,
		storedCdr.ReqType,
		storedCdr.Direction,
		storedCdr.Tenant,
		storedCdr.Category,
		storedCdr.Account,
		storedCdr.Subject,
		storedCdr.Destination,
		storedCdr.SetupTime,
		storedCdr.AnswerTime,
		storedCdr.Usage.Seconds(),
		storedCdr.Pdd.Seconds(),
		storedCdr.Supplier,
		storedCdr.DisconnectCause,
		storedCdr.Cost,
		storedCdr.ExtraInfo,
		time.Now().Format(time.RFC3339),
		time.Now().Format(time.RFC3339)))
	if err != nil {
		Logger.Err(fmt.Sprintf("failed to execute cdr insert statement: %s", err.Error()))
	}
	return
}
