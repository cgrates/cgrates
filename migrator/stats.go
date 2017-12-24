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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type v1Stat struct {
	Id              string        // Config id, unique per config instance
	QueueLength     int           // Number of items in the stats buffer
	TimeWindow      time.Duration // Will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
	SaveInterval    time.Duration
	Metrics         []string        // ASR, ACD, ACC
	SetupInterval   []time.Time     // CDRFieldFilter on SetupInterval, 2 or less items (>= start interval,< stop_interval)
	TOR             []string        // CDRFieldFilter on TORs
	CdrHost         []string        // CDRFieldFilter on CdrHosts
	CdrSource       []string        // CDRFieldFilter on CdrSources
	ReqType         []string        // CDRFieldFilter on RequestTypes
	Direction       []string        // CDRFieldFilter on Directions
	Tenant          []string        // CDRFieldFilter on Tenants
	Category        []string        // CDRFieldFilter on Categories
	Account         []string        // CDRFieldFilter on Accounts
	Subject         []string        // CDRFieldFilter on Subjects
	DestinationIds  []string        // CDRFieldFilter on DestinationPrefixes
	UsageInterval   []time.Duration // CDRFieldFilter on UsageInterval, 2 or less items (>= Usage, <Usage)
	PddInterval     []time.Duration // CDRFieldFilter on PddInterval, 2 or less items (>= Pdd, <Pdd)
	Supplier        []string        // CDRFieldFilter on Suppliers
	DisconnectCause []string        // Filter on DisconnectCause
	MediationRunIds []string        // CDRFieldFilter on MediationRunIds
	RatedAccount    []string        // CDRFieldFilter on RatedAccounts
	RatedSubject    []string        // CDRFieldFilter on RatedSubjects
	CostInterval    []float64       // CDRFieldFilter on CostInterval, 2 or less items, (>=Cost, <Cost)
	Triggers        engine.ActionTriggers
}

type v1Stats []*v1Stat

func (m *Migrator) migrateCurrentStats() (err error) {
	var ids []string
	tenant := config.CgrConfig().DefaultTenant
	//StatQueue
	ids, err = m.dmIN.DataDB().GetKeysForPrefix(utils.StatQueuePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.StatQueuePrefix+tenant+":")
		sgs, err := m.dmIN.GetStatQueue(tenant, idg, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if sgs != nil {
			if m.dryRun != true {
				if err := m.dmOut.SetStatQueue(sgs); err != nil {
					return err
				}
				m.stats[utils.StatS] += 1
			}
		}
	}
	//StatQueueProfile
	ids, err = m.dmIN.DataDB().GetKeysForPrefix(utils.StatQueueProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.StatQueueProfilePrefix+tenant+":")
		sgs, err := m.dmIN.GetStatQueueProfile(tenant, idg, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		if sgs != nil {
			if m.dryRun != true {
				if err := m.dmOut.SetStatQueueProfile(sgs); err != nil {
					return err
				}
			}
		}
	}

	return
}

func (m *Migrator) migrateV1CDRSTATS() (err error) {
	var v1Sts *v1Stat
	for {
		v1Sts, err = m.oldDataDB.getV1Stats()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if v1Sts.Id != "" {
			if len(v1Sts.Triggers) != 0 {
				for _, Trigger := range v1Sts.Triggers {
					if err := m.SasThreshold(Trigger); err != nil {
						return err

					}
				}
			}
			filter, sq, sts, err := v1Sts.AsStatQP()
			if err != nil {
				return err
			}
			if !m.dryRun {
				if err := m.dmOut.SetFilter(filter); err != nil {
					return err
				}
				if err := m.dmOut.SetStatQueue(sq); err != nil {
					return err
				}
				if err := m.dmOut.SetStatQueueProfile(sts); err != nil {
					return err
				}
				m.stats[utils.StatS] += 1
			}
		}
	}
	if m.dryRun != true {
		// All done, update version wtih current one
		vrs := engine.Versions{utils.StatS: engine.CurrentStorDBVersions()[utils.StatS]}
		if err = m.dmOut.DataDB().SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating Stats version into dataDB", err.Error()))
		}
	}
	return
}

func (m *Migrator) migrateStats() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	vrs, err = m.dmOut.DataDB().GetVersions(utils.TBLVersions)
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
	} else if len(vrs) == 0 {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UndefinedVersion,
			"version number is not defined for ActionTriggers model")
	}
	switch vrs[utils.StatS] {
	case current[utils.StatS]:
		if m.sameDataDB {
			return
		}
		if err := m.migrateCurrentStats(); err != nil {
			return err
		}
		return

	case 1:
		if err := m.migrateV1CDRSTATS(); err != nil {
			return err
		}
	}
	return
}

