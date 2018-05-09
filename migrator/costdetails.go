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

/*
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCostDetails() (err error) {
	if m.storDBOut == nil {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.NoStorDBConnection,
			"no connection to StorDB")
	}
	vrs, err := m.storDBOut.StorDB().GetVersions(utils.COST_DETAILS)
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying storDB for versions", err.Error()))
	} else if len(vrs) == 0 {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UndefinedVersion,
			"version number is not defined for CostDetails model")
	}
	if vrs[utils.COST_DETAILS] != 1 { // Right now we only support migrating from version 1
		log.Print("Wrong version")
		return
	}
	var storSQL *sql.DB
	switch m.storDBType {
	case utils.MYSQL:
		storSQL = m.storDBOut.(*engine.SQLStorage).Db
	case utils.POSTGRES:
		storSQL = m.storDBOut.(*engine.SQLStorage).Db
	default:
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UnsupportedDB,
			fmt.Sprintf("unsupported database type: <%s>", m.storDBType))
	}
	rows, err := storSQL.Query("SELECT id, tor, direction, tenant, category, account, subject, destination, cost, cost_details FROM cdrs")
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying storDB for cdrs", err.Error()))
	}

	defer rows.Close()

	for cnt := 0; rows.Next(); cnt++ {
		var id int64
		var ccDirection, ccCategory, ccTenant, ccSubject, ccAccount, ccDestination, ccTor sql.NullString
		var ccCost sql.NullFloat64
		var tts []byte

		if err := rows.Scan(&id, &ccTor, &ccDirection, &ccTenant, &ccCategory, &ccAccount, &ccSubject, &ccDestination, &ccCost, &tts); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when scanning at count: <%d>", err.Error(), cnt))
		}
		var v1tmsps v1TimeSpans
		if err := json.Unmarshal(tts, &v1tmsps); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<Migrator> Unmarshalling timespans at CDR with id: <%d>, error: <%s>", id, err.Error()))
			continue
		}
		v1CC := &v1CallCost{Direction: ccDirection.String, Category: ccCategory.String, Tenant: ccTenant.String,
			Subject: ccSubject.String, Account: ccAccount.String, Destination: ccDestination.String, TOR: ccTor.String,
			Cost: ccCost.Float64, Timespans: v1tmsps}

		cc := v1CC.AsCallCost()
		if cc == nil {
			utils.Logger.Warning(
				fmt.Sprintf("<Migrator> Error: <%s> when converting into CallCost CDR with id: <%d>", err.Error(), id))
			continue
		}
		if m.dryRun != true {
			if _, err := storSQL.Exec(fmt.Sprintf("UPDATE cdrs SET cost_details='%s' WHERE id=%d", cc.AsJSON(), id)); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<Migrator> Error: <%s> updating CDR with id <%d> into StorDB", err.Error(), id))
				continue
			}

			m.stats[utils.COST_DETAILS] += 1
			// All done, update version wtih current one
			vrs = engine.Versions{utils.COST_DETAILS: engine.CurrentStorDBVersions()[utils.COST_DETAILS]}
			if err := m.storDBOut.SetVersions(vrs, false); err != nil {
				return utils.NewCGRError(utils.Migrator,
					utils.ServerErrorCaps,
					err.Error(),
					fmt.Sprintf("error: <%s> when updating CostDetails version into StorDB", err.Error()))
			}
		}
	}
	return
}

type v1CallCost struct {
	Direction, Category, Tenant, Subject, Account, Destination, TOR string
	Cost                                                            float64
	Timespans                                                       v1TimeSpans
}

type v1TimeSpans []*v1TimeSpan

type v1TimeSpan struct {
	TimeStart, TimeEnd                                         time.Time
	Cost                                                       float64
	RateInterval                                               *engine.RateInterval
	DurationIndex                                              time.Duration
	Increments                                                 v1Increments
	MatchedSubject, MatchedPrefix, MatchedDestId, RatingPlanId string
}

type v1Increments []*v1Increment

type v1Increment struct {
	Duration            time.Duration
	Cost                float64
	BalanceRateInterval *engine.RateInterval
	BalanceInfo         *v1BalanceInfo
	UnitInfo            *v1UnitInfo
	CompressFactor      int
}

type v1BalanceInfo struct {
	UnitBalanceUuid  string
	MoneyBalanceUuid string
	AccountId        string // used when debited from shared balance
}

type v1UnitInfo struct {
	DestinationId string
	Quantity      float64
	TOR           string
}

func (v1cc *v1CallCost) AsCallCost() (cc *engine.CallCost) {
	cc = new(engine.CallCost)
	cc.Direction = v1cc.Direction
	cc.Category = v1cc.Category
	cc.Tenant = v1cc.Tenant
	cc.Account = v1cc.Account
	cc.Subject = v1cc.Subject
	cc.Destination = v1cc.Destination
	cc.TOR = v1cc.TOR
	cc.Cost = v1cc.Cost
	cc.Timespans = make(engine.TimeSpans, len(v1cc.Timespans))
	for i, v1ts := range v1cc.Timespans {
		cc.Timespans[i] = &engine.TimeSpan{TimeStart: v1ts.TimeStart,
			TimeEnd:        v1ts.TimeEnd,
			Cost:           v1ts.Cost,
			RateInterval:   v1ts.RateInterval,
			DurationIndex:  v1ts.DurationIndex,
			Increments:     make(engine.Increments, len(v1ts.Increments)),
			MatchedSubject: v1ts.MatchedSubject,
			MatchedPrefix:  v1ts.MatchedPrefix,
			MatchedDestId:  v1ts.MatchedDestId,
			RatingPlanId:   v1ts.RatingPlanId,
		}
		for j, v1Incrm := range v1ts.Increments {
			cc.Timespans[i].Increments[j] = &engine.Increment{
				Duration:       v1Incrm.Duration,
				Cost:           v1Incrm.Cost,
				CompressFactor: v1Incrm.CompressFactor,
				BalanceInfo: &engine.DebitInfo{
					AccountID: v1Incrm.BalanceInfo.AccountId,
				},
			}
			if v1Incrm.BalanceInfo.UnitBalanceUuid != "" {
				cc.Timespans[i].Increments[j].BalanceInfo.Unit = &engine.UnitInfo{
					UUID:          v1Incrm.BalanceInfo.UnitBalanceUuid,
					Value:         v1Incrm.UnitInfo.Quantity,
					DestinationID: v1Incrm.UnitInfo.DestinationId,
					TOR:           v1Incrm.UnitInfo.TOR,
				}
			} else if v1Incrm.BalanceInfo.MoneyBalanceUuid != "" {
				cc.Timespans[i].Increments[j].BalanceInfo.Monetary = &engine.MonetaryInfo{
					UUID: v1Incrm.BalanceInfo.MoneyBalanceUuid,
					//Value: v1Incrm.UnitInfo.Quantity,
				}
			}
		}
	}
	return
}
*/
