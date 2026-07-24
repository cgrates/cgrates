// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"strings"

	"github.com/cgrates/cgrates/utils"
)

func NewDestinationFromTPDestination(tpDst *utils.TPDestination) *Destination {
	return &Destination{Id: tpDst.ID, Prefixes: tpDst.Prefixes}

}

/*
Structure that gathers multiple destination prefixes under a common id.
*/
type Destination struct {
	Id       string
	Prefixes []string
}

// returns prefix precision
func (d *Destination) containsPrefix(prefix string) int {
	if d == nil {
		return 0
	}
	for _, p := range d.Prefixes {
		if strings.Index(prefix, p) == 0 {
			return len(p)
		}
	}
	return 0
}

func (d *Destination) String() (result string) {
	result = d.Id + ": "
	for _, k := range d.Prefixes {
		result += k + ", "
	}
	result = strings.TrimRight(result, ", ")
	return result
}

func (d *Destination) AddPrefix(pfx string) {
	d.Prefixes = append(d.Prefixes, pfx)
}

// Reverse search in cache to see if prefix belongs to destination id
func CachedDestHasPrefix(destId, prefix string) bool {
	if cached, err := dm.GetReverseDestination(prefix, false, utils.NonTransactional); err == nil {
		return utils.IsSliceMember(cached, destId)
	}
	return false
}

/*func CleanStalePrefixes(destIds []string) {
	utils.Logger.Info("Cleaning stale dest prefixes: " + utils.ToJSON(destIds))
	prefixMap := cache.GetAllEntries(utils.REVERSE_DESTINATION_PREFIX)
	for prefix, idIDs := range prefixMap {
		dIDs := idIDs.(map[string]struct{})
		changed := false
		for _, searchedDID := range destIds {
			if _, found := dIDs[searchedDID]; found {
				if len(dIDs) == 1 {
					// remove de prefix from cache
					cache.RemKey(utils.REVERSE_DESTINATION_PREFIX + prefix)
				} else {
					// delete the destination from list and put the new list in chache
					delete(dIDs, searchedDID)
					changed = true
				}
			}
		}
		if changed {
			cache.Set(utils.REVERSE_DESTINATION_PREFIX+prefix, dIDs)
		}
	}
}
*/
