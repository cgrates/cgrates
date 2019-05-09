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
	"testing"

	"github.com/cgrates/cgrates/engine"
)

func TestThresholdSv1Interface(t *testing.T) {
	_ = ThresholdSv1Interface(NewDispatcherThresholdSv1(nil))
	_ = ThresholdSv1Interface(NewThresholdSv1(nil))
}

func TestStatSv1Interface(t *testing.T) {
	_ = StatSv1Interface(NewDispatcherStatSv1(nil))
	_ = StatSv1Interface(NewStatSv1(nil))
}

func TestResourceSv1Interface(t *testing.T) {
	_ = ResourceSv1Interface(NewDispatcherResourceSv1(nil))
	_ = ResourceSv1Interface(NewResourceSv1(nil))
}

func TestSupplierSv1Interface(t *testing.T) {
	_ = SupplierSv1Interface(NewDispatcherSupplierSv1(nil))
	_ = SupplierSv1Interface(NewSupplierSv1(nil))
}

func TestAttributeSv1Interface(t *testing.T) {
	_ = AttributeSv1Interface(NewDispatcherAttributeSv1(nil))
	_ = AttributeSv1Interface(NewAttributeSv1(nil))
}

func TestChargerSv1Interface(t *testing.T) {
	_ = ChargerSv1Interface(NewDispatcherChargerSv1(nil))
	_ = ChargerSv1Interface(NewChargerSv1(nil))
}

func TestSessionSv1Interface(t *testing.T) {
	_ = SessionSv1Interface(NewDispatcherSessionSv1(nil))
	_ = SessionSv1Interface(NewSessionSv1(nil))
}

func TestResponderInterface(t *testing.T) {
	_ = ResponderInterface(NewDispatcherResponder(nil))
	_ = ResponderInterface(&engine.Responder{})
}

func TestCacheSv1Interface(t *testing.T) {
	_ = CacheSv1Interface(NewDispatcherCacheSv1(nil))
	_ = CacheSv1Interface(NewCacheSv1(nil))
}

func TestGuardianSv1Interface(t *testing.T) {
	_ = GuardianSv1Interface(NewDispatcherGuardianSv1(nil))
	_ = GuardianSv1Interface(NewGuardianSv1())
}

func TestSchedulerSv1Interface(t *testing.T) {
	_ = SchedulerSv1Interface(NewDispatcherSchedulerSv1(nil))
	_ = SchedulerSv1Interface(NewSchedulerSv1(nil))
}

func TestCDRsV1Interface(t *testing.T) {
	_ = CDRsV1Interface(NewDispatcherSCDRsV1(nil))
	_ = CDRsV1Interface(NewCDRsV1(nil))
}

func TestServiceManagerV1Interface(t *testing.T) {
	_ = ServiceManagerV1Interface(NewDispatcherSServiceManagerV1(nil))
	_ = ServiceManagerV1Interface(NewServiceManagerV1(nil))
}
