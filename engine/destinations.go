// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"slices"
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

// Clone returns a clone of Destination
func (d *Destination) Clone() *Destination {
	if d == nil {
		return nil
	}
	result := &Destination{
		Id: d.Id,
	}
	if d.Prefixes != nil {
		result.Prefixes = make([]string, len(d.Prefixes))
		copy(result.Prefixes, d.Prefixes)
	}
	return result
}

// CacheClone returns a clone of Destination used by ltcache CacheCloner
func (d *Destination) CacheClone() any {
	return d.Clone()
}

type DestinationWithAPIOpts struct {
	*Destination
	Tenant  string
	APIOpts map[string]any
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
	if cached, err := dm.GetReverseDestination(prefix, true, true, utils.NonTransactional); err == nil {
		return slices.Contains(cached, destId)
	}
	return false
}
