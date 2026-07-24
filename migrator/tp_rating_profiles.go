// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPratingprofiles() (err error) {

	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPRateProfiles)
	if err != nil {
		return err
	}

	for _, tpid := range tpids {
		ratingProfile, err := m.storDBIn.StorDB().GetTPRatingProfiles(&utils.TPRatingProfile{TPid: tpid})
		if err != nil {
			return err
		}
		if ratingProfile != nil {
			if m.dryRun != true {
				if err := m.storDBOut.StorDB().SetTPRatingProfiles(ratingProfile); err != nil {
					return err
				}
				for _, ratPrf := range ratingProfile {
					if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPRateProfiles, ratPrf.TPid,
						map[string]string{"loadid": ratPrf.LoadId,
							"tenant": ratPrf.Tenant, "category": ratPrf.Category,
							"subject": ratPrf.Subject}); err != nil {
						return err
					}
				}
				m.stats[utils.TpRatingProfiles] += 1
			}
		}
	}
	return
}

func (m *Migrator) migrateTPratingprofiles() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	vrs, err = m.storDBOut.StorDB().GetVersions("")
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
	switch vrs[utils.TpRatingProfiles] {
	case current[utils.TpRatingProfiles]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPratingprofiles(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPRateProfiles)
}
