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

var dcGetMapKeys = func(m utils.StringMap) (keys []string) {
	keys = make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	// sort.Strings(keys)
	return keys
}

type v1DerivedCharger struct {
	RunID      string // Unique runId in the chain
	RunFilters string // Only run the charger if all the filters match

	RequestTypeField     string // Field containing request type info, number in case of csv source, '^' as prefix in case of static values
	DirectionField       string // Field containing direction info
	TenantField          string // Field containing tenant info
	CategoryField        string // Field containing tor info
	AccountField         string // Field containing account information
	SubjectField         string // Field containing subject information
	DestinationField     string // Field containing destination information
	SetupTimeField       string // Field containing setup time information
	PDDField             string // Field containing setup time information
	AnswerTimeField      string // Field containing answer time information
	UsageField           string // Field containing usage information
	SupplierField        string // Field containing supplier information
	DisconnectCauseField string // Field containing disconnect cause information
	CostField            string // Field containing cost information
	PreRatedField        string // Field marking rated request in CDR
}

type v1DerivedChargers struct {
	DestinationIDs utils.StringMap
	Chargers       []*v1DerivedCharger
}

type v1DerivedChargersWithKey struct {
	Key   string
	Value *v1DerivedChargers
}

func fieldinfo2Attribute(attr []*engine.Attribute, fieldName, fieldInfo string) (a []*engine.Attribute) {
	var rp config.RSRParsers
	if fieldInfo == utils.MetaDefault || len(fieldInfo) == 0 {
		return attr
	}
	if strings.HasPrefix(fieldInfo, utils.StaticValuePrefix) {
		fieldInfo = fieldInfo[1:]
	}
	var err error
	if rp, err = config.NewRSRParsers(fieldInfo, utils.INFIELD_SEP); err != nil {
		utils.Logger.Err(fmt.Sprintf("On Migrating rule: <%s>, error: %s", fieldInfo, err.Error()))
		return attr
	}
	var path string
	if fieldName == utils.EmptyString {
		return attr // do not append attribute if fieldName is empty
	}
	path = utils.MetaReq + utils.NestingSep + fieldName
	return append(attr, &engine.Attribute{
		Path:  path,
		Value: rp,
		Type:  utils.MetaVariable,
	})
}

func derivedChargers2AttributeProfile(dc *v1DerivedCharger, tenant, key string, filters []string) (attr *engine.AttributeProfile) {
	attr = &engine.AttributeProfile{
		Tenant:             tenant,
		ID:                 key,
		Contexts:           []string{utils.MetaChargers},
		FilterIDs:          filters,
		ActivationInterval: nil,
		Attributes:         make([]*engine.Attribute, 0),
		Blocker:            false,
		Weight:             10,
	}
	attr.Attributes = fieldinfo2Attribute(attr.Attributes, utils.RequestType, dc.RequestTypeField)
	attr.Attributes = fieldinfo2Attribute(attr.Attributes, utils.Direction, dc.DirectionField) //still in use?
	attr.Attributes = fieldinfo2Attribute(attr.Attributes, utils.Tenant, dc.TenantField)
	attr.Attributes = fieldinfo2Attribute(attr.Attributes, utils.Category, dc.CategoryField)
	attr.Attributes = fieldinfo2Attribute(attr.Attributes, utils.AccountField, dc.AccountField)
	attr.Attributes = fieldinfo2Attribute(attr.Attributes, utils.Subject, dc.SubjectField)
	attr.Attributes = fieldinfo2Attribute(attr.Attributes, utils.Destination, dc.DestinationField)
	attr.Attributes = fieldinfo2Attribute(attr.Attributes, utils.SetupTime, dc.SetupTimeField)
	attr.Attributes = fieldinfo2Attribute(attr.Attributes, utils.PDD, dc.PDDField)
	attr.Attributes = fieldinfo2Attribute(attr.Attributes, utils.AnswerTime, dc.AnswerTimeField)
	attr.Attributes = fieldinfo2Attribute(attr.Attributes, utils.Usage, dc.UsageField)
	attr.Attributes = fieldinfo2Attribute(attr.Attributes, SUPPLIER, dc.SupplierField)
	attr.Attributes = fieldinfo2Attribute(attr.Attributes, utils.DISCONNECT_CAUSE, dc.DisconnectCauseField)
	attr.Attributes = fieldinfo2Attribute(attr.Attributes, utils.Cost, dc.CostField)
	attr.Attributes = fieldinfo2Attribute(attr.Attributes, utils.PreRated, dc.PreRatedField)
	return
}

