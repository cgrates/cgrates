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
	"errors"
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
	ToR             []string        // CDRFieldFilter on TORs
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
}

type v1Stats []*v1Stat

func (m *Migrator) moveStatQueueProfile() (err error) {
	//StatQueueProfile
	var ids []string
	if ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.StatQueueProfilePrefix); err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.StatQueueProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating stat queue profiles", id)
		}
		sgs, err := m.dmIN.DataManager().GetStatQueueProfile(tntID[0], tntID[1], false, false, utils.NonTransactional)
		fmt.Println("sgs", utils.ToJSON(sgs))
		if err != nil {
			return err
		}
		if sgs == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetStatQueueProfile(sgs, true); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveStatQueueProfile(tntID[0], tntID[1], utils.NonTransactional, false); err != nil {
			return err
		}
	}
	return
}

func (m *Migrator) migrateCurrentStats() (err error) {
	var ids []string
	//StatQueue
	if ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.StatQueuePrefix); err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.StatQueuePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating stat queues", id)
		}
		sgs, err := m.dmIN.DataManager().GetStatQueue(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {

			return err
		}
		if sgs == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetStatQueue(sgs, nil, 0, nil, 0, true); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveStatQueue(tntID[0], tntID[1], utils.NonTransactional); err != nil {
			return err
		}
		m.stats[utils.StatS]++
	}

	return m.moveStatQueueProfile()
}

func (m *Migrator) migrateV1Stats() (filter *engine.Filter, v2Stats *engine.StatQueue, sts *engine.StatQueueProfile, err error) {
	var v1Sts *v1Stat
	v1Sts, err = m.dmIN.getV1Stats()
	if err != nil {
		return nil, nil, nil, err
	}
	if v1Sts.Id != utils.EmptyString {
		if filter, v2Stats, sts, err = v1Sts.AsStatQP(); err != nil {
			return nil, nil, nil, err
		}
	}
	return
}

func remakeQueue(sq *engine.StatQueue) (out *engine.StatQueue) {
	out = &engine.StatQueue{
		Tenant:    sq.Tenant,
		ID:        sq.ID,
		SQItems:   sq.SQItems,
		SQMetrics: make(map[string]engine.StatMetric),
	}
	for mID, metric := range sq.SQMetrics {
		out.SQMetrics[mID] = metric
	}
	return
}

func (m *Migrator) migrateV2Stats(v2Stats *engine.StatQueue) (v3Stats *engine.StatQueue, err error) {
	if v2Stats == nil {
		// read from DB
		v2Stats, err = m.dmIN.getV2Stats()
		if err != nil {
			return nil, err
		} else if v2Stats == nil {
			return nil, errors.New("Stats NIL")
		}
	}
	v3Stats = remakeQueue(v2Stats)
	return
}

func (m *Migrator) migrateStats() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.StatS); err != nil {
		return
	}
	migrated := true
	var filter *engine.Filter
	var v3sts *engine.StatQueueProfile
	var v4sts *engine.StatQueueProfile
	var v2Stats *engine.StatQueue
	var v3Stats *engine.StatQueue
	for {
		version := vrs[utils.StatS]
		for {
			switch version {
			default:
				return fmt.Errorf("Unsupported version %v", version)
			case current[utils.StatS]:
				migrated = false
				if m.sameDataDB {
					break
				}
				if err = m.migrateCurrentStats(); err != nil {
					return
				}
				version = 3
			case 1: // migrate from V1 to V2
				if filter, v2Stats, v3sts, err = m.migrateV1Stats(); err != nil && err != utils.ErrNoMoreData {
					return
				} else if err == utils.ErrNoMoreData {
					break
				}
				version = 2
			case 2: // migrate from V2 to V3 (actual)
				if v3Stats, err = m.migrateV2Stats(v2Stats); err != nil && err != utils.ErrNoMoreData {
					return
				} else if err == utils.ErrNoMoreData {
					break
				}
				version = 3
			case 3:
				if v4sts, err = m.migrateV3ToV4Stats(v3sts); err != nil && err != utils.ErrNoMoreData {
					return
				} else if err == utils.ErrNoMoreData {
					break
				}
				version = 4
			}
			if version == current[utils.StatS] || err == utils.ErrNoMoreData {
				break
			}
		}
		if err == utils.ErrNoMoreData || !migrated {
			break
		}
		if !m.dryRun {
			if vrs[utils.StatS] == 1 {
				if err = m.dmOut.DataManager().SetFilter(filter, true); err != nil {
					return
				}
			}
			// Set the fresh-migrated Stats into DB
			if err = m.dmOut.DataManager().SetStatQueueProfile(v4sts, true); err != nil {
				return
			}
			if err = m.dmOut.DataManager().SetStatQueue(v3Stats, nil, 0, nil, 0, true); err != nil {
				return
			}
		}
		m.stats[utils.StatS]++
	}
	if m.dryRun || !migrated {
		return nil
	}
	// call the remove function here

	// All done, update version wtih current one
	if err = m.setVersions(utils.StatS); err != nil {
		return err
	}
	return m.ensureIndexesDataDB(engine.ColSqs)
}

