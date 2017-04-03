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
package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cgrates/cgrates/utils"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

func NewPostgresStorage(host, port, name, user, password string, maxConn, maxIdleConn int) (*PostgresStorage, error) {
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

type PostgresStorage struct {
	*SQLStorage
}

func (self *PostgresStorage) SetVersions(vrs Versions, overwrite bool) (err error) {
	tx := self.db.Begin()
	if overwrite {
		tx.Table(utils.TBLVersions).Delete(nil)
	}
	for key, val := range vrs {
		vrModel := &TBLVersion{Item: key, Version: val}
		if !overwrite {
			if err = tx.Model(&TBLVersion{}).Where(
				TBLVersion{Item: vrModel.Item}).Delete(TBLVersion{Version: val}).Error; err != nil {
				tx.Rollback()
				return
			}
		}
		if err = tx.Save(vrModel).Error; err != nil {
			tx.Rollback()
			return
		}
	}
	tx.Commit()
	return
}

// Todo: Make it a template method using interfaces so as not to repeat code
func (self *PostgresStorage) GetCDRs(qryFltr *utils.CDRsFilter, remove bool) ([]*CDR, int64, error) {
	var cdrs []*CDR
	q := self.db.Table(utils.TBLCDRs).Select("*")
	if qryFltr.Unscoped {
		q = q.Unscoped()
	}
	// Add filters, use in to replace the high number of ORs
	if len(qryFltr.CGRIDs) != 0 {
		q = q.Where("cgrid in (?)", qryFltr.CGRIDs)
	}
	if len(qryFltr.NotCGRIDs) != 0 {
		q = q.Where("cgrid not in (?)", qryFltr.NotCGRIDs)
	}
	if len(qryFltr.RunIDs) != 0 {
		q = q.Where("run_id in (?)", qryFltr.RunIDs)
	}
	if len(qryFltr.NotRunIDs) != 0 {
		q = q.Where("run_id not in (?)", qryFltr.NotRunIDs)
	}
	if len(qryFltr.ToRs) != 0 {
		q = q.Where("tor in (?)", qryFltr.ToRs)
	}
	if len(qryFltr.NotToRs) != 0 {
		q = q.Where("tor not in (?)", qryFltr.NotToRs)
	}
	if len(qryFltr.OriginHosts) != 0 {
		q = q.Where("origin_host in (?)", qryFltr.OriginHosts)
	}
	if len(qryFltr.NotOriginHosts) != 0 {
		q = q.Where("origin_host not in (?)", qryFltr.NotOriginHosts)
	}
	if len(qryFltr.Sources) != 0 {
		q = q.Where("source in (?)", qryFltr.Sources)
	}
	if len(qryFltr.NotSources) != 0 {
		q = q.Where("source not in (?)", qryFltr.NotSources)
	}
	if len(qryFltr.RequestTypes) != 0 {
		q = q.Where("request_type in (?)", qryFltr.RequestTypes)
	}
	if len(qryFltr.NotRequestTypes) != 0 {
		q = q.Where("request_type not in (?)", qryFltr.NotRequestTypes)
	}
	if len(qryFltr.Directions) != 0 {
		q = q.Where("direction in (?)", qryFltr.Directions)
	}
	if len(qryFltr.NotDirections) != 0 {
		q = q.Where("direction not in (?)", qryFltr.NotDirections)
	}
	if len(qryFltr.Tenants) != 0 {
		q = q.Where("tenant in (?)", qryFltr.Tenants)
	}
	if len(qryFltr.NotTenants) != 0 {
		q = q.Where("tenant not in (?)", qryFltr.NotTenants)
	}
	if len(qryFltr.Categories) != 0 {
		q = q.Where("category in (?)", qryFltr.Categories)
	}
	if len(qryFltr.NotCategories) != 0 {
		q = q.Where("category not in (?)", qryFltr.NotCategories)
	}
	if len(qryFltr.Accounts) != 0 {
		q = q.Where("account in (?)", qryFltr.Accounts)
	}
	if len(qryFltr.NotAccounts) != 0 {
		q = q.Where("account not in (?)", qryFltr.NotAccounts)
	}
	if len(qryFltr.Subjects) != 0 {
		q = q.Where("subject in (?)", qryFltr.Subjects)
	}
	if len(qryFltr.NotSubjects) != 0 {
		q = q.Where("subject not in (?)", qryFltr.NotSubjects)
	}
	if len(qryFltr.DestinationPrefixes) != 0 { // A bit ugly but still more readable than scopes provided by gorm
		qIds := bytes.NewBufferString("(")
		for idx, destPrefix := range qryFltr.DestinationPrefixes {
			if idx != 0 {
				qIds.WriteString(" OR")
			}
			qIds.WriteString(fmt.Sprintf(" destination LIKE '%s%%'", destPrefix))
		}
		qIds.WriteString(" )")
		q = q.Where(qIds.String())
	}
	if len(qryFltr.NotDestinationPrefixes) != 0 { // A bit ugly but still more readable than scopes provided by gorm
		qIds := bytes.NewBufferString("(")
		for idx, destPrefix := range qryFltr.NotDestinationPrefixes {
			if idx != 0 {
				qIds.WriteString(" AND")
			}
			qIds.WriteString(fmt.Sprintf(" destination not LIKE '%s%%'", destPrefix))
		}
		qIds.WriteString(" )")
		q = q.Where(qIds.String())
	}
	if len(qryFltr.Suppliers) != 0 {
		q = q.Where("supplier in (?)", qryFltr.Subjects)
	}
	if len(qryFltr.NotSuppliers) != 0 {
		q = q.Where("supplier not in (?)", qryFltr.NotSubjects)
	}
	if len(qryFltr.DisconnectCauses) != 0 {
		q = q.Where("disconnect_cause in (?)", qryFltr.DisconnectCauses)
	}
	if len(qryFltr.NotDisconnectCauses) != 0 {
		q = q.Where("disconnect_cause not in (?)", qryFltr.NotDisconnectCauses)
	}
	if len(qryFltr.Costs) != 0 {
		q = q.Where(utils.TBLCDRs+".cost in (?)", qryFltr.Costs)
	}
	if len(qryFltr.NotCosts) != 0 {
		q = q.Where(utils.TBLCDRs+".cost not in (?)", qryFltr.NotCosts)
	}
	if len(qryFltr.ExtraFields) != 0 { // Extra fields searches, implemented as contains in extra field
		qIds := bytes.NewBufferString("(")
		needOr := false
		for field, value := range qryFltr.ExtraFields {
			if needOr {
				qIds.WriteString(" OR")
			}
			if value == utils.MetaExists {
				qIds.WriteString(fmt.Sprintf(" extra_fields ?'%s'", field))
			} else {
				qIds.WriteString(fmt.Sprintf(" (extra_fields ->> '%s') = '%s'", field, value))
			}
			needOr = true
		}
		qIds.WriteString(" )")
		q = q.Where(qIds.String())
	}
	if len(qryFltr.NotExtraFields) != 0 { // Extra fields searches, implemented as contains in extra field
		qIds := bytes.NewBufferString("(")
		needAnd := false
		for field, value := range qryFltr.NotExtraFields {
			if needAnd {
				qIds.WriteString(" AND")
			}
			if value == utils.MetaExists {
				qIds.WriteString(fmt.Sprintf(" NOT extra_fields ?'%s'", field))
			} else {
				qIds.WriteString(fmt.Sprintf(" NOT (extra_fields ?'%s' AND (extra_fields ->> '%s') = '%s')", field, field, value))
			}
			needAnd = true
		}
		qIds.WriteString(" )")
		q = q.Where(qIds.String())
	}
	if qryFltr.OrderIDStart != nil { // Keep backwards compatible by testing 0 value
		q = q.Where(utils.TBLCDRs+".id >= ?", *qryFltr.OrderIDStart)
	}
	if qryFltr.OrderIDEnd != nil {
		q = q.Where(utils.TBLCDRs+".id < ?", *qryFltr.OrderIDEnd)
	}
	if qryFltr.SetupTimeStart != nil {
		q = q.Where("setup_time >= ?", qryFltr.SetupTimeStart)
	}
	if qryFltr.SetupTimeEnd != nil {
		q = q.Where("setup_time < ?", qryFltr.SetupTimeEnd)
	}
	if qryFltr.AnswerTimeStart != nil && !qryFltr.AnswerTimeStart.IsZero() { // With IsZero we keep backwards compatible with ApierV1
		q = q.Where("answer_time >= ?", qryFltr.AnswerTimeStart)
	}
	if qryFltr.AnswerTimeEnd != nil && !qryFltr.AnswerTimeEnd.IsZero() {
		q = q.Where("answer_time < ?", qryFltr.AnswerTimeEnd)
	}
	if qryFltr.CreatedAtStart != nil && !qryFltr.CreatedAtStart.IsZero() { // With IsZero we keep backwards compatible with ApierV1
		q = q.Where("created_at >= ?", qryFltr.CreatedAtStart)
	}
	if qryFltr.CreatedAtEnd != nil && !qryFltr.CreatedAtEnd.IsZero() {
		q = q.Where("created_at < ?", qryFltr.CreatedAtEnd)
	}
	if qryFltr.UpdatedAtStart != nil && !qryFltr.UpdatedAtStart.IsZero() { // With IsZero we keep backwards compatible with ApierV1
		q = q.Where("updated_at >= ?", qryFltr.UpdatedAtStart)
	}
	if qryFltr.UpdatedAtEnd != nil && !qryFltr.UpdatedAtEnd.IsZero() {
		q = q.Where("updated_at < ?", qryFltr.UpdatedAtEnd)
	}
	if len(qryFltr.MinUsage) != 0 {
		if minUsage, err := utils.ParseDurationWithSecs(qryFltr.MinUsage); err != nil {
			return nil, 0, err
		} else {
			if self.db.Dialect().GetName() == utils.MYSQL { // MySQL needs escaping for usage
				q = q.Where("`usage` >= ?", minUsage.Seconds())
			} else {
				q = q.Where("usage >= ?", minUsage.Seconds())
			}
		}
	}
	if len(qryFltr.MaxUsage) != 0 {
		if maxUsage, err := utils.ParseDurationWithSecs(qryFltr.MaxUsage); err != nil {
			return nil, 0, err
		} else {
			if self.db.Dialect().GetName() == utils.MYSQL { // MySQL needs escaping for usage
				q = q.Where("`usage` < ?", maxUsage.Seconds())
			} else {
				q = q.Where("usage < ?", maxUsage.Seconds())
			}
		}

	}
	if len(qryFltr.MinPDD) != 0 {
		if minPDD, err := utils.ParseDurationWithSecs(qryFltr.MinPDD); err != nil {
			return nil, 0, err
		} else {
			q = q.Where("pdd >= ?", minPDD.Seconds())
		}

	}
	if len(qryFltr.MaxPDD) != 0 {
		if maxPDD, err := utils.ParseDurationWithSecs(qryFltr.MaxPDD); err != nil {
			return nil, 0, err
		} else {
			q = q.Where("pdd < ?", maxPDD.Seconds())
		}
	}

	if qryFltr.MinCost != nil {
		if qryFltr.MaxCost == nil {
			q = q.Where("cost >= ?", *qryFltr.MinCost)
		} else if *qryFltr.MinCost == 0.0 && *qryFltr.MaxCost == -1.0 { // Special case when we want to skip errors
			q = q.Where("( cost IS NULL OR cost >= 0.0 )")
		} else {
			q = q.Where("cost >= ?", *qryFltr.MinCost)
			q = q.Where("cost < ?", *qryFltr.MaxCost)
		}
	} else if qryFltr.MaxCost != nil {
		if *qryFltr.MaxCost == -1.0 { // Non-rated CDRs
			q = q.Where("cost IS NULL") // Need to include it otherwise all CDRs will be returned
		} else { // Above limited CDRs, since MinCost is empty, make sure we query also NULL cost
			q = q.Where(fmt.Sprintf("( cost IS NULL OR cost < %f )", *qryFltr.MaxCost))
		}
	}
	if qryFltr.Paginator.Limit != nil {
		q = q.Limit(*qryFltr.Paginator.Limit)
	}
	if qryFltr.Paginator.Offset != nil {
		q = q.Offset(*qryFltr.Paginator.Offset)
	}
	if remove { // Remove CDRs instead of querying them
		if err := q.Delete(nil).Error; err != nil {
			q.Rollback()
			return nil, 0, err
		}
	}
	if qryFltr.Count { // Count CDRs
		var cnt int64
		if err := q.Count(&cnt).Error; err != nil {
			//if err := q.Debug().Count(&cnt).Error; err != nil {
			return nil, 0, err
		}
		return nil, cnt, nil
	}
	// Execute query
	results := make([]*TBLCDRs, 0)
	if err := q.Find(&results).Error; err != nil {
		return nil, 0, err
	}
	for _, result := range results {
		extraFieldsMp := make(map[string]string)
		if result.ExtraFields != "" {
			if err := json.Unmarshal([]byte(result.ExtraFields), &extraFieldsMp); err != nil {
				return nil, 0, fmt.Errorf("JSON unmarshal error for cgrid: %s, runid: %v, error: %s", result.Cgrid, result.RunID, err.Error())
			}
		}
		var callCost CallCost
		if result.CostDetails != "" {
			if err := json.Unmarshal([]byte(result.CostDetails), &callCost); err != nil {
				return nil, 0, fmt.Errorf("JSON unmarshal callcost error for cgrid: %s, runid: %v, error: %s", result.Cgrid, result.RunID, err.Error())
			}
		}
		acntSummary, err := NewAccountSummaryFromJSON(result.AccountSummary)
		if err != nil {
			return nil, 0, fmt.Errorf("JSON unmarshal account summary error for cgrid: %s, runid: %v, error: %s", result.Cgrid, result.RunID, err.Error())
		}
		usageDur := time.Duration(result.Usage * utils.NANO_MULTIPLIER)
		pddDur := time.Duration(result.Pdd * utils.NANO_MULTIPLIER)
		storCdr := &CDR{
			CGRID:           result.Cgrid,
			RunID:           result.RunID,
			OrderID:         result.ID,
			OriginHost:      result.OriginHost,
			Source:          result.Source,
			OriginID:        result.OriginID,
			ToR:             result.Tor,
			RequestType:     result.RequestType,
			Direction:       result.Direction,
			Tenant:          result.Tenant,
			Category:        result.Category,
			Account:         result.Account,
			Subject:         result.Subject,
			Destination:     result.Destination,
			SetupTime:       result.SetupTime,
			PDD:             pddDur,
			AnswerTime:      result.AnswerTime,
			Usage:           usageDur,
			Supplier:        result.Supplier,
			DisconnectCause: result.DisconnectCause,
			ExtraFields:     extraFieldsMp,
			CostSource:      result.CostSource,
			Cost:            result.Cost,
			CostDetails:     &callCost,
			AccountSummary:  acntSummary,
			ExtraInfo:       result.ExtraInfo,
		}
		cdrs = append(cdrs, storCdr)
	}
	if len(cdrs) == 0 && !remove {
		return cdrs, 0, utils.ErrNotFound
	}
	return cdrs, 0, nil
}
