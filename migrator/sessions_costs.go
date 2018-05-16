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
	"encoding/json"
	"fmt"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentSessionSCost() (err error) {
	if m.sameStorDB { // no move
		return
	}
	smCosts, err := m.storDBIn.StorDB().GetSMCosts("", "", "", "")
	if err != nil {
		return err
	}
	for _, smCost := range smCosts {
		if err := m.storDBOut.StorDB().SetSMCost(smCost); err != nil {
			return err
		}
		if err := m.storDBIn.StorDB().RemoveSMCost(smCost); err != nil {
			return err
		}
	}
	return
}

func (m *Migrator) migrateSessionSCosts() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	vrs, err = m.storDBOut.StorDB().GetVersions("")
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying OutStorDB for versions", err.Error()))
	} else if len(vrs) == 0 {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UndefinedVersion,
			"version number is not defined for SessionsCosts model")
	}
	switch vrs[utils.SessionSCosts] {
	case 0, 1:
		if err := m.migrateV1SessionSCosts(); err != nil {
			return err
		}
	case 2:
		if err := m.migrateV2SessionSCosts(); err != nil {
			return err
		}
	case current[utils.SessionSCosts]:
		if err := m.migrateCurrentSessionSCost(); err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) migrateV1SessionSCosts() (err error) {
	if err = m.storDBIn.renameV1SMCosts(); err != nil {
		return err
	}
	if m.dryRun != true {
		vrs := engine.Versions{utils.SessionSCosts: 2}
		if err = m.storDBOut.StorDB().SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating SessionSCosts version into StorDB", err.Error()))
		}
	}
	return
}

func (m *Migrator) migrateV2SessionSCosts() (err error) {
	var v2Cost *v2SessionsCost
	for {
		v2Cost, err = m.storDBIn.getV2SMCost()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if v2Cost != nil {
			smCost := v2Cost.V2toV3Cost()
			if m.dryRun != true {
				if err = m.storDBOut.StorDB().SetSMCost(smCost); err != nil {
					return err
				}
				if err = m.storDBIn.remV2SMCost(v2Cost); err != nil {
					return err
				}
				m.stats[utils.SessionSCosts] += 1
			}
		}
	}
	if m.dryRun != true {
		// All done, update version wtih current one
		vrs := engine.Versions{utils.SessionSCosts: engine.CurrentStorDBVersions()[utils.SessionSCosts]}
		if err = m.storDBOut.StorDB().SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating SessionSCosts version into StorDB", err.Error()))
		}
	}
	return
}

type v2SessionsCost struct {
	CGRID       string
	RunID       string
	OriginHost  string
	OriginID    string
	CostSource  string
	Usage       time.Duration
	CostDetails *engine.CallCost
}

func (v2Cost *v2SessionsCost) V2toV3Cost() (cost *engine.SMCost) {
	cost = &engine.SMCost{
		CGRID:       v2Cost.CGRID,
		RunID:       v2Cost.RunID,
		OriginHost:  v2Cost.OriginHost,
		OriginID:    v2Cost.OriginID,
		Usage:       v2Cost.Usage,
		CostSource:  v2Cost.CostSource,
		CostDetails: engine.NewEventCostFromCallCost(v2Cost.CostDetails, v2Cost.CGRID, v2Cost.RunID),
	}
	return
}

func NewV2SessionsCostFromSessionsCostSql(smSql *engine.SessionsCostsSQL) (smV2 *v2SessionsCost, err error) {
	smV2 = new(v2SessionsCost)
	smV2.CGRID = smSql.Cgrid
	smV2.RunID = smSql.RunID
	smV2.OriginHost = smSql.OriginHost
	smV2.OriginID = smSql.OriginID
	smV2.CostSource = smSql.CostSource
	smV2.Usage = time.Duration(smSql.Usage)
	smV2.CostDetails = new(engine.CallCost)
	if err := json.Unmarshal([]byte(smSql.CostDetails), smV2.CostDetails); err != nil {
		return nil, err
	}
	return
}

func (v2Cost *v2SessionsCost) AsSessionsCostSql() (smSql *engine.SessionsCostsSQL) {
	smSql = new(engine.SessionsCostsSQL)
	smSql.Cgrid = v2Cost.CGRID
	smSql.RunID = v2Cost.RunID
	smSql.OriginHost = v2Cost.OriginHost
	smSql.OriginID = v2Cost.OriginID
	smSql.CostSource = v2Cost.CostSource
	smSql.CostDetails = utils.ToJSON(v2Cost.CostDetails)
	smSql.Usage = v2Cost.Usage.Nanoseconds()
	smSql.CreatedAt = time.Now()
	return
}
