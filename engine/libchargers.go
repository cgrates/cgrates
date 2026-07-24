// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"sort"

	"github.com/cgrates/cgrates/utils"
)

// ChargerProfile is the config for one Charger
type ChargerProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	RunID              string
	AttributeIDs       []string // perform data aliasing based on these Attributes
	Weight             float64
}

func (cP *ChargerProfile) TenantID() string {
	return utils.ConcatenatedKey(cP.Tenant, cP.ID)
}

// ChargerProfiles is a sortable list of Charger profiles
type ChargerProfiles []*ChargerProfile

// Sort is part of sort interface, sort based on Weight
func (cps ChargerProfiles) Sort() {
	sort.Slice(cps, func(i, j int) bool { return cps[i].Weight > cps[j].Weight })
}
