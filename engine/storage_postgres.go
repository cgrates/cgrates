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
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	*SQLStorage
}

func NewPostgresStorage(host, port, name, user, password string, maxConn, maxIdleConn int) (Storage, error) {
	connectString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", host, port, name, user, password)
	db, err := gorm.Open("postgres", connectString)
	if err != nil {
		return nil, err
	}
	err = db.DB().Ping()
	if err != nil {
		return nil, err
	}
	db.DB().SetMaxIdleConns(maxIdleConn)
	db.DB().SetMaxOpenConns(maxConn)
	//db.LogMode(true)

	return &PostgresStorage{&SQLStorage{Db: db.DB(), db: db}}, nil
}

func (self *PostgresStorage) Flush(scriptsPath string) (err error) {
	for _, scriptName := range []string{utils.CREATE_CDRS_TABLES_SQL, utils.CREATE_TARIFFPLAN_TABLES_SQL} {
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

func (self *PostgresStorage) LogCallCost(cgrid, source, runid string, cc *CallCost) (err error) {
	if cc == nil {
		return nil
	}
	tss, err := json.Marshal(cc.Timespans)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error marshalling timespans to json: %v", err))
		return err
	}
	tx := self.db.Begin()
	cd := &TblCostDetail{
		Cgrid:       cgrid,
		Runid:       runid,
		Tor:         cc.TOR,
		Direction:   cc.Direction,
		Tenant:      cc.Tenant,
		Category:    cc.Category,
		Account:     cc.Account,
		Subject:     cc.Subject,
		Destination: cc.Destination,
		Cost:        cc.Cost,
		Timespans:   string(tss),
		CostSource:  source,
		CreatedAt:   time.Now(),
	}

	if tx.Save(cd).Error != nil { // Check further since error does not properly reflect duplicates here (sql: no rows in result set)
		tx.Rollback()
		tx = self.db.Begin()
		updated := tx.Model(TblCostDetail{}).Where(&TblCostDetail{Cgrid: cgrid, Runid: runid}).Updates(&TblCostDetail{Tor: cc.TOR, Direction: cc.Direction, Tenant: cc.Tenant, Category: cc.Category,
			Account: cc.Account, Subject: cc.Subject, Destination: cc.Destination, Cost: cc.Cost, Timespans: string(tss), CostSource: source, UpdatedAt: time.Now()})
		if updated.Error != nil {
			tx.Rollback()
			return updated.Error
		}
	}
	tx.Commit()
	return nil
}

func (self *PostgresStorage) SetRatedCdr(cdr *StoredCdr) (err error) {
	tx := self.db.Begin()
	saved := tx.Save(&TblRatedCdr{
		Cgrid:           cdr.CgrId,
		Runid:           cdr.MediationRunId,
		Reqtype:         cdr.ReqType,
		Direction:       cdr.Direction,
		Tenant:          cdr.Tenant,
		Category:        cdr.Category,
		Account:         cdr.Account,
		Subject:         cdr.Subject,
		Destination:     cdr.Destination,
		SetupTime:       cdr.SetupTime,
		AnswerTime:      cdr.AnswerTime,
		Usage:           cdr.Usage.Seconds(),
		Pdd:             cdr.Pdd.Seconds(),
		Supplier:        cdr.Supplier,
		DisconnectCause: cdr.DisconnectCause,
		Cost:            cdr.Cost,
		ExtraInfo:       cdr.ExtraInfo,
		CreatedAt:       time.Now(),
	})
	if saved.Error != nil {
		tx.Rollback()
		tx = self.db.Begin()
		updated := tx.Model(TblRatedCdr{}).Where(&TblRatedCdr{Cgrid: cdr.CgrId, Runid: cdr.MediationRunId}).Updates(&TblRatedCdr{Reqtype: cdr.ReqType,
			Direction: cdr.Direction, Tenant: cdr.Tenant, Category: cdr.Category, Account: cdr.Account, Subject: cdr.Subject, Destination: cdr.Destination,
			SetupTime: cdr.SetupTime, AnswerTime: cdr.AnswerTime, Usage: cdr.Usage.Seconds(), Pdd: cdr.Pdd.Seconds(), Supplier: cdr.Supplier, DisconnectCause: cdr.DisconnectCause,
			Cost: cdr.Cost, ExtraInfo: cdr.ExtraInfo,
			UpdatedAt: time.Now()})
		if updated.Error != nil {
			tx.Rollback()
			return updated.Error
		}
	}
	tx.Commit()
	return nil

}