func (v1Sts v1Stat) AsStatQP() (filter *engine.Filter, sq *engine.StatQueue, stq *engine.StatQueueProfile, err error) {
	var filters []*engine.FilterRule
	if len(v1Sts.SetupInterval) == 1 {
		x, err := engine.NewFilterRule(utils.MetaGreaterOrEqual,
			"SetupInterval", []string{v1Sts.SetupInterval[0].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	} else if len(v1Sts.SetupInterval) == 2 {
		x, err := engine.NewFilterRule(utils.MetaLessThan,
			"SetupInterval", []string{v1Sts.SetupInterval[1].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}

	if len(v1Sts.ToR) != 0 {
		x, err := engine.NewFilterRule(utils.MetaPrefix, "ToR", v1Sts.ToR)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.CdrHost) != 0 {
		x, err := engine.NewFilterRule(utils.MetaPrefix, "CdrHost", v1Sts.CdrHost)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.ReqType) != 0 {
		x, err := engine.NewFilterRule(utils.MetaPrefix, "ReqType", v1Sts.ReqType)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Direction) != 0 {
		x, err := engine.NewFilterRule(utils.MetaPrefix, "Direction", v1Sts.Direction)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Category) != 0 {
		x, err := engine.NewFilterRule(utils.MetaPrefix, "Category", v1Sts.Category)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Account) != 0 {
		x, err := engine.NewFilterRule(utils.MetaPrefix, "Account", v1Sts.Account)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Subject) != 0 {
		x, err := engine.NewFilterRule(utils.MetaPrefix, "Subject", v1Sts.Subject)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Supplier) != 0 {
		x, err := engine.NewFilterRule(utils.MetaPrefix, "Supplier", v1Sts.Supplier)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.UsageInterval) == 1 {
		x, err := engine.NewFilterRule(utils.MetaGreaterOrEqual, "UsageInterval", []string{v1Sts.UsageInterval[0].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	} else if len(v1Sts.UsageInterval) == 2 {
		x, err := engine.NewFilterRule(utils.MetaLessThan, "UsageInterval", []string{v1Sts.UsageInterval[1].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.PddInterval) == 1 {
		x, err := engine.NewFilterRule(utils.MetaGreaterOrEqual, "PddInterval", []string{v1Sts.PddInterval[0].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	} else if len(v1Sts.PddInterval) == 2 {
		x, err := engine.NewFilterRule(utils.MetaLessThan, "PddInterval", []string{v1Sts.PddInterval[1].String()})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.Supplier) != 0 {
		x, err := engine.NewFilterRule(utils.MetaPrefix, "Supplier", v1Sts.Supplier)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.DisconnectCause) != 0 {
		x, err := engine.NewFilterRule(utils.MetaPrefix, "DisconnectCause", v1Sts.DisconnectCause)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.MediationRunIds) != 0 {
		x, err := engine.NewFilterRule(utils.MetaPrefix, "MediationRunIds", v1Sts.MediationRunIds)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.RatedSubject) != 0 {
		x, err := engine.NewFilterRule(utils.MetaPrefix, "RatedSubject", v1Sts.RatedSubject)
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	}
	if len(v1Sts.CostInterval) == 1 {
		x, err := engine.NewFilterRule(utils.MetaGreaterOrEqual, "CostInterval", []string{strconv.FormatFloat(v1Sts.CostInterval[0], 'f', 6, 64)})
		if err != nil {
			return nil, nil, nil, err
		}
		filters = append(filters, x)
	} else if len(v1Sts.CostInterval) == 2 {
		x, err := engine.NewFilterRule(utils.MetaLessThan, "CostInterval", []string{strconv.FormatFloat(v1Sts.CostInterval[1], 'f', 6, 64)})
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
		Metrics:      make([]*engine.MetricWithFilters, 0),
		Tenant:       config.CgrConfig().GeneralCfg().DefaultTenant,
		Blocker:      false,
		Stored:       false,
		ThresholdIDs: []string{},
		FilterIDs:    []string{v1Sts.Id},
	}
	if v1Sts.SaveInterval != 0 {
		stq.Stored = true
	}
	sq = &engine.StatQueue{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        v1Sts.Id,
		SQMetrics: make(map[string]engine.StatMetric),
	}
	if len(v1Sts.Metrics) != 0 {
		for i := range v1Sts.Metrics {
			if !strings.HasPrefix(v1Sts.Metrics[i], utils.Meta) {
				v1Sts.Metrics[i] = utils.Meta + v1Sts.Metrics[i]
			}
			v1Sts.Metrics[i] = strings.ToLower(v1Sts.Metrics[i])
			stq.Metrics = append(stq.Metrics, &engine.MetricWithFilters{MetricID: v1Sts.Metrics[i]})
			if sq.SQMetrics[stq.Metrics[i].MetricID], err = engine.NewStatMetric(stq.Metrics[i].MetricID, 0, []string{}); err != nil {
				return nil, nil, nil, err
			}
		}
	}
	return filter, sq, stq, nil
}

func (m *Migrator) migrateV3ToV4Stats(v3sts *engine.StatQueueProfile) (v4Cpp *engine.StatQueueProfile, err error) {
	if v3sts == nil {
		// read data from DataDB
		if v3sts, err = m.dmIN.getV3Stats(); err != nil {
			return
		}
	}
	if v3sts.FilterIDs, err = migrateInlineFilterV4(v3sts.FilterIDs); err != nil {
		return
	}
	return v3sts, nil
}
