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
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type v2ActionTrigger struct {
	ID                string // original csv tag
	UniqueID          string // individual id
	ThresholdType     string //*min_event_counter, *max_event_counter, *min_balance_counter, *max_balance_counter, *min_balance, *max_balance, *balance_expired
	ThresholdValue    float64
	Recurrent         bool          // reset excuted flag each run
	MinSleep          time.Duration // Minimum duration between two executions in case of recurrent triggers
	ExpirationDate    time.Time
	ActivationDate    time.Time
	Balance           *engine.BalanceFilter //filtru
	Weight            float64
	ActionsID         string
	MinQueuedItems    int // Trigger actions only if this number is hit (stats only) MINHITS
	Executed          bool
	LastExecutionTime time.Time
}

func (m *Migrator) migrateCurrentThresholds() (err error) {
	var ids []string
	//Thresholds
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.ThresholdPrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating thresholds", id)
		}

	}
	//ThresholdProfiles
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ThresholdProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.ThresholdProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating threshold profiles", id)
		}
		thps, err := m.dmIN.DataManager().GetThresholdProfile(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		ths, err := m.dmIN.DataManager().GetThreshold(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if thps == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetThresholdProfile(thps, true); err != nil {
			return err
		}
		// update the threshold in the new DB
		if ths != nil {
			if err := m.dmOut.DataManager().SetThreshold(ths); err != nil {
				return err
			}
		}
		if err := m.dmIN.DataManager().RemoveThresholdProfile(tntID[0], tntID[1], utils.NonTransactional, false); err != nil {
			return err
		}
		m.stats[utils.Thresholds]++
	}
	return
}

func (m *Migrator) migrateV2ActionTriggers() (thp *engine.ThresholdProfile, th *engine.Threshold, filter *engine.Filter, err error) {
	var v2ACT *v2ActionTrigger
	if v2ACT, err = m.dmIN.getV2ActionTrigger(); err != nil {
		return
	}
	if v2ACT.ID != "" {
		if thp, th, filter, err = v2ACT.AsThreshold(); err != nil {
			return
		}
	}
	return
}

func (m *Migrator) removeV2Thresholds() (err error) {
	var v2T *v2Threshold
	for {
		v2T, err = m.dmIN.getV2ThresholdProfile()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if err = m.dmIN.remV2ThresholdProfile(v2T.Tenant, v2T.ID); err != nil {
			return err
		}
	}
	return
}

func (m *Migrator) migrateV2Thresholds() (v3 *engine.ThresholdProfile, err error) {
	var v2T *v2Threshold
	if v2T, err = m.dmIN.getV2ThresholdProfile(); err != nil {
		return
	}
	if v2T == nil {
		return
	}
	v3 = v2T.V2toV3Threshold()
	return
}

