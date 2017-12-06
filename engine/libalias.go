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

type AliasEntry struct {
	FieldName string
	Initial   string
	Alias     string
}

type AliasProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval    // Activation interval
	Aliases            map[string]map[string]string // map[FieldName][InitialValue]AliasValue
	Weight             float64
}

func (als *AliasProfile) TenantID() string {
	return utils.ConcatenatedKey(als.Tenant, als.ID)
}

// AliasProfiles is a sortable list of Alias profiles
type AliasProfiles []*AliasProfile

// Sort is part of sort interface, sort based on Weight
func (aps AliasProfiles) Sort() {
	sort.Slice(aps, func(i, j int) bool { return aps[i].Weight > aps[j].Weight })
}

type ExternalAliasProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	Aliases            []*AliasEntry
	Weight             float64
}

func (eap *ExternalAliasProfile) AsAliasProfile() *AliasProfile {
	alsPrf := &AliasProfile{
		Tenant:             eap.Tenant,
		ID:                 eap.ID,
		Weight:             eap.Weight,
		FilterIDs:          eap.FilterIDs,
		ActivationInterval: eap.ActivationInterval,
	}
	alsMap := make(map[string]map[string]string)
	for _, als := range eap.Aliases {
		alsMap[als.FieldName] = make(map[string]string)
		alsMap[als.FieldName][als.Initial] = als.Alias
	}
	alsPrf.Aliases = alsMap
	return alsPrf
}

func NewExternalAliasProfileFromAliasProfile(alsPrf *AliasProfile) *ExternalAliasProfile {
	extals := &ExternalAliasProfile{
		Tenant:             alsPrf.Tenant,
		ID:                 alsPrf.ID,
		Weight:             alsPrf.Weight,
		ActivationInterval: alsPrf.ActivationInterval,
		FilterIDs:          alsPrf.FilterIDs,
	}
	for key, val := range alsPrf.Aliases {
		for key2, val2 := range val {
			extals.Aliases = append(extals.Aliases, &AliasEntry{
				FieldName: key,
				Initial:   key2,
				Alias:     val2,
			})
		}
	}
	return extals

}