func derivedChargers2Charger(dc *v1DerivedCharger, tenant string, key string, filters []string) (ch *engine.ChargerProfile) {
	ch = &engine.ChargerProfile{
		Tenant:             tenant,
		ID:                 key,
		FilterIDs:          filters,
		ActivationInterval: nil,
		RunID:              dc.RunID,
		AttributeIDs:       make([]string, 0),
		Weight:             10,
	}

	filter := dc.RunFilters
	if len(filter) != 0 {
		if strings.HasPrefix(filter, utils.StaticValuePrefix) {
			filter = filter[1:]
		}
		if strings.HasPrefix(filter, utils.DynamicDataPrefix) {
			filter = filter[1:]
		}
		flt, err := migrateInlineFilterV4([]string{"*rsr::" + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + filter})
		if err != nil {
			return
		}
		ch.FilterIDs = append(ch.FilterIDs, flt...)
	}
	return
}

func (m *Migrator) derivedChargers2Chargers(dck *v1DerivedChargersWithKey) (err error) {
	// (direction, tenant, category, account, subject)
	skey := utils.SplitConcatenatedKey(dck.Key)
	destination := ""
	if len(dck.Value.DestinationIDs) != 0 {
		destination = fmt.Sprintf("%s:~%s:", utils.MetaDestinations, utils.MetaReq+utils.NestingSep+utils.Destination)
		keys := dcGetMapKeys(dck.Value.DestinationIDs)
		destination += strings.Join(keys, utils.INFIELD_SEP)
	}
	filter := make([]string, 0)

	if len(destination) != 0 {
		filter = append(filter, destination)
	}
	if len(skey[2]) != 0 && skey[2] != utils.MetaAny {
		filter = append(filter, fmt.Sprintf("%s:~%s:%s", utils.MetaString, utils.MetaReq+utils.NestingSep+utils.Category, skey[2]))
	}
	if len(skey[3]) != 0 && skey[3] != utils.MetaAny {
		filter = append(filter, fmt.Sprintf("%s:~%s:%s", utils.MetaString, utils.MetaReq+utils.NestingSep+utils.AccountField, skey[3]))
	}
	if len(skey[4]) != 0 && skey[4] != utils.MetaAny {
		filter = append(filter, fmt.Sprintf("%s:~%s:%s", utils.MetaString, utils.MetaReq+utils.NestingSep+utils.Subject, skey[4]))
	}
	for i, dc := range dck.Value.Chargers {
		attr := derivedChargers2AttributeProfile(dc, skey[1], fmt.Sprintf("%s_%v", dck.Key, i), filter)
		ch := derivedChargers2Charger(dc, skey[1], fmt.Sprintf("%s_%v", dck.Key, i), filter)
		if len(attr.Attributes) != 0 {
			if err = m.dmOut.DataManager().SetAttributeProfile(attr, true); err != nil {
				return err
			}
			ch.AttributeIDs = append(ch.AttributeIDs, attr.ID)
		}
		if err = m.dmOut.DataManager().SetChargerProfile(ch, true); err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) removeV1DerivedChargers() (err error) {
	for {
		var dck *v1DerivedChargersWithKey
		dck, err = m.dmIN.getV1DerivedChargers()
		if err == utils.ErrNoMoreData {
			break
		}
		if err != nil {
			return
		}
		if err = m.dmIN.remV1DerivedChargers(dck.Key); err != nil {
			return
		}
	}
	return
}

func (m *Migrator) migrateV1DerivedChargers() (err error) {
	for {
		var dck *v1DerivedChargersWithKey
		dck, err = m.dmIN.getV1DerivedChargers()
		if err == utils.ErrNoMoreData {
			break
		}
		if err != nil {
			return
		}
		if dck == nil || m.dryRun {
			continue
		}
		if err = m.derivedChargers2Chargers(dck); err != nil {
			return
		}

		m.stats[utils.DerivedChargersV]++
	}
	if m.dryRun {
		return
	}
	if err = m.removeV1DerivedChargers(); err != nil && err != utils.ErrNoMoreData {
		return
	}
	// All done, update version wtih current one
	vrs := engine.Versions{utils.DerivedChargersV: 0}
	if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating DerivedChargers version into dataDB", err.Error()))
	}
	return
}

func (m *Migrator) migrateDerivedChargers() (err error) {
	if err = m.migrateV1DerivedChargers(); err != nil {
		return
	}
	return m.ensureIndexesDataDB(engine.ColCpp, engine.ColAttr)
}
