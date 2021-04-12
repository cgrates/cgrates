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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentThresholds() (err error) {
	var ids []string
	//Thresholds
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(context.TODO(), utils.ThresholdPrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.ThresholdPrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating thresholds", id)
		}
		ths, err := m.dmIN.DataManager().GetThreshold(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if ths == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetThreshold(ths, 0, true); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveThreshold(tntID[0], tntID[1], utils.NonTransactional); err != nil {
			return err
		}
		m.stats[utils.Thresholds]++
	}
	//ThresholdProfiles
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(context.TODO(), utils.ThresholdProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.ThresholdProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating threshold profiles", id)
		}
		ths, err := m.dmIN.DataManager().GetThresholdProfile(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if ths == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetThresholdProfile(ths, true); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveThresholdProfile(tntID[0], tntID[1], utils.NonTransactional, false); err != nil {
			return err
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
				if err = m.dmOut.DataManager().SetThreshold(th, 0, true); err != nil {
					return
				}
			}
			if err = m.dmOut.DataManager().SetThresholdProfile(v4, true); err != nil {
				return
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
