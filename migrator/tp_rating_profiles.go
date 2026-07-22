// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPratingprofiles() (err error) {

	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPRatingProfiles)
	if err != nil {
		return err
	}

	for _, tpid := range tpids {
		ratingProfile, err := m.storDBIn.StorDB().GetTPRatingProfiles(&utils.TPRatingProfile{TPid: tpid})
		if err != nil {
			return err
		}
		if ratingProfile != nil {
			if !m.dryRun {
				if err := m.storDBOut.StorDB().SetTPRatingProfiles(ratingProfile); err != nil {
					return err
				}
				for _, ratPrf := range ratingProfile {
					if err := m.storDBIn.StorDB().RemTpData(utils.TBLTPRatingProfiles, ratPrf.TPid,
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
	if vrs, err = m.getVersions(utils.TpRatingProfiles); err != nil {
		return
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
	return m.ensureIndexesStorDB(utils.TBLTPRatingProfiles)
}