func (v1Sts v1Stat) AsStatQP() (filter *engine.Filter, sq *engine.StatQueue, stq *engine.StatQueueProfile, err error) {
	var filters []*engine.RequestFilter
	if len(v1Sts.SetupInterval) == 1 {
		x, err := engine.NewRequestFilter(engine.MetaGreaterOrEqual, "SetupInterval", []string{v1Sts.SetupInterval[0].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	} else if len(v1Sts.SetupInterval) == 2 {
		x, err := engine.NewRequestFilter(engine.MetaLessThan, "SetupInterval", []string{v1Sts.SetupInterval[1].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}

	if len(v1Sts.TOR) != 0 {
		x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "TOR", v1Sts.TOR)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.CdrHost) != 0 {
		x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "CdrHost", v1Sts.CdrHost)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.ReqType) != 0 {
		x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "ReqType", v1Sts.ReqType)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Direction) != 0 {
		x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "Direction", v1Sts.Direction)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Category) != 0 {
		x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "Category", v1Sts.Category)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Account) != 0 {
		x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "Account", v1Sts.Account)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Subject) != 0 {
		x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "Subject", v1Sts.Subject)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Supplier) != 0 {
		x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "Supplier", v1Sts.Supplier)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.UsageInterval) == 1 {
		x, err := engine.NewRequestFilter(engine.MetaGreaterOrEqual, "UsageInterval", []string{v1Sts.UsageInterval[0].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	} else if len(v1Sts.UsageInterval) == 2 {
		x, err := engine.NewRequestFilter(engine.MetaLessThan, "UsageInterval", []string{v1Sts.UsageInterval[1].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.PddInterval) == 1 {
		x, err := engine.NewRequestFilter(engine.MetaGreaterOrEqual, "PddInterval", []string{v1Sts.PddInterval[0].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	} else if len(v1Sts.PddInterval) == 2 {
		x, err := engine.NewRequestFilter(engine.MetaLessThan, "PddInterval", []string{v1Sts.PddInterval[1].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Supplier) != 0 {
		x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "Supplier", v1Sts.Supplier)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.DisconnectCause) != 0 {
		x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "DisconnectCause", v1Sts.DisconnectCause)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.MediationRunIds) != 0 {
		x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "MediationRunIds", v1Sts.MediationRunIds)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.RatedSubject) != 0 {
		x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "RatedSubject", v1Sts.RatedSubject)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.CostInterval) == 1 {
		x, err := engine.NewRequestFilter(engine.MetaGreaterOrEqual, "CostInterval", []string{strconv.FormatFloat(v1Sts.CostInterval[0], 'f', 6, 64)})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	} else if len(v1Sts.CostInterval) == 2 {
		x, err := engine.NewRequestFilter(engine.MetaLessThan, "CostInterval", []string{strconv.FormatFloat(v1Sts.CostInterval[1], 'f', 6, 64)})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	filter = &engine.Filter{Tenant: config.CgrConfig().DefaultTenant, ID: v1Sts.Id, RequestFilters: filters}
	stq = &engine.StatQueueProfile{
		ID:          v1Sts.Id,
		QueueLength: v1Sts.QueueLength,
		Metrics:     []*utils.MetricWithParams{},
		Tenant:      config.CgrConfig().DefaultTenant,
		Blocker:     false,
		Stored:      false,
		Thresholds:  []string{},
		FilterIDs:   []string{v1Sts.Id},
	}
	if v1Sts.SaveInterval != 0 {
		stq.Stored = true
	}
	if len(v1Sts.Triggers) != 0 {
		for i, _ := range v1Sts.Triggers {
			stq.Thresholds = append(stq.Thresholds, v1Sts.Triggers[i].ID)
		}
	}
	sq = &engine.StatQueue{Tenant: config.CgrConfig().DefaultTenant,
		ID:        v1Sts.Id,
		SQMetrics: make(map[string]map[string]engine.StatMetric),
	}
	if len(v1Sts.Metrics) != 0 {
		for i, _ := range v1Sts.Metrics {
			if !strings.HasPrefix(v1Sts.Metrics[i], "*") {
				v1Sts.Metrics[i] = "*" + v1Sts.Metrics[i]
			}
			v1Sts.Metrics[i] = strings.ToLower(v1Sts.Metrics[i])

			stq.Metrics = append(stq.Metrics, &utils.MetricWithParams{MetricID: v1Sts.Metrics[i]})
			if metric, err := engine.NewStatMetric(stq.Metrics[i].MetricID, 0, ""); err != nil {
				return nil, nil, nil, err
			} else {
				if _, has := sq.SQMetrics[stq.Metrics[i].MetricID]; !has {
					sq.SQMetrics[stq.Metrics[i].MetricID] = make(map[string]engine.StatMetric)
				}
				sq.SQMetrics[stq.Metrics[i].MetricID][""] = metric
			}
		}
	}
	return filter, sq, stq, nil
}