func (m *Migrator) migrateThresholds() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	if vrs, err = m.getVersions(utils.Thresholds); err != nil {
		return
	}
	migrated := true
	migratedFrom := 0
	var th *engine.Threshold
	var filter *engine.Filter
	var v3 *engine.ThresholdProfile
	var v4 *engine.ThresholdProfile
	for {
		version := vrs[utils.Thresholds]
		migratedFrom = int(version)
		for {
			switch version {
			default:
				return fmt.Errorf("Unsupported version %v", version)
			case current[utils.Thresholds]:
				migrated = false
				if m.sameDataDB {
					break
				}
				if err = m.migrateCurrentThresholds(); err != nil {
					return
				}
			case 1:
				if v3, th, filter, err = m.migrateV2ActionTriggers(); err != nil && err != utils.ErrNoMoreData {
					return
				}
				version = 3
			case 2:
				if v3, err = m.migrateV2Thresholds(); err != nil && err != utils.ErrNoMoreData {
					return
				}
				version = 3
			case 3:
				if v4, err = m.migrateV3ToV4Threshold(v3); err != nil && err != utils.ErrNoMoreData {
					return
				} else if err == utils.ErrNoMoreData {
					break
				}
				version = 4
			}
			if version == current[utils.Thresholds] || err == utils.ErrNoMoreData {
				break
			}
		}
		if err == utils.ErrNoMoreData || !migrated {
			break
		}

		if !m.dryRun {
			//set threshond
			if migratedFrom == 1 {
				if err = m.dmOut.DataManager().SetFilter(filter, true); err != nil {
					return
				}
			}
			if err = m.dmOut.DataManager().SetThresholdProfile(v4, true); err != nil {
				return
			}
			if migratedFrom == 1 { // do it after SetThresholdProfile to overwrite the created threshold
				if err = m.dmOut.DataManager().SetThreshold(th); err != nil {
					return
				}
			}
		}
		m.stats[utils.Thresholds]++
	}
	if m.dryRun || !migrated {
		return nil
	}
	// remove old threshonds
	if !m.sameDataDB && migratedFrom == 2 {
		if err = m.removeV2Thresholds(); err != nil && err != utils.ErrNoMoreData {
			return
		}
	}
	// All done, update version wtih current one
	if err = m.setVersions(utils.Thresholds); err != nil {
		return
	}
	return m.ensureIndexesDataDB(engine.ColTps)
}

func (v2ATR v2ActionTrigger) AsThreshold() (thp *engine.ThresholdProfile, th *engine.Threshold, filter *engine.Filter, err error) {
	var filterIDS []string
	var filters []*engine.FilterRule
	if v2ATR.Balance.ID != nil && *v2ATR.Balance.ID != "" {
		//TO DO:
		// if v2ATR.Balance.ExpirationDate != nil { //MetaLess
		// 	x, err := engine.NewRequestFilter(utils.MetaTimings, "ExpirationDate", v2ATR.Balance.ExpirationDate)
		// 	if err != nil {
		// 		return nil, nil, err
		// 	}
		// 	filters = append(filters, x)
		// }
		// if v2ATR.Balance.Weight != nil { //MetaLess /MetaRSRFields
		// 	x, err := engine.NewRequestFilter(utils.MetaLessOrEqual, "Weight", []string{strconv.FormatFloat(*v2ATR.Balance.Weight, 'f', 6, 64)})
		// 	if err != nil {
		// 		return nil, nil, err
		// 	}
		// 	filters = append(filters, x)
		// }
		if v2ATR.Balance.DestinationIDs != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(utils.MetaDestinations, "DestinationIDs", v2ATR.Balance.DestinationIDs.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.RatingSubject != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(utils.MetaPrefix, "RatingSubject", []string{*v2ATR.Balance.RatingSubject})
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.Categories != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(utils.MetaPrefix, "Categories", v2ATR.Balance.Categories.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.SharedGroups != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(utils.MetaPrefix, "SharedGroups", v2ATR.Balance.SharedGroups.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.TimingIDs != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(utils.MetaPrefix, "TimingIDs", v2ATR.Balance.TimingIDs.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}

		filter = &engine.Filter{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     *v2ATR.Balance.ID,
			Rules:  filters}
		filterIDS = append(filterIDS, filter.ID)

	}
	thp = &engine.ThresholdProfile{
		ID:     v2ATR.ID,
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		Weight: v2ATR.Weight,
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: v2ATR.ActivationDate,
			ExpiryTime:     v2ATR.ExpirationDate},
		FilterIDs: []string{},
		MinSleep:  v2ATR.MinSleep,
	}
	th = &engine.Threshold{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     v2ATR.ID,
	}
	return thp, th, filter, nil
}

func (m *Migrator) SasThreshold(v2ATR *engine.ActionTrigger) (err error) {
	var vrs engine.Versions
	if m.dmOut.DataManager().DataDB() == nil {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.NoStorDBConnection,
			"no connection to datadb")
	}
	if v2ATR.ID != "" {
		thp, th, filter, err := AsThreshold2(*v2ATR)
		if err != nil {
			return err
		}
		if filter != nil {
			if err := m.dmOut.DataManager().SetFilter(filter, true); err != nil {
				return err
			}
		}
		if err := m.dmOut.DataManager().SetThresholdProfile(thp, true); err != nil {
			return err
		}
		if err := m.dmOut.DataManager().SetThreshold(th); err != nil {
			return err
		}
		m.stats[utils.Thresholds]++
	}
	// All done, update version wtih current one
	vrs = engine.Versions{utils.Thresholds: engine.CurrentStorDBVersions()[utils.Thresholds]}
	if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating Thresholds version into dataDB", err.Error()))
	}
	return
}

