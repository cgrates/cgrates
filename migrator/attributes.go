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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type v1Attribute struct {
	FieldName  string
	Initial    string
	Substitute string
	Append     bool
}

type v1AttributeProfile struct {
	Tenant             string
	ID                 string
	Contexts           []string // bind this AttributeProfile to multiple contexts
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval          // Activation interval
	Attributes         map[string]map[string]*v1Attribute // map[FieldName][InitialValue]*Attribute
	Weight             float64
}

func (m *Migrator) migrateCurrentAttributeProfile() (err error) {
	var ids []string
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.AttributeProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		idg := strings.TrimPrefix(id, utils.AttributeProfilePrefix+tenant+":")
		attrPrf, err := m.dmIN.DataManager().GetAttributeProfile(tenant, idg, false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if attrPrf != nil {
			if m.dryRun != true {
				if err := m.dmOut.DataManager().SetAttributeProfile(attrPrf, true); err != nil {
					return err
				}
				if err := m.dmIN.DataManager().RemoveAttributeProfile(tenant,
					idg, utils.NonTransactional, false); err != nil {
					return err
				}
				m.stats[utils.Attributes] += 1
			}
		}
	}
	return
}

func (m *Migrator) migrateV1Attributes() (err error) {
	var v1Attr *v1AttributeProfile
	for {
		v1Attr, err = m.dmIN.getV1AttributeProfile()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if v1Attr != nil {
			attrPrf, err := v1Attr.AsAttributeProfile()
			if err != nil {
				return err
			}
			if m.dryRun != true {
				if err := m.dmOut.DataManager().DataDB().SetAttributeProfileDrv(attrPrf); err != nil {
					return err
				}
				if err := m.dmOut.DataManager().SetAttributeProfile(attrPrf, true); err != nil {
					return err
				}
				m.stats[utils.Attributes] += 1
			}
		}
	}
	if m.dryRun != true {
		// All done, update version wtih current one
		vrs := engine.Versions{utils.Attributes: engine.CurrentDataDBVersions()[utils.Attributes]}
		if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating Thresholds version into dataDB", err.Error()))
		}
	}
	return
}

func (m *Migrator) migrateAttributeProfile() (err error) {
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
	switch vrs[utils.Attributes] {
	case current[utils.Attributes]:
		if m.sameDataDB {
			return
		}
		if err := m.migrateCurrentAttributeProfile(); err != nil {
			return err
		}
		return
	case 1:
		if err := m.migrateV1Attributes(); err != nil {
			return err
		}
	}
	return
}

func (v1AttrPrf v1AttributeProfile) AsAttributeProfile() (attrPrf *engine.AttributeProfile, err error) {
	attrPrf = &engine.AttributeProfile{
		Tenant:             v1AttrPrf.Tenant,
		ID:                 v1AttrPrf.ID,
		Contexts:           v1AttrPrf.Contexts,
		FilterIDs:          v1AttrPrf.FilterIDs,
		Weight:             v1AttrPrf.Weight,
		ActivationInterval: v1AttrPrf.ActivationInterval,
	}
	for _, mp := range v1AttrPrf.Attributes {
		for _, attr := range mp {
			initIface := utils.StringToInterface(attr.Initial)
			sbstPrsr, err := config.NewRSRParsers(attr.Substitute, true, config.CgrConfig().GeneralCfg().RsrSepatarot)
			if err != nil {
				return nil, err
			}
			attrPrf.Attributes = append(attrPrf.Attributes, &engine.Attribute{
				FieldName:  attr.FieldName,
				Initial:    initIface,
				Substitute: sbstPrsr,
				Append:     attr.Append,
			})
		}
	}
	return
}
