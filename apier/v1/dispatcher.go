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

package v1

import (
	"github.com/cgrates/cgrates/dispatcher"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewDispatcherThresholdSv1(dps *dispatcher.DispatcherService) *DispatcherThresholdSv1 {
	return &DispatcherThresholdSv1{dS: dps}
}

// Exports RPC from RLs
type DispatcherThresholdSv1 struct {
	dS *dispatcher.DispatcherService
}

// Ping implements ThresholdSv1Ping
func (dT *DispatcherThresholdSv1) Ping(ign string, reply *string) error {
	return dT.dS.ThresholdSv1Ping(ign, reply)
}

// GetThresholdIDs implements ThresholdSv1GetThresholdIDs
func (dT *DispatcherThresholdSv1) GetThresholdIDs(tenant string,
	tIDs *[]string) error {
	return dT.dS.ThresholdSv1GetThresholdIDs(tenant, tIDs)
}

// GetThreshold implements ThresholdSv1GetThreshold
func (dT *DispatcherThresholdSv1) GetThreshold(tntID *utils.TenantID,
	t *engine.Threshold) error {
	return dT.dS.ThresholdSv1GetThreshold(tntID, t)
}

// ProcessEvent implements ThresholdSv1ProcessEvent
func (dT *DispatcherThresholdSv1) ProcessEvent(args *engine.ArgsProcessEvent,
	tIDs *[]string) error {
	return dT.dS.ThresholdSv1ProcessEvent(args, tIDs)
}

func NewDispatcherStatSv1(dps *dispatcher.DispatcherService) *DispatcherStatSv1 {
	return &DispatcherStatSv1{dS: dps}
}

// Exports RPC from RLs
type DispatcherStatSv1 struct {
	dS *dispatcher.DispatcherService
}

// Ping implements StatSv1Ping
func (dSts *DispatcherStatSv1) Ping(ign string, reply *string) error {
	return dSts.dS.StatSv1Ping(ign, reply)
}

// GetStatQueuesForEvent implements StatSv1GetStatQueuesForEvent
func (dSts *DispatcherStatSv1) GetStatQueuesForEvent(ev *utils.CGREvent, reply *[]string) error {
	return dSts.dS.StatSv1GetStatQueuesForEvent(ev, reply)
}

// GetQueueStringMetrics implements StatSv1GetQueueStringMetrics
func (dSts *DispatcherStatSv1) GetQueueStringMetrics(args *utils.TenantID,
	reply *map[string]string) error {
	return dSts.dS.StatSv1GetQueueStringMetrics(args, reply)
}

// GetQueueStringMetrics implements StatSv1ProcessEvent
func (dSts *DispatcherStatSv1) ProcessEvent(ev *utils.CGREvent, reply *[]string) error {
	return dSts.dS.StatSv1ProcessEvent(ev, reply)
}

func NewDispatcherResourceSv1(dps *dispatcher.DispatcherService) *DispatcherResourceSv1 {
	return &DispatcherResourceSv1{dRs: dps}
}

// Exports RPC from RLs
type DispatcherResourceSv1 struct {
	dRs *dispatcher.DispatcherService
}

// Ping implements ResourceSv1Ping
func (dRs *DispatcherResourceSv1) Ping(ign string, reply *string) error {
	return dRs.dRs.ResourceSv1Ping(ign, reply)
}

// GetResourcesForEvent implements ResourceSv1GetResourcesForEvent
func (dRs *DispatcherResourceSv1) GetResourcesForEvent(args utils.ArgRSv1ResourceUsage,
	reply *engine.Resources) error {
	return dRs.dRs.ResourceSv1GetResourcesForEvent(args, reply)
}

func NewDispatcherSupplierSv1(dps *dispatcher.DispatcherService) *DispatcherSupplierSv1 {
	return &DispatcherSupplierSv1{dSup: dps}
}

// Exports RPC from RLs
type DispatcherSupplierSv1 struct {
	dSup *dispatcher.DispatcherService
}

// Ping implements SupplierSv1Ping
func (dSup *DispatcherSupplierSv1) Ping(ign string, reply *string) error {
	return dSup.dSup.SupplierSv1Ping(ign, reply)
}

// GetSuppliers implements SupplierSv1GetSuppliers
func (dSup *DispatcherSupplierSv1) GetSuppliers(args *engine.ArgsGetSuppliers,
	reply *engine.SortedSuppliers) error {
	return dSup.dSup.SupplierSv1GetSuppliers(args, reply)
}

func NewDispatcherAttributeSv1(dps *dispatcher.DispatcherService) *DispatcherAttributeSv1 {
	return &DispatcherAttributeSv1{dA: dps}
}

// Exports RPC from RLs
type DispatcherAttributeSv1 struct {
	dA *dispatcher.DispatcherService
}

// Ping implements SupplierSv1Ping
func (dA *DispatcherAttributeSv1) Ping(ign string, reply *string) error {
	return dA.dA.AttributeSv1Ping(ign, reply)
}

// GetAttributeForEvent implements AttributeSv1GetAttributeForEvent
func (dA *DispatcherAttributeSv1) GetAttributeForEvent(ev *utils.CGREvent,
	reply *engine.AttributeProfile) error {
	return dA.dA.AttributeSv1GetAttributeForEvent(ev, reply)
}

// ProcessEvent implements AttributeSv1ProcessEvent
func (dA *DispatcherAttributeSv1) ProcessEvent(ev *utils.CGREvent,
	reply *engine.AttrSProcessEventReply) error {
	return dA.dA.AttributeSv1ProcessEvent(ev, reply)
}
