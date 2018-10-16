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
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	//StatQueue
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.StatQueuePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.StatQueuePrefix+tenant+":")
		sgs, err := m.dmIN.DataManager().GetStatQueue(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {

			return err
		}
		if sgs != nil {
			if m.dryRun != true {
				if err := m.dmOut.DataManager().SetStatQueue(sgs); err != nil {
					return err
				}
				m.stats[utils.StatS] += 1
			}
		}
	}
	//StatQueueProfile
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.StatQueueProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.StatQueueProfilePrefix+tenant+":")
		sgs, err := m.dmIN.DataManager().GetStatQueueProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if sgs != nil {
			if m.dryRun != true {
				if err := m.dmOut.DataManager().SetStatQueueProfile(sgs, true); err != nil {
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
		v1Sts, err = m.dmIN.getV1Stats()
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
				if err := m.dmOut.DataManager().SetFilter(filter); err != nil {
					return err
				}
				if err := m.dmOut.DataManager().SetStatQueue(sq); err != nil {
					return err
				}
				if err := m.dmOut.DataManager().SetStatQueueProfile(sts, true); err != nil {
					return err
				}
				m.stats[utils.StatS] += 1
			}
		}
	}
	if m.dryRun != true {
		// All done, update version wtih current one
		vrs := engine.Versions{utils.StatS: engine.CurrentDataDBVersions()[utils.StatS]}
		if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
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
	vrs, err = m.dmOut.DataManager().DataDB().GetVersions("")
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
	case 1:
		if err := m.migrateV1CDRSTATS(); err != nil {
			return err
		}
	case current[utils.StatS]:
		if m.sameDataDB {
			return
		}
		if err := m.migrateCurrentStats(); err != nil {
			return err
		}
		return
	}
	return
}

func (v1Sts v1Stat) AsStatQP() (filter *engine.Filter, sq *engine.StatQueue, stq *engine.StatQueueProfile, err error) {
	var filters []*engine.FilterRule
	if len(v1Sts.SetupInterval) == 1 {
		x, err := engine.NewFilterRule(engine.MetaGreaterOrEqual,
			"SetupInterval", []string{v1Sts.SetupInterval[0].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	} else if len(v1Sts.SetupInterval) == 2 {
		x, err := engine.NewFilterRule(engine.MetaLessThan,
			"SetupInterval", []string{v1Sts.SetupInterval[1].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}

	if len(v1Sts.TOR) != 0 {
		x, err := engine.NewFilterRule(engine.MetaPrefix, "TOR", v1Sts.TOR)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.CdrHost) != 0 {
		x, err := engine.NewFilterRule(engine.MetaPrefix, "CdrHost", v1Sts.CdrHost)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.ReqType) != 0 {
		x, err := engine.NewFilterRule(engine.MetaPrefix, "ReqType", v1Sts.ReqType)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Direction) != 0 {
		x, err := engine.NewFilterRule(engine.MetaPrefix, "Direction", v1Sts.Direction)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Category) != 0 {
		x, err := engine.NewFilterRule(engine.MetaPrefix, "Category", v1Sts.Category)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Account) != 0 {
		x, err := engine.NewFilterRule(engine.MetaPrefix, "Account", v1Sts.Account)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Subject) != 0 {
		x, err := engine.NewFilterRule(engine.MetaPrefix, "Subject", v1Sts.Subject)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Supplier) != 0 {
		x, err := engine.NewFilterRule(engine.MetaPrefix, "Supplier", v1Sts.Supplier)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.UsageInterval) == 1 {
		x, err := engine.NewFilterRule(engine.MetaGreaterOrEqual, "UsageInterval", []string{v1Sts.UsageInterval[0].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	} else if len(v1Sts.UsageInterval) == 2 {
		x, err := engine.NewFilterRule(engine.MetaLessThan, "UsageInterval", []string{v1Sts.UsageInterval[1].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.PddInterval) == 1 {
		x, err := engine.NewFilterRule(engine.MetaGreaterOrEqual, "PddInterval", []string{v1Sts.PddInterval[0].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	} else if len(v1Sts.PddInterval) == 2 {
		x, err := engine.NewFilterRule(engine.MetaLessThan, "PddInterval", []string{v1Sts.PddInterval[1].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Supplier) != 0 {
		x, err := engine.NewFilterRule(engine.MetaPrefix, "Supplier", v1Sts.Supplier)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.DisconnectCause) != 0 {
		x, err := engine.NewFilterRule(engine.MetaPrefix, "DisconnectCause", v1Sts.DisconnectCause)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.MediationRunIds) != 0 {
		x, err := engine.NewFilterRule(engine.MetaPrefix, "MediationRunIds", v1Sts.MediationRunIds)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.RatedSubject) != 0 {
		x, err := engine.NewFilterRule(engine.MetaPrefix, "RatedSubject", v1Sts.RatedSubject)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.CostInterval) == 1 {
		x, err := engine.NewFilterRule(engine.MetaGreaterOrEqual, "CostInterval", []string{strconv.FormatFloat(v1Sts.CostInterval[0], 'f', 6, 64)})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	} else if len(v1Sts.CostInterval) == 2 {
		x, err := engine.NewFilterRule(engine.MetaLessThan, "CostInterval", []string{strconv.FormatFloat(v1Sts.CostInterval[1], 'f', 6, 64)})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	filter = &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     v1Sts.Id,
		Rules:  filters}
	stq = &engine.StatQueueProfile{
		ID:           v1Sts.Id,
		QueueLength:  v1Sts.QueueLength,
		Metrics:      []*utils.MetricWithParams{},
		Tenant:       config.CgrConfig().GeneralCfg().DefaultTenant,
		Blocker:      false,
		Stored:       false,
		ThresholdIDs: []string{},
		FilterIDs:    []string{v1Sts.Id},
	}
	if v1Sts.SaveInterval != 0 {
		stq.Stored = true
	}
	if len(v1Sts.Triggers) != 0 {
		for i := range v1Sts.Triggers {
			stq.ThresholdIDs = append(stq.ThresholdIDs, v1Sts.Triggers[i].ID)
		}
	}
	sq = &engine.StatQueue{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        v1Sts.Id,
		SQMetrics: make(map[string]engine.StatMetric),
	}
	if len(v1Sts.Metrics) != 0 {
		for i := range v1Sts.Metrics {
			if !strings.HasPrefix(v1Sts.Metrics[i], "*") {
				v1Sts.Metrics[i] = "*" + v1Sts.Metrics[i]
			}
			v1Sts.Metrics[i] = strings.ToLower(v1Sts.Metrics[i])

			stq.Metrics = append(stq.Metrics, &utils.MetricWithParams{MetricID: v1Sts.Metrics[i]})
			if metric, err := engine.NewStatMetric(stq.Metrics[i].MetricID, 0, ""); err != nil {
				return nil, nil, nil, err
			} else {
				sq.SQMetrics[stq.Metrics[i].MetricID] = metric
			}
		}
	}
	return filter, sq, stq, nil
}
