/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestResourceProfileTenantID(t *testing.T) {
	testStruct := ResourceProfile{
		Tenant: "test_tenant",
		ID:     "test_id",
	}
	result := testStruct.TenantID()
	expected := utils.ConcatenatedKey(testStruct.Tenant, testStruct.ID)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, result)
	}
}

func TestResourceUsageTenantID(t *testing.T) {
	testStruct := ResourceUsage{
		Tenant: "test_tenant",
		ID:     "test_id",
	}
	result := testStruct.TenantID()
	expected := utils.ConcatenatedKey(testStruct.Tenant, testStruct.ID)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, result)
	}
}

func TestResourceUsageisActive(t *testing.T) {
	testStruct := ResourceUsage{
		Tenant:     "test_tenant",
		ID:         "test_id",
		ExpiryTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC),
	}
	result := testStruct.isActive(time.Date(2014, 1, 13, 0, 0, 0, 0, time.UTC))
	if !reflect.DeepEqual(true, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", true, result)
	}
}

func TestResourceUsageisActiveFalse(t *testing.T) {
	testStruct := ResourceUsage{
		Tenant:     "test_tenant",
		ID:         "test_id",
		ExpiryTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC),
	}
	result := testStruct.isActive(time.Date(2014, 1, 15, 0, 0, 0, 0, time.UTC))
	if !reflect.DeepEqual(false, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", false, result)
	}
}

func TestResourceUsageClone(t *testing.T) {
	testStruct := &ResourceUsage{
		Tenant:     "test_tenant",
		ID:         "test_id",
		ExpiryTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC),
	}
	result := testStruct.Clone()
	if !reflect.DeepEqual(testStruct, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct, result)
	}
	testStruct.Tenant = "test_tenant2"
	testStruct.ID = "test_id2"
	testStruct.ExpiryTime = time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC)
	if reflect.DeepEqual(testStruct.Tenant, result.Tenant) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct.Tenant, result.Tenant)
	}
	if reflect.DeepEqual(testStruct.ID, result.ID) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct.ID, result.ID)
	}
	if reflect.DeepEqual(testStruct.ExpiryTime, result.ExpiryTime) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct.ExpiryTime, result.ExpiryTime)
	}
}

func TestResourceTenantID(t *testing.T) {
	testStruct := Resource{
		Tenant: "test_tenant",
	}
	result := testStruct.TenantID()
	if reflect.DeepEqual(testStruct.Tenant, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct.Tenant, result)
	}
}