func AsThreshold2(v2ATR engine.ActionTrigger) (thp *engine.ThresholdProfile, th *engine.Threshold, filter *engine.Filter, err error) {
	var filterIDS []string
	var filters []*engine.FilterRule
	if v2ATR.Balance.ID != nil && *v2ATR.Balance.ID != "" {
		if v2ATR.Balance.DestinationIDs != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(utils.MetaDestinations, "DestinationIDs", v2ATR.Balance.DestinationIDs.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.RatingSubject != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(utils.MetaPrefix, "RatingSubject", []string{*v2ATR.Balance.RatingSubject})
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.Categories != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(utils.MetaPrefix, "Categories", v2ATR.Balance.Categories.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.SharedGroups != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(utils.MetaPrefix, "SharedGroups", v2ATR.Balance.SharedGroups.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.TimingIDs != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(utils.MetaPrefix, "TimingIDs", v2ATR.Balance.TimingIDs.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		filter = &engine.Filter{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     *v2ATR.Balance.ID,
			Rules:  filters}
		filterIDS = append(filterIDS, filter.ID)
	}
	th = &engine.Threshold{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     v2ATR.ID,
	}

	thp = &engine.ThresholdProfile{
		ID:                 v2ATR.ID,
		Tenant:             config.CgrConfig().GeneralCfg().DefaultTenant,
		Weight:             v2ATR.Weight,
		ActivationInterval: &utils.ActivationInterval{ActivationTime: v2ATR.ActivationDate, ExpiryTime: v2ATR.ExpirationDate},
		FilterIDs:          filterIDS,
		MinSleep:           v2ATR.MinSleep,
	}

	return thp, th, filter, nil
}

type v2Threshold struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Time when this limit becomes active and expires
	Recurrent          bool
	MinHits            int
	MinSleep           time.Duration
	Blocker            bool    // blocker flag to stop processing on filters matched
	Weight             float64 // Weight to sort the thresholds
	ActionIDs          []string
	Async              bool
}

func (v2T v2Threshold) V2toV3Threshold() (th *engine.ThresholdProfile) {
	th = &engine.ThresholdProfile{
		Tenant:             v2T.Tenant,
		ID:                 v2T.ID,
		FilterIDs:          v2T.FilterIDs,
		ActivationInterval: v2T.ActivationInterval,
		MinHits:            v2T.MinHits,
		MinSleep:           v2T.MinSleep,
		Blocker:            v2T.Blocker,
		Weight:             v2T.Weight,
		ActionIDs:          v2T.ActionIDs,
		Async:              v2T.Async,
	}
	th.MaxHits = 1
	if v2T.Recurrent {
		th.MaxHits = -1
	}
	return
}

func (m *Migrator) migrateV3ToV4Threshold(v3sts *engine.ThresholdProfile) (v4Cpp *engine.ThresholdProfile, err error) {
	if v3sts == nil {
		// read data from DataDB
		if v3sts, err = m.dmIN.getV3ThresholdProfile(); err != nil {
			return
		}
	}
	if v3sts.FilterIDs, err = migrateInlineFilterV4(v3sts.FilterIDs); err != nil {
		return
	}
	return v3sts, nil
}
