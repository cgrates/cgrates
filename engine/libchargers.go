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

// Clone method for ChargerProfile
func (cp *ChargerProfile) Clone() *ChargerProfile {
	if cp == nil {
		return nil
	}
	clone := &ChargerProfile{
		Tenant: cp.Tenant,
		ID:     cp.ID,
		RunID:  cp.RunID,
		Weight: cp.Weight,
	}
	if cp.FilterIDs != nil {
		clone.FilterIDs = make([]string, len(cp.FilterIDs))
		copy(clone.FilterIDs, cp.FilterIDs)
	}
	if cp.AttributeIDs != nil {
		clone.AttributeIDs = make([]string, len(cp.AttributeIDs))
		copy(clone.AttributeIDs, cp.AttributeIDs)
	}
	if cp.ActivationInterval != nil {
		clone.ActivationInterval = cp.ActivationInterval.Clone()
	}
	return clone
}

// CacheClone returns a clone of ChargerProfile used by ltcache CacheCloner
func (cp *ChargerProfile) CacheClone() any {
	return cp.Clone()
}

// ChargerProfileWithAPIOpts is used in replicatorV1 for dispatcher
type ChargerProfileWithAPIOpts struct {
	*ChargerProfile
	APIOpts map[string]any
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
