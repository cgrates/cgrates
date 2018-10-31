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
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
)

func NewDispatcherThresholdSv1(dps *dispatchers.DispatcherService) *DispatcherThresholdSv1 {
	return &DispatcherThresholdSv1{dS: dps}
}

// Exports RPC from RLs
type DispatcherThresholdSv1 struct {
	dS *dispatchers.DispatcherService
}

// Ping implements ThresholdSv1Ping
func (dT *DispatcherThresholdSv1) Ping(ign string, reply *string) error {
	return dT.dS.ThresholdSv1Ping(ign, reply)
}

// GetThresholdsForEvent implements ThresholdSv1GetThresholdsForEvent
func (dT *DispatcherThresholdSv1) GetThresholdsForEvent(tntID *dispatchers.ArgsProcessEventWithApiKey,
	t *engine.Thresholds) error {
	return dT.dS.ThresholdSv1GetThresholdsForEvent(tntID, t)
}

// ProcessEvent implements ThresholdSv1ProcessEvent
func (dT *DispatcherThresholdSv1) ProcessEvent(args *dispatchers.ArgsProcessEventWithApiKey,
	tIDs *[]string) error {
	return dT.dS.ThresholdSv1ProcessEvent(args, tIDs)
}

func NewDispatcherStatSv1(dps *dispatchers.DispatcherService) *DispatcherStatSv1 {
	return &DispatcherStatSv1{dS: dps}
}

// Exports RPC from RLs
type DispatcherStatSv1 struct {
	dS *dispatchers.DispatcherService
}

// Ping implements StatSv1Ping
func (dSts *DispatcherStatSv1) Ping(ign string, reply *string) error {
	return dSts.dS.StatSv1Ping(ign, reply)
}

// GetStatQueuesForEvent implements StatSv1GetStatQueuesForEvent
func (dSts *DispatcherStatSv1) GetStatQueuesForEvent(args *dispatchers.ArgsStatProcessEventWithApiKey, reply *[]string) error {
	return dSts.dS.StatSv1GetStatQueuesForEvent(args, reply)
}

// GetQueueStringMetrics implements StatSv1GetQueueStringMetrics
func (dSts *DispatcherStatSv1) GetQueueStringMetrics(args *dispatchers.TntIDWithApiKey,
	reply *map[string]string) error {
	return dSts.dS.StatSv1GetQueueStringMetrics(args, reply)
}

// GetQueueStringMetrics implements StatSv1ProcessEvent
func (dSts *DispatcherStatSv1) ProcessEvent(args *dispatchers.ArgsStatProcessEventWithApiKey, reply *[]string) error {
	return dSts.dS.StatSv1ProcessEvent(args, reply)
}

func NewDispatcherResourceSv1(dps *dispatchers.DispatcherService) *DispatcherResourceSv1 {
	return &DispatcherResourceSv1{dRs: dps}
}

// Exports RPC from RLs
type DispatcherResourceSv1 struct {
	dRs *dispatchers.DispatcherService
}

// Ping implements ResourceSv1Ping
func (dRs *DispatcherResourceSv1) Ping(ign string, reply *string) error {
	return dRs.dRs.ResourceSv1Ping(ign, reply)
}

// GetResourcesForEvent implements ResourceSv1GetResourcesForEvent
func (dRs *DispatcherResourceSv1) GetResourcesForEvent(args *dispatchers.ArgsV1ResUsageWithApiKey,
	reply *engine.Resources) error {
	return dRs.dRs.ResourceSv1GetResourcesForEvent(args, reply)
}

func NewDispatcherSupplierSv1(dps *dispatchers.DispatcherService) *DispatcherSupplierSv1 {
	return &DispatcherSupplierSv1{dSup: dps}
}

// Exports RPC from RLs
type DispatcherSupplierSv1 struct {
	dSup *dispatchers.DispatcherService
}

// Ping implements SupplierSv1Ping
func (dSup *DispatcherSupplierSv1) Ping(ign string, reply *string) error {
	return dSup.dSup.SupplierSv1Ping(ign, reply)
}

// GetSuppliers implements SupplierSv1GetSuppliers
func (dSup *DispatcherSupplierSv1) GetSuppliers(args *dispatchers.ArgsGetSuppliersWithApiKey,
	reply *engine.SortedSuppliers) error {
	return dSup.dSup.SupplierSv1GetSuppliers(args, reply)
}

func NewDispatcherAttributeSv1(dps *dispatchers.DispatcherService) *DispatcherAttributeSv1 {
	return &DispatcherAttributeSv1{dA: dps}
}

