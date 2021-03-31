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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	SupplierProfilePrefix = "spp_"
	ColSpp                = "supplier_profiles"
)

type Supplier struct {
	ID                 string // SupplierID
	FilterIDs          []string
	AccountIDs         []string
	RatingPlanIDs      []string // used when computing price
	ResourceIDs        []string // queried in some strategies
	StatIDs            []string // queried in some strategies
	Weight             float64
	Blocker            bool // do not process further supplier after this one
	SupplierParameters string
}

type SupplierProfile struct {
	Tenant             string
	ID                 string // LCR Profile ID
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	Sorting            string                    // Sorting strategy
	SortingParameters  []string
	Suppliers          []*Supplier
	Weight             float64
}

func (m *Migrator) removeSupplier() (err error) {
	for {
		var spp *SupplierProfile
		spp, err = m.dmIN.getSupplier()
		if err == utils.ErrNoMoreData {
			break
		}
		if err != nil {
			return
		}
		if err = m.dmIN.remSupplier(spp.Tenant, spp.ID); err != nil {
			return
		}
	}
	return

}

func (m *Migrator) migrateFromSupplierToRoute() (err error) {
	for {
		var spp *SupplierProfile
		spp, err = m.dmIN.getSupplier()
		if err == utils.ErrNoMoreData {
			break
		}
		if err != nil {
			return err
		}
		if spp == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetRouteProfile(convertSupplierToRoute(spp), true); err != nil {
			return err
		}
		m.stats[utils.Routes]++
	}
	if m.dryRun {
		return
	}
	if err = m.removeSupplier(); err != nil && err != utils.ErrNoMoreData {
		return
	}
	// All done, update version with current one
	vrs := engine.Versions{utils.Routes: 1}
	if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating RouteProfiles version into dataDB", err.Error()))
	}
	return
}

func (m *Migrator) migrateCurrentRouteProfile() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.RouteProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.RouteProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("invalid key <%s> when migrating route profiles", id)
		}
		rPrf, err := m.dmIN.DataManager().GetRouteProfile(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if rPrf == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetRouteProfile(rPrf, true); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveRouteProfile(tntID[0], tntID[1], utils.NonTransactional, true); err != nil {
			return err
		}
		m.stats[utils.Routes]++
	}
	return
}

func convertSupplierToRoute(spp *SupplierProfile) (route *engine.RouteProfile) {
	route = &engine.RouteProfile{
		Tenant:             spp.Tenant,
		ID:                 spp.ID,
		FilterIDs:          spp.FilterIDs,
		ActivationInterval: spp.ActivationInterval,
		Sorting:            spp.Sorting,
		SortingParameters:  spp.SortingParameters,
		Weight:             spp.Weight,
	}
	route.Routes = make([]*engine.Route, len(spp.Suppliers))
	for i, supl := range spp.Suppliers {
		route.Routes[i] = &engine.Route{
			ID:              supl.ID,
			FilterIDs:       supl.FilterIDs,
			AccountIDs:      supl.AccountIDs,
			RatingPlanIDs:   supl.RatingPlanIDs,
			ResourceIDs:     supl.ResourceIDs,
			StatIDs:         supl.StatIDs,
			Weight:          supl.Weight,
			Blocker:         supl.Blocker,
			RouteParameters: supl.SupplierParameters,
		}
	}
	return
}
