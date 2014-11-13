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
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"path"
	"strconv"
	"time"

	_ "github.com/bmizerany/pq"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
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

func (self *PostgresStorage) SetTPTiming(tm *utils.ApierTPTiming) error {
	if tm == nil {
		return nil //Nothing to set
	}
	tx := self.db.Begin()
	if err := tx.Save(&TpTiming{Tpid: tm.TPid, Tag: tm.TimingId, Years: tm.Years, Months: tm.Months, MonthDays: tm.MonthDays, WeekDays: tm.WeekDays, Time: tm.Time, CreatedAt: time.Now()}).Error; err != nil {
		tx.Rollback()
		tx = self.db.Begin()
		updated := tx.Model(TpTiming{}).Where(&TpTiming{Tpid: tm.TPid, Tag: tm.TimingId}).Updates(&TpTiming{Years: tm.Years, Months: tm.Months, MonthDays: tm.MonthDays, WeekDays: tm.WeekDays, Time: tm.Time})
		if updated.Error != nil {
			tx.Rollback()
			return updated.Error
		}
	}
	tx.Commit()
	return nil
}

func (self *PostgresStorage) LogCallCost(cgrid, source, runid string, cc *CallCost) (err error) {
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

func (self *PostgresStorage) SetRatedCdr(cdr *utils.StoredCdr, extraInfo string) (err error) {
	tx := self.db.Begin()
	saved := tx.Save(&TblRatedCdr{
		Cgrid:       cdr.CgrId,
		Runid:       cdr.MediationRunId,
		Reqtype:     cdr.ReqType,
		Direction:   cdr.Direction,
		Tenant:      cdr.Tenant,
		Category:    cdr.Category,
		Account:     cdr.Account,
		Subject:     cdr.Subject,
		Destination: cdr.Destination,
		SetupTime:   cdr.SetupTime,
		AnswerTime:  cdr.AnswerTime,
		Usage:       cdr.Usage.Seconds(),
		Cost:        cdr.Cost,
		ExtraInfo:   extraInfo,
		CreatedAt:   time.Now(),
	})
	if saved.Error != nil {
		tx.Rollback()
		tx = self.db.Begin()
		updated := tx.Model(TblRatedCdr{}).Where(&TblRatedCdr{Cgrid: cdr.CgrId, Runid: cdr.MediationRunId}).Updates(&TblRatedCdr{Reqtype: cdr.ReqType,
			Direction: cdr.Direction, Tenant: cdr.Tenant, Category: cdr.Category, Account: cdr.Account, Subject: cdr.Subject, Destination: cdr.Destination,
			SetupTime: cdr.SetupTime, AnswerTime: cdr.AnswerTime, Usage: cdr.Usage.Seconds(), Cost: cdr.Cost, ExtraInfo: extraInfo, UpdatedAt: time.Now()})
		if updated.Error != nil {
			tx.Rollback()
			return updated.Error
		}
	}
	tx.Commit()
	return nil

}

func (self *PostgresStorage) GetStoredCdrs(cgrIds, runIds, tors, cdrHosts, cdrSources, reqTypes, directions, tenants, categories, accounts, subjects, destPrefixes, ratedAccounts, ratedSubjects []string,
	orderIdStart, orderIdEnd int64, timeStart, timeEnd time.Time, ignoreErr, ignoreRated, ignoreDerived bool, pagination *utils.Paginator) ([]*utils.StoredCdr, error) {
	var cdrs []*utils.StoredCdr
	var q *bytes.Buffer // Need to query differently since in case of primary, unmediated CDRs some values will be missing
	if ignoreDerived {
		q = bytes.NewBufferString(fmt.Sprintf("SELECT %s.cgrid,%s.id,%s.tor,%s.accid,%s.cdrhost,%s.cdrsource,%s.reqtype,%s.direction,%s.tenant,%s.category,%s.account,%s.subject,%s.destination,%s.setup_time,%s.answer_time,%s.usage,%s.extra_fields,%s.runid,%s.account,%s.subject,%s.cost FROM %s LEFT JOIN %s ON %s.cgrid=%s.cgrid LEFT JOIN %s ON %s.cgrid=%s.cgrid LEFT JOIN %s ON %s.cgrid=%s.cgrid AND %s.runid=%s.runid",
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_RATED_CDRS,
			utils.TBL_COST_DETAILS,
			utils.TBL_COST_DETAILS,
			utils.TBL_RATED_CDRS,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_RATED_CDRS,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_RATED_CDRS,
			utils.TBL_COST_DETAILS,
			utils.TBL_RATED_CDRS,
			utils.TBL_COST_DETAILS,
			utils.TBL_RATED_CDRS,
			utils.TBL_COST_DETAILS))
	} else {
		q = bytes.NewBufferString(fmt.Sprintf("SELECT %s.cgrid,%s.id,%s.tor,%s.accid,%s.cdrhost,%s.cdrsource,%s.reqtype,%s.direction,%s.tenant,%s.category,%s.account,%s.subject,%s.destination,%s.setup_time,%s.answer_time,%s.usage,%s.extra_fields,%s.runid,%s.account,%s.subject,%s.cost FROM %s LEFT JOIN %s ON %s.cgrid=%s.cgrid LEFT JOIN %s ON %s.cgrid=%s.cgrid LEFT JOIN %s ON %s.cgrid=%s.cgrid AND %s.runid=%s.runid",
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_RATED_CDRS,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_RATED_CDRS,
			utils.TBL_COST_DETAILS,
			utils.TBL_COST_DETAILS,
			utils.TBL_RATED_CDRS,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_CDRS_EXTRA,
			utils.TBL_RATED_CDRS,
			utils.TBL_CDRS_PRIMARY,
			utils.TBL_RATED_CDRS,
			utils.TBL_COST_DETAILS,
			utils.TBL_RATED_CDRS,
			utils.TBL_COST_DETAILS,
			utils.TBL_RATED_CDRS,
			utils.TBL_COST_DETAILS))
	}
	q.WriteString(" WHERE")
	for idx, tblName := range []string{utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA, utils.TBL_COST_DETAILS, utils.TBL_RATED_CDRS} {
		if idx != 0 {
			q.WriteString(" AND")
		}
		q.WriteString(fmt.Sprintf(" (%s.deleted_at IS NULL OR %s.deleted_at <= '0001-01-02')", tblName, tblName))
	}
	fltr := new(bytes.Buffer)
	if len(cgrIds) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idxId, cgrId := range cgrIds {
			if idxId != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.cgrid='%s'", utils.TBL_CDRS_PRIMARY, cgrId))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(runIds) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, runId := range runIds {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.runid='%s'", utils.TBL_RATED_CDRS, runId))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(tors) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, host := range tors {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.tor='%s'", utils.TBL_CDRS_PRIMARY, host))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(cdrHosts) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, host := range cdrHosts {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.cdrhost='%s'", utils.TBL_CDRS_PRIMARY, host))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(cdrSources) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, src := range cdrSources {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.cdrsource='%s'", utils.TBL_CDRS_PRIMARY, src))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(reqTypes) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, reqType := range reqTypes {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.reqtype='%s'", utils.TBL_CDRS_PRIMARY, reqType))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(directions) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, direction := range directions {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.direction='%s'", utils.TBL_CDRS_PRIMARY, direction))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(tenants) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, tenant := range tenants {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf("  %s.tenant='%s'", utils.TBL_CDRS_PRIMARY, tenant))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(categories) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, category := range categories {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.category='%s'", utils.TBL_CDRS_PRIMARY, category))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(accounts) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, account := range accounts {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.account='%s'", utils.TBL_CDRS_PRIMARY, account))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(subjects) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, subject := range subjects {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.subject='%s'", utils.TBL_CDRS_PRIMARY, subject))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(destPrefixes) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, destPrefix := range destPrefixes {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.destination LIKE '%s%%'", utils.TBL_CDRS_PRIMARY, destPrefix))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(ratedAccounts) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, ratedAccount := range ratedAccounts {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.account='%s'", utils.TBL_COST_DETAILS, ratedAccount))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if len(ratedSubjects) != 0 {
		qIds := bytes.NewBufferString(" (")
		for idx, ratedSubject := range ratedSubjects {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" %s.subject='%s'", utils.TBL_COST_DETAILS, ratedSubject))
		}
		qIds.WriteString(" )")
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.Write(qIds.Bytes())
	}
	if orderIdStart != 0 {
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.WriteString(fmt.Sprintf(" %s.id>=%d", utils.TBL_CDRS_PRIMARY, orderIdStart))
	}
	if orderIdEnd != 0 {
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.WriteString(fmt.Sprintf(" %s.id<%d", utils.TBL_CDRS_PRIMARY, orderIdEnd))
	}
	if !timeStart.IsZero() {
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.WriteString(fmt.Sprintf(" %s.answer_time>='%s'", utils.TBL_CDRS_PRIMARY, timeStart.Format(time.RFC3339Nano)))
	}
	if !timeEnd.IsZero() {
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.WriteString(fmt.Sprintf(" %s.answer_time<'%s'", utils.TBL_CDRS_PRIMARY, timeEnd.Format(time.RFC3339Nano)))
	}
	if ignoreRated {
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		if ignoreErr {
			fltr.WriteString(fmt.Sprintf(" %s.cost IS NULL", utils.TBL_RATED_CDRS))
		} else {
			fltr.WriteString(fmt.Sprintf(" (%s.cost=-1 OR %s.cost IS NULL)", utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS))
		}
	} else if ignoreErr {
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.WriteString(fmt.Sprintf(" (%s.cost<>-1 OR %s.cost IS NULL)", utils.TBL_RATED_CDRS, utils.TBL_RATED_CDRS))
	}
	if ignoreDerived {
		if fltr.Len() != 0 {
			fltr.WriteString(" AND")
		}
		fltr.WriteString(fmt.Sprintf(" (%s.runid='%s' OR %s.cost IS NULL)", utils.TBL_RATED_CDRS, utils.DEFAULT_RUNID, utils.TBL_RATED_CDRS))
	}
	if fltr.Len() != 0 {
		q.WriteString(fmt.Sprintf(" AND %s", fltr.String()))
	}
	if pagination != nil {
		limLow, limHigh := pagination.GetLimits()
		q.WriteString(fmt.Sprintf(" LIMIT %d,%d", limLow, limHigh))
	}
	rows, err := self.Db.Query(q.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cgrid, tor, accid, cdrhost, cdrsrc, reqtype, direction, tenant, category, account, subject, destination, runid, ratedAccount, ratedSubject sql.NullString
		var extraFields []byte
		var setupTime, answerTime mysql.NullTime
		var orderid int64
		var usage, cost sql.NullFloat64
		var extraFieldsMp map[string]string
		if err := rows.Scan(&cgrid, &orderid, &tor, &accid, &cdrhost, &cdrsrc, &reqtype, &direction, &tenant, &category, &account, &subject, &destination, &setupTime, &answerTime, &usage,
			&extraFields, &runid, &ratedAccount, &ratedSubject, &cost); err != nil {
			return nil, err
		}
		if len(extraFields) != 0 {
			if err := json.Unmarshal(extraFields, &extraFieldsMp); err != nil {
				return nil, fmt.Errorf("JSON unmarshal error for cgrid: %s, runid: %v, error: %s", cgrid.String, runid.String, err.Error())
			}
		}
		usageDur, _ := time.ParseDuration(strconv.FormatFloat(usage.Float64, 'f', -1, 64) + "s")
		storCdr := &utils.StoredCdr{
			CgrId: cgrid.String, OrderId: orderid, TOR: tor.String, AccId: accid.String, CdrHost: cdrhost.String, CdrSource: cdrsrc.String, ReqType: reqtype.String,
			Direction: direction.String, Tenant: tenant.String,
			Category: category.String, Account: account.String, Subject: subject.String, Destination: destination.String,
			SetupTime: setupTime.Time, AnswerTime: answerTime.Time, Usage: usageDur,
			ExtraFields: extraFieldsMp, MediationRunId: runid.String, RatedAccount: ratedAccount.String, RatedSubject: ratedSubject.String, Cost: cost.Float64,
		}
		if !cost.Valid { //There was no cost provided, will fakely insert 0 if we do not handle it and reflect on re-rating
			storCdr.Cost = -1
		}
		cdrs = append(cdrs, storCdr)
	}
	return cdrs, nil
}
