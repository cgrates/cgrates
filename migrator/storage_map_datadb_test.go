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
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestInternalMigratorSetV1Account(t *testing.T) {
	iDBMig := &internalMigrator{}
	account := &v1Account{
		Id: "id",
		BalanceMap: map[string]v1BalanceChain{
			"chain1": {},
		},
		UnitCounters: []*v1UnitsCounter{
			{},
		},
		ActionTriggers: v1ActionTriggers{},
		AllowNegative:  true,
		Disabled:       false,
	}
	err := iDBMig.setV1Account(account)
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error to be ErrNotImplemented, but got %v", err)
	}
}

func TestInternalMigratorSetAndRemV2Account(t *testing.T) {
	iDBMig := &internalMigrator{}
	Account := &v2Account{
		ID:             "id",
		BalanceMap:     make(map[string]engine.Balances),
		UnitCounters:   engine.UnitCounters{},
		ActionTriggers: engine.ActionTriggers{},
		AllowNegative:  true,
		Disabled:       false,
	}
	ID := "id"
	err := iDBMig.setV2Account(Account)
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error for setV2Account to be ErrNotImplemented, but got %v", err)
	}
	err = iDBMig.remV2Account(ID)
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error for remV2Account to be ErrNotImplemented, but got %v", err)
	}
}

func TestInternalMigratorMethods(t *testing.T) {
	iDBMig := &internalMigrator{}
	v1aps, err := iDBMig.getV1ActionPlans()
	if v1aps != nil {
		t.Fatalf("Expected v1ActionPlans to be nil, but got %v", v1aps)
	}
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error for getV1ActionPlans to be ErrNotImplemented, but got %v", err)
	}
	v1acs, err := iDBMig.getV1Actions()
	if v1acs != nil {
		t.Fatalf("Expected v1Actions to be nil, but got %v", v1acs)
	}
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error for getV1Actions to be ErrNotImplemented, but got %v", err)
	}

}

func TestInternalMigratorGetV1RouteProfile(t *testing.T) {
	iDBMig := &internalMigrator{}
	v1chrPrf, err := iDBMig.getV1RouteProfile()
	if v1chrPrf != nil {
		t.Fatalf("Expected v1chrPrf to be nil, but got %v", v1chrPrf)
	}
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error to be ErrNotImplemented, but got %v", err)
	}
}

func TestInternalMigratorDataManager(t *testing.T) {
	dataManager := &engine.DataManager{}
	iDBMig := &internalMigrator{
		dm: dataManager,
	}
	returnedDM := iDBMig.DataManager()
	if returnedDM != dataManager {
		t.Fatalf("Expected DataManager to be %v, but got %v", dataManager, returnedDM)
	}
}

func TestInternalMigratorGetv1Account(t *testing.T) {
	iDBMig := &internalMigrator{}
	v1Acnt, err := iDBMig.getv1Account()
	if v1Acnt != nil {
		t.Fatalf("Expected v1Acnt to be nil, but got %v", v1Acnt)
	}
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error to be ErrNotImplemented, but got %v", err)
	}
}

func TestInternalMigratorRemV1Account(t *testing.T) {
	iDBMig := &internalMigrator{}
	testID := "id"
	err := iDBMig.remV1Account(testID)
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error to be ErrNotImplemented, but got %v", err)
	}
}

func TestInternalMigratorGetv2Account(t *testing.T) {
	iDBMig := &internalMigrator{}
	v2Acnt, err := iDBMig.getv2Account()
	if v2Acnt != nil {
		t.Fatalf("Expected v2Account to be nil, but got %v", v2Acnt)
	}
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error to be ErrNotImplemented, but got %v", err)
	}
}

func TestInternalMigratorProfileMethods(t *testing.T) {
	iDBMig := &internalMigrator{}
	v1chrPrf, err := iDBMig.getV1ChargerProfile()
	if v1chrPrf != nil {
		t.Fatalf("Expected ChargerProfile to be nil, but got %v", v1chrPrf)
	}
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error for getV1ChargerProfile to be ErrNotImplemented, but got %v", err)
	}

}

func TestInternalMigratorGetV1DispatcherProfile(t *testing.T) {
	iDBMig := &internalMigrator{}
	v1chrPrf, err := iDBMig.getV1DispatcherProfile()
	if v1chrPrf != nil {
		t.Fatalf("Expected DispatcherProfile to be nil, but got %v", v1chrPrf)
	}
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error to be ErrNotImplemented, but got %v", err)
	}
}

func TestInternalMigratorRemSupplier(t *testing.T) {
	iDBMig := &internalMigrator{}
	tenant := "cgrates.org"
	id := "ID"
	err := iDBMig.remSupplier(tenant, id)
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error to be ErrNotImplemented, but got %v", err)
	}
}

func TestInternalMigratorSetSupplier(t *testing.T) {
	iDBMig := &internalMigrator{}
	supplierProfile := &SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "1",
		FilterIDs: []string{"id1", "id2", "id3"},
		Sorting:   "sorting",
		Weight:    10,
	}
	err := iDBMig.setSupplier(supplierProfile)
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error to be ErrNotImplemented, but got %v", err)
	}
}

func TestInternalMigratorGetSupplier(t *testing.T) {
	iDBMig := &internalMigrator{}
	supplier, err := iDBMig.getSupplier()
	if supplier != nil {
		t.Fatalf("Expected supplier to be nil, but got %v", supplier)
	}
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error to be ErrNotImplemented, but got %v", err)
	}
}

func TestInternalMigratorRemV1Filter(t *testing.T) {
	iDBMig := &internalMigrator{}
	tenant := "cgrates.org"
	id := "ID"
	err := iDBMig.remV1Filter(tenant, id)
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error to be ErrNotImplemented, but got %v", err)
	}
}

func TestInternalMigratorSetV1Filter(t *testing.T) {
	iDBMig := &internalMigrator{}
	sampleFilter := &v1Filter{
		Tenant: "cgrates.org",
		ID:     "ID",
	}
	err := iDBMig.setV1Filter(sampleFilter)
	if err != utils.ErrNotImplemented {
		t.Fatalf("Expected error to be ErrNotImplemented, but got %v", err)
	}
}
