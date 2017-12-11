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

package engine

import (
	"sort"

	"github.com/cgrates/cgrates/utils"
)

type Substitute struct {
	FieldName string
	Initial   string
	Alias     string
	Append    bool
}

type AttributeProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval         // Activation interval
	Context            string                            // bind this AttributeProfile to specific context
	Substitutes        map[string]map[string]*Substitute // map[FieldName][InitialValue]*Attribute
	Weight             float64
}

func (als *AttributeProfile) TenantID() string {
	return utils.ConcatenatedKey(als.Tenant, als.ID)
}

// AttributeProfiles is a sortable list of Attribute profiles
type AttributeProfiles []*AttributeProfile

// Sort is part of sort interface, sort based on Weight
func (aps AttributeProfiles) Sort() {
	sort.Slice(aps, func(i, j int) bool { return aps[i].Weight > aps[j].Weight })
}

type ExternalAttributeProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	Context            string                    // bind this AttributeProfile to specific context
	Substitute         []*Substitute
	Weight             float64
}

func (eap *ExternalAttributeProfile) AsAttributeProfile() *AttributeProfile {
	alsPrf := &AttributeProfile{
		Tenant:             eap.Tenant,
		ID:                 eap.ID,
		Weight:             eap.Weight,
		FilterIDs:          eap.FilterIDs,
		ActivationInterval: eap.ActivationInterval,
		Context:            eap.Context,
	}
	alsMap := make(map[string]map[string]*Substitute)
	for _, als := range eap.Substitute {
		alsMap[als.FieldName] = make(map[string]*Substitute)
		alsMap[als.FieldName][als.Initial] = als
	}
	alsPrf.Substitutes = alsMap
	return alsPrf
}

func NewExternalAttributeProfileFromAttributeProfile(alsPrf *AttributeProfile) *ExternalAttributeProfile {
	extals := &ExternalAttributeProfile{
		Tenant:             alsPrf.Tenant,
		ID:                 alsPrf.ID,
		Weight:             alsPrf.Weight,
		ActivationInterval: alsPrf.ActivationInterval,
		Context:            alsPrf.Context,
		FilterIDs:          alsPrf.FilterIDs,
	}
	for key, val := range alsPrf.Substitutes {
		for key2, val2 := range val {
			extals.Substitute = append(extals.Substitute, &Substitute{
				FieldName: key,
				Initial:   key2,
				Alias:     val2.Alias,
				Append:    val2.Append,
			})
		}
	}
	return extals
}
