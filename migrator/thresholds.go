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
	"log"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type v2ActionTrigger struct {
	ID            string // original csv tag
	UniqueID      string // individual id
	ThresholdType string //*min_event_counter, *max_event_counter, *min_balance_counter, *max_balance_counter, *min_balance, *max_balance, *balance_expired
	// stats: *min_asr, *max_asr, *min_acd, *max_acd, *min_tcd, *max_tcd, *min_acc, *max_acc, *min_tcc, *max_tcc, *min_ddc, *max_ddc
	ThresholdValue float64
	Recurrent      bool          // reset excuted flag each run
	MinSleep       time.Duration // Minimum duration between two executions in case of recurrent triggers
	ExpirationDate time.Time
	ActivationDate time.Time
	//BalanceType       string // *monetary/*voice etc
	Balance           *engine.BalanceFilter //filtru
	Weight            float64
	ActionsID         string
	MinQueuedItems    int // Trigger actions only if this number is hit (stats only) MINHITS
	Executed          bool
	LastExecutionTime time.Time
}

type v2ActionTriggers []*v2ActionTrigger

func (m *Migrator) migratev1ActionTriggers() (err error) {
	var vrs engine.Versions
	if m.dm.DataDB() == nil {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.NoStorDBConnection,
			"no connection to datadb")
	}
	vrs, err = m.dm.DataDB().GetVersions(utils.TBLVersions)
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
	} else if len(vrs) == 0 {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UndefinedVersion,
			"version number is not defined for Stats model")
	}
	if vrs[utils.Thresholds] != 1 { // Right now we only support migrating from version 1
		log.Print("Wrong version")
		return
	}
	var v2ACT *v2ActionTrigger
	for {
		v2ACT, err = m.oldDataDB.getV2ActionTrigger()
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
				if err := m.dm.SetFilter(filter); err != nil {
					return err
				}
				if err := m.dm.SetThreshold(th); err != nil {
					return err
				}
				if err := m.dm.SetThresholdProfile(thp); err != nil {
					return err
				}
				m.stats[utils.Thresholds] += 1
			}
		}
	}
	if m.dryRun != true {
		// All done, update version wtih current one
		vrs := engine.Versions{utils.Thresholds: engine.CurrentStorDBVersions()[utils.Thresholds]}
		if err = m.dm.DataDB().SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating Thresholds version into dataDB", err.Error()))
		}
	}
	return
}
func (v2ATR v2ActionTrigger) AsThreshold() (thp *engine.ThresholdProfile, th *engine.Threshold, filter *engine.Filter, err error) {
	var filters []*engine.RequestFilter
	if *v2ATR.Balance.ID != "" {
		if v2ATR.Balance.Directions != nil {
			x, err := engine.NewRequestFilter(engine.MetaRSRFields, "Directions", v2ATR.Balance.Directions.Slice())
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
			x, err := engine.NewRequestFilter(engine.MetaDestinations, "DestinationIDs", v2ATR.Balance.DestinationIDs.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.RatingSubject != nil { //MetaLess /RSRfields
			x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "RatingSubject", []string{*v2ATR.Balance.RatingSubject})
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.Categories != nil { //MetaLess /RSRfields
			x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "Categories", v2ATR.Balance.Categories.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.SharedGroups != nil { //MetaLess /RSRfields
			x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "SharedGroups", v2ATR.Balance.SharedGroups.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.TimingIDs != nil { //MetaLess /RSRfields
			x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "TimingIDs", v2ATR.Balance.TimingIDs.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
	}
	filter = &engine.Filter{Tenant: config.CgrConfig().DefaultTenant, ID: *v2ATR.Balance.ID, RequestFilters: filters}

	th = &engine.Threshold{
		Tenant: config.CgrConfig().DefaultTenant,
		ID:     v2ATR.ID,
	}

	thp = &engine.ThresholdProfile{
		ID:                 v2ATR.ID,
		Tenant:             config.CgrConfig().DefaultTenant,
		Weight:             v2ATR.Weight,
		ActivationInterval: &utils.ActivationInterval{ActivationTime: v2ATR.ActivationDate, ExpiryTime: v2ATR.ExpirationDate},
		FilterIDs:          []string{filter.ID},
		MinSleep:           v2ATR.MinSleep,
	}
	return thp, th, filter, nil
}

func (m *Migrator) SasThreshold(v2ATR *engine.ActionTrigger) (err error) {
	var vrs engine.Versions
	if m.dm.DataDB() == nil {
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
			if err := m.dm.SetFilter(filter); err != nil {
				return err
			}
		}
		if err := m.dm.SetThreshold(th); err != nil {
			return err
		}
		if err := m.dm.SetThresholdProfile(thp); err != nil {
			return err
		}
		m.stats[utils.Thresholds] += 1
	}
	// All done, update version wtih current one
	vrs = engine.Versions{utils.Thresholds: engine.CurrentStorDBVersions()[utils.Thresholds]}
	if err = m.dm.DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating Thresholds version into dataDB", err.Error()))
	}
	return
}

func AsThreshold2(v2ATR engine.ActionTrigger) (thp *engine.ThresholdProfile, th *engine.Threshold, filter *engine.Filter, err error) {
	var filterIDS []string
	var filters []*engine.RequestFilter
	if v2ATR.Balance.ID != nil && *v2ATR.Balance.ID != "" {
		if v2ATR.Balance.Directions != nil {
			x, err := engine.NewRequestFilter(engine.MetaRSRFields, "Directions", v2ATR.Balance.Directions.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.DestinationIDs != nil { //MetaLess /RSRfields
			x, err := engine.NewRequestFilter(engine.MetaDestinations, "DestinationIDs", v2ATR.Balance.DestinationIDs.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.RatingSubject != nil { //MetaLess /RSRfields
			x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "RatingSubject", []string{*v2ATR.Balance.RatingSubject})
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.Categories != nil { //MetaLess /RSRfields
			x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "Categories", v2ATR.Balance.Categories.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.SharedGroups != nil { //MetaLess /RSRfields
			x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "SharedGroups", v2ATR.Balance.SharedGroups.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		if v2ATR.Balance.TimingIDs != nil { //MetaLess /RSRfields
			x, err := engine.NewRequestFilter(engine.MetaStringPrefix, "TimingIDs", v2ATR.Balance.TimingIDs.Slice())
			if err != nil {
				return nil, nil, nil, err
			}
			filters = append(filters, x)
		}
		filter = &engine.Filter{Tenant: config.CgrConfig().DefaultTenant, ID: *v2ATR.Balance.ID, RequestFilters: filters}
		filterIDS = append(filterIDS, filter.ID)
	}
	th = &engine.Threshold{
		Tenant: config.CgrConfig().DefaultTenant,
		ID:     v2ATR.ID,
	}

	thp = &engine.ThresholdProfile{
		ID:                 v2ATR.ID,
		Tenant:             config.CgrConfig().DefaultTenant,
		Weight:             v2ATR.Weight,
		ActivationInterval: &utils.ActivationInterval{ActivationTime: v2ATR.ActivationDate, ExpiryTime: v2ATR.ExpirationDate},
		FilterIDs:          filterIDS,
		MinSleep:           v2ATR.MinSleep,
	}

	return thp, th, filter, nil
}
