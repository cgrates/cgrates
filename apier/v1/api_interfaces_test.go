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

func TestRouteSv1Interface(t *testing.T) {
	_ = RouteSv1Interface(NewDispatcherRouteSv1(nil))
	_ = RouteSv1Interface(NewRouteSv1(nil))
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
	_ = SessionSv1Interface(NewSessionSv1(nil, nil))
}

func TestResponderInterface(t *testing.T) {
	_ = ResponderInterface(NewDispatcherResponder(nil))
	_ = ResponderInterface(&engine.Responder{})
}

func TestRateProfileInterface(t *testing.T) {
	_ = RateProfileSv1Interface(NewDispatcherRateSv1(nil))
	_ = RateProfileSv1Interface(NewRateSv1(nil))
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
	_ = SchedulerSv1Interface(NewSchedulerSv1(nil, nil))
}

func TestCDRsV1Interface(t *testing.T) {
	_ = CDRsV1Interface(NewDispatcherSCDRsV1(nil))
	_ = CDRsV1Interface(NewCDRsV1(nil))
}

func TestServiceManagerV1Interface(t *testing.T) {
	_ = ServiceManagerV1Interface(NewDispatcherSServiceManagerV1(nil))
	_ = ServiceManagerV1Interface(NewServiceManagerV1(nil))
}

func TestRALsV1Interface(t *testing.T) {
	_ = RALsV1Interface(NewDispatcherRALsV1(nil))
	_ = RALsV1Interface(NewRALsV1())
}

func TestConfigSv1Interface(t *testing.T) {
	_ = ConfigSv1Interface(NewDispatcherConfigSv1(nil))
	_ = ConfigSv1Interface(NewConfigSv1(nil))
}

func TestCoreSv1Interface(t *testing.T) {
	_ = CoreSv1Interface(NewDispatcherCoreSv1(nil))
	_ = CoreSv1Interface(NewCoreSv1(nil))
}

func TestReplicatorSv1Interface(t *testing.T) {
	_ = ReplicatorSv1Interface(NewDispatcherReplicatorSv1(nil))
	_ = ReplicatorSv1Interface(NewReplicatorSv1(nil))
}

func TestRateSv1Interface(t *testing.T) {
	_ = RateSv1Interface(NewDispatcherRateSv1(nil))
	_ = RateSv1Interface(NewRateSv1(nil))
}

func TestAccountSv1Interface(t *testing.T) {
	_ = AccountSv1Interface(NewDispatcherAccountSv1(nil))
	_ = AccountSv1Interface(NewAccountSv1(nil))
}

func TestActionSv1Interface(t *testing.T) {
	_ = ActionSv1Interface(NewDispatcherActionSv1(nil))
	_ = ActionSv1Interface(NewActionSv1(nil))
}