// Exports RPC from RLs
type DispatcherAttributeSv1 struct {
	dA *dispatchers.DispatcherService
}

// Ping implements SupplierSv1Ping
func (dA *DispatcherAttributeSv1) Ping(ign string, reply *string) error {
	return dA.dA.AttributeSv1Ping(ign, reply)
}

// GetAttributeForEvent implements AttributeSv1GetAttributeForEvent
func (dA *DispatcherAttributeSv1) GetAttributeForEvent(args *dispatchers.ArgsAttrProcessEventWithApiKey,
	reply *engine.AttributeProfile) error {
	return dA.dA.AttributeSv1GetAttributeForEvent(args, reply)
}

// ProcessEvent implements AttributeSv1ProcessEvent
func (dA *DispatcherAttributeSv1) ProcessEvent(args *dispatchers.ArgsAttrProcessEventWithApiKey,
	reply *engine.AttrSProcessEventReply) error {
	return dA.dA.AttributeSv1ProcessEvent(args, reply)
}

func NewDispatcherSessionSv1(dps *dispatchers.DispatcherService) *DispatcherSessionSv1 {
	return &DispatcherSessionSv1{dS: dps}
}

// Exports RPC from RLs
type DispatcherSessionSv1 struct {
	dS *dispatchers.DispatcherService
}

// Ping implements SessionSv1Ping
func (dS *DispatcherSessionSv1) Ping(ign string, reply *string) error {
	return dS.dS.SessionSv1Ping(ign, reply)
}

// AuthorizeEventWithDigest implements SessionSv1AuthorizeEventWithDigest
func (dS *DispatcherSessionSv1) AuthorizeEventWithDigest(args *dispatchers.AuthorizeArgsWithApiKey,
	reply *sessions.V1AuthorizeReplyWithDigest) error {
	return dS.dS.SessionSv1AuthorizeEventWithDigest(args, reply)
}

// InitiateSessionWithDigest implements SessionSv1InitiateSessionWithDigest
func (dS *DispatcherSessionSv1) InitiateSessionWithDigest(args *dispatchers.InitArgsWithApiKey,
	reply *sessions.V1InitSessionReply) (err error) {
	return dS.dS.SessionSv1InitiateSessionWithDigest(args, reply)
}

// ProcessCDR implements SessionSv1ProcessCDR
func (dS *DispatcherSessionSv1) ProcessCDR(args *dispatchers.CGREvWithApiKey,
	reply *string) (err error) {
	return dS.dS.SessionSv1ProcessCDR(args, reply)
}

// ProcessEvent implements SessionSv1ProcessEvent
func (dS *DispatcherSessionSv1) ProcessEvent(args *dispatchers.ProcessEventWithApiKey,
	reply *sessions.V1ProcessEventReply) (err error) {
	return dS.dS.SessionSv1ProcessEvent(args, reply)
}

// TerminateSession implements SessionSv1TerminateSession
func (dS *DispatcherSessionSv1) TerminateSession(args *dispatchers.TerminateSessionWithApiKey,
	reply *string) (err error) {
	return dS.dS.SessionSv1TerminateSession(args, reply)
}

// UpdateSession implements SessionSv1UpdateSession
func (dS *DispatcherSessionSv1) UpdateSession(args *dispatchers.UpdateSessionWithApiKey,
	reply *sessions.V1UpdateSessionReply) (err error) {
	return dS.dS.SessionSv1UpdateSession(args, reply)
}

func NewDispatcherChargerSv1(dps *dispatchers.DispatcherService) *DispatcherChargerSv1 {
	return &DispatcherChargerSv1{dC: dps}
}

// Exports RPC from RLs
type DispatcherChargerSv1 struct {
	dC *dispatchers.DispatcherService
}

// Ping implements ChargerSv1Ping
func (dC *DispatcherChargerSv1) Ping(ign string, reply *string) error {
	return dC.dC.ChargerSv1Ping(ign, reply)
}

// GetChargersForEvent implements ChargerSv1GetChargersForEvent
func (dC *DispatcherChargerSv1) GetChargersForEvent(args *dispatchers.CGREvWithApiKey,
	reply *engine.ChargerProfiles) (err error) {
	return dC.dC.ChargerSv1GetChargersForEvent(args, reply)
}

// ProcessEvent implements ChargerSv1ProcessEvent
func (dC *DispatcherChargerSv1) ProcessEvent(args *dispatchers.CGREvWithApiKey,
	reply *[]*engine.AttrSProcessEventReply) (err error) {
	return dC.dC.ChargerSv1ProcessEvent(args, reply)
}
