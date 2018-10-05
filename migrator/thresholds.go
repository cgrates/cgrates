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

type v2ActionTriggers []*v2ActionTrigger

func (m *Migrator) migrateCurrentThresholds() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	//Thresholds
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ThresholdPrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.ThresholdPrefix+tenant+":")
		ths, err := m.dmIN.DataManager().GetThreshold(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if ths != nil {
			if m.dryRun != true {
				if err := m.dmOut.DataManager().SetThreshold(ths); err != nil {
					return err
				}
				m.stats[utils.Thresholds] += 1
			}
		}
	}
	//ThresholdProfiles
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.ThresholdProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.ThresholdProfilePrefix+tenant+":")
		ths, err := m.dmIN.DataManager().GetThresholdProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if ths != nil {
			if m.dryRun != true {
				if err := m.dmOut.DataManager().SetThresholdProfile(ths, true); err != nil {
					return err
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateV2ActionTriggers() (err error) {
	var v2ACT *v2ActionTrigger
	for {
		v2ACT, err = m.dmIN.getV2ActionTrigger()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if v2ACT.ID != "" {
			thp, th, filter, err := v2ACT.AsThreshold()
			if err != nil {
				return err
			}
			if m.dryRun != true {
				if err := m.dmOut.DataManager().SetFilter(filter); err != nil {
					return err
				}
				if err := m.dmOut.DataManager().SetThreshold(th); err != nil {
					return err
				}
				if err := m.dmOut.DataManager().SetThresholdProfile(thp, true); err != nil {
					return err
				}
				m.stats[utils.Thresholds] += 1
			}
		}
	}
	if m.dryRun != true {
		// All done, update version wtih current one
		vrs := engine.Versions{utils.Thresholds: engine.CurrentStorDBVersions()[utils.Thresholds]}
		if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating Thresholds version into dataDB", err.Error()))
		}
	}
	return
}

func (m *Migrator) migrateV2Thresholds() (err error) {
	var v2T *v2Threshold
	for {
		v2T, err = m.dmIN.getV2ThresholdProfile()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if v2T != nil {
			th := v2T.V2toV3Threshold()
			if m.dryRun != true {
				if err = m.dmIN.remV2ThresholdProfile(v2T.Tenant, v2T.ID); err != nil {
					return err
				}
				if err = m.dmOut.DataManager().SetThresholdProfile(th, true); err != nil {
					return err
				}
				m.stats[utils.Thresholds] += 1
			}
		}
	}
	if m.dryRun != true {
		// All done, update version wtih current one
		vrs := engine.Versions{utils.Thresholds: engine.CurrentDataDBVersions()[utils.Thresholds]}
		if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating Thresholds version into dataDB", err.Error()))
		}
	}
	return
}

func (m *Migrator) migrateThresholds() (err error) {
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
	switch vrs[utils.Thresholds] {
	case current[utils.Thresholds]:
		if m.sameDataDB {
			return
		}
		if err := m.migrateCurrentThresholds(); err != nil {
			return err
		}
		return

	case 1:
		return m.migrateV2ActionTriggers()

	case 2:
		return m.migrateV2Thresholds()
	}
	return
}

func (v2ATR v2ActionTrigger) AsThreshold() (thp *engine.ThresholdProfile, th *engine.Threshold, filter *engine.Filter, err error) {
	var filterIDS []string
	var filters []*engine.FilterRule
	if v2ATR.Balance.ID != nil && *v2ATR.Balance.ID != "" {
		if v2ATR.Balance.Directions != nil {
			x, err := engine.NewFilterRule(engine.MetaRSR, "Directions", v2ATR.Balance.Directions.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		//TO DO:
		// if v2ATR.Balance.ExpirationDate != nil { //MetaLess
		// 	x, err := engine.NewRequestFilter(engine.MetaTimings, "ExpirationDate", v2ATR.Balance.ExpirationDate)
		// 	if err != nil {
		// 		return nil, nil, err
		// 	}
		// 	filters = append(filters, x)
		// }
		// if v2ATR.Balance.Weight != nil { //MetaLess /MetaRSRFields
		// 	x, err := engine.NewRequestFilter(engine.MetaLessOrEqual, "Weight", []string{strconv.FormatFloat(*v2ATR.Balance.Weight, 'f', 6, 64)})
		// 	if err != nil {
		// 		return nil, nil, err
		// 	}
		// 	filters = append(filters, x)
		// }
		if v2ATR.Balance.DestinationIDs != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(engine.MetaDestinations, "DestinationIDs", v2ATR.Balance.DestinationIDs.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.RatingSubject != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(engine.MetaPrefix, "RatingSubject", []string{*v2ATR.Balance.RatingSubject})
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.Categories != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(engine.MetaPrefix, "Categories", v2ATR.Balance.Categories.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.SharedGroups != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(engine.MetaPrefix, "SharedGroups", v2ATR.Balance.SharedGroups.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.TimingIDs != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(engine.MetaPrefix, "TimingIDs", v2ATR.Balance.TimingIDs.Slice())
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
			if err := m.dmOut.DataManager().SetFilter(filter); err != nil {
				return err
			}
		}
		if err := m.dmOut.DataManager().SetThreshold(th); err != nil {
			return err
		}
		if err := m.dmOut.DataManager().SetThresholdProfile(thp, true); err != nil {
			return err
		}
		m.stats[utils.Thresholds] += 1
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
		if v2ATR.Balance.Directions != nil {
			x, err := engine.NewFilterRule(engine.MetaRSR, "Directions", v2ATR.Balance.Directions.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.DestinationIDs != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(engine.MetaDestinations, "DestinationIDs", v2ATR.Balance.DestinationIDs.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.RatingSubject != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(engine.MetaPrefix, "RatingSubject", []string{*v2ATR.Balance.RatingSubject})
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.Categories != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(engine.MetaPrefix, "Categories", v2ATR.Balance.Categories.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.SharedGroups != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(engine.MetaPrefix, "SharedGroups", v2ATR.Balance.SharedGroups.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.TimingIDs != nil { //MetaLess /RSRfields
			x, err := engine.NewFilterRule(engine.MetaPrefix, "TimingIDs", v2ATR.Balance.TimingIDs.Slice())
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
	if v2T.Recurrent == true {
		th.MaxHits = -1
	} else {
		th.MaxHits = 1
	}
	return
}
