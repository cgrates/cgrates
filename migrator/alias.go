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
	//"log"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type v1Alias struct {
	Direction string
	Tenant    string
	Category  string
	Account   string
	Subject   string
	Context   string
	Values    v1AliasValues
}

var (
	ALIASES_PREFIX = "als_"
	Alias          = "Alias"
)

type v1AliasValues []*v1AliasValue

type v1AliasValue struct {
	DestinationId string //=Destination
	Pairs         v1AliasPairs
	Weight        float64
}

type v1AliasPairs map[string]map[string]string

func (al *v1Alias) SetId(id string) error {
	vals := strings.Split(id, utils.CONCATENATED_KEY_SEP)
	if len(vals) != 6 {
		return utils.ErrInvalidKey
	}
	al.Direction = vals[0]
	al.Tenant = vals[1]
	al.Category = vals[2]
	al.Account = vals[3]
	al.Subject = vals[4]
	al.Context = vals[5]
	return nil
}

func (al *v1Alias) GetId() string {
	return utils.ConcatenatedKey(al.Direction, al.Tenant, al.Category, al.Account, al.Subject, al.Context)
}

func alias2AtttributeProfile(alias *v1Alias, defaultTenant string) *engine.AttributeProfile {
	out := &engine.AttributeProfile{
		Tenant:             alias.Tenant,
		ID:                 alias.GetId(),
		Contexts:           []string{"*any"},
		FilterIDs:          make([]string, 0),
		ActivationInterval: nil,
		Attributes:         make([]*engine.Attribute, 0),
		Blocker:            false,
		Weight:             10,
	}
	if len(out.Tenant) == 0 || out.Tenant == utils.META_ANY {
		out.Tenant = defaultTenant
	}
	if len(alias.Category) != 0 && alias.Category != utils.META_ANY {
		out.FilterIDs = append(out.FilterIDs, "*string:Category:"+alias.Category)
	}
	if len(alias.Account) != 0 && alias.Account != utils.META_ANY {
		out.FilterIDs = append(out.FilterIDs, "*string:Account:"+alias.Account)
	}
	if len(alias.Subject) != 0 && alias.Subject != utils.META_ANY {
		out.FilterIDs = append(out.FilterIDs, "*string:Subject:"+alias.Subject)
	}
	var destination string
	for _, av := range alias.Values {
		if len(destination) == 0 || destination == utils.META_ANY {
			destination = av.DestinationId
		}
		for fieldname, vals := range av.Pairs {
			for initial, substitute := range vals {
				out.Attributes = append(out.Attributes, &engine.Attribute{
					FieldName:  fieldname,
					Initial:    initial,
					Substitute: config.NewRSRParsersMustCompile(substitute, true, utils.INFIELD_SEP),
					Append:     true,
				})
			}
		}
	}
	if len(destination) != 0 && destination != utils.META_ANY {
		out.FilterIDs = append(out.FilterIDs, "*destination:Destination:"+destination)
	}
	return out
}

func (m *Migrator) migrateAlias2Attributes() (err error) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		return err
	}
	defaultTenant := cfg.GeneralCfg().DefaultTenant
	for {
		alias, err := m.dmIN.getV1Alias()
		if err == utils.ErrNoMoreData {
			break
		}
		if err != nil {
			return err
		}
		if alias == nil || m.dryRun {
			continue
		}
		attr := alias2AtttributeProfile(alias, defaultTenant)
		if len(attr.Attributes) == 0 {
			continue
		}
		if err := m.dmIN.remV1Alias(alias.GetId()); err != nil {
			return err
		}
		if err := m.dmOut.DataManager().DataDB().SetAttributeProfileDrv(attr); err != nil {
			return err
		}
		m.stats[Alias] += 1
	}
	if m.dryRun {
		return
	}
	// All done, update version wtih current one
	vrs := engine.Versions{Alias: 0} //engine.CurrentDataDBVersions()[utils.Alias]}
	if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating Alias version into dataDB", err.Error()))
	}
	return
}

// func (m *Migrator) migrateV1Alias() (err error) {
// 	var ids []string
// 	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(ALIASES_PREFIX)
// 	if err != nil {
// 		return err
// 	}
// 	for _, id := range ids {
// 		idg := strings.TrimPrefix(id, ALIASES_PREFIX)
// 		usr, err := m.dmIN.DataManager().DataDB().GetAlias(idg, true, utils.NonTransactional)
// 		if err != nil {
// 			return err
// 		}
// 		if usr == nil || m.dryRun {
// 			continue
// 		}
// 		if err := m.dmOut.DataManager().DataDB().SetAlias(usr, utils.NonTransactional); err != nil {
// 			return err
// 		}
// 		m.stats[utils.Alias] += 1
// 	}
// 	return
// }

func (m *Migrator) migrateAlias() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	vrs, err = m.dmIN.DataManager().DataDB().GetVersions("")
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
	switch vrs[Alias] {
	case current[Alias]:
		if m.sameDataDB {
			return
		}
		return utils.ErrNotImplemented
	case 1:
		return m.migrateAlias2Attributes()
	}
	return
}

func (m *Migrator) migrateReverseAlias() (err error) {
	// var vrs engine.Versions
	// current := engine.CurrentDataDBVersions()
	// vrs, err = m.dmOut.DataDB().GetVersions("")
	// if err != nil {
	// 	return utils.NewCGRError(utils.Migrator,
	// 		utils.ServerErrorCaps,
	// 		err.Error(),
	// 		fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
	// } else if len(vrs) == 0 {
	// 	return utils.NewCGRError(utils.Migrator,
	// 		utils.MandatoryIEMissingCaps,
	// 		utils.UndefinedVersion,
	// 		"version number is not defined for ActionTriggers model")
	// }
	// switch vrs[utils.ReverseAlias] {
	// case current[utils.ReverseAlias]:
	// 	if m.sameDataDB {
	// 		return
	// 	}
	// 	if err := m.migrateCurrentReverseAlias(); err != nil {
	// 		return err
	// 	}
	// 	return
	// }
	return
}
