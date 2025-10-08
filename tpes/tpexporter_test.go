/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package tpes

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestNewTPExporter(t *testing.T) {
	// Attributes
	rcv, err := newTPExporter(utils.MetaAttributes, nil)
	if err != nil {
		t.Error(err)
	}
	expAttr := &TPAttributes{
		dm: nil,
	}
	if !reflect.DeepEqual(rcv, expAttr) {
		t.Errorf("Expected %v\n but received %v", expAttr, rcv)
	}

	//Resources
	rcv, err = newTPExporter(utils.MetaResources, nil)
	if err != nil {
		t.Error(err)
	}
	expRsc := &TPResources{
		dm: nil,
	}
	if !reflect.DeepEqual(rcv, expRsc) {
		t.Errorf("Expected %v\n but received %v", expRsc, rcv)
	}

	//Filters
	rcv, err = newTPExporter(utils.MetaFilters, nil)
	if err != nil {
		t.Error(err)
	}
	expFltr := &TPFilters{
		dm: nil,
	}
	if !reflect.DeepEqual(rcv, expFltr) {
		t.Errorf("Expected %v\n but received %v", expFltr, rcv)
	}

	//Rates
	rcv, err = newTPExporter(utils.MetaRates, nil)
	if err != nil {
		t.Error(err)
	}
	expRt := &TPRates{
		dm: nil,
	}
	if !reflect.DeepEqual(rcv, expRt) {
		t.Errorf("Expected %v\n but received %v", expRt, rcv)
	}

	//Chargers
	rcv, err = newTPExporter(utils.MetaChargers, nil)
	if err != nil {
		t.Error(err)
	}
	expChg := &TPChargers{
		dm: nil,
	}
	if !reflect.DeepEqual(rcv, expChg) {
		t.Errorf("Expected %v\n but received %v", expChg, rcv)
	}

	//Routes
	rcv, err = newTPExporter(utils.MetaRoutes, nil)
	if err != nil {
		t.Error(err)
	}
	expRts := &TPRoutes{
		dm: nil,
	}
	if !reflect.DeepEqual(rcv, expRts) {
		t.Errorf("Expected %v\n but received %v", expRts, rcv)
	}

	//Accounts
	rcv, err = newTPExporter(utils.MetaAccounts, nil)
	if err != nil {
		t.Error(err)
	}
	expAcc := &TPAccounts{
		dm: nil,
	}
	if !reflect.DeepEqual(rcv, expAcc) {
		t.Errorf("Expected %v\n but received %v", expAcc, rcv)
	}

	//Stats
	rcv, err = newTPExporter(utils.MetaStats, nil)
	if err != nil {
		t.Error(err)
	}
	expSt := &TPStats{
		dm: nil,
	}
	if !reflect.DeepEqual(rcv, expSt) {
		t.Errorf("Expected %v\n but received %v", expSt, rcv)
	}

	//Actions
	rcv, err = newTPExporter(utils.MetaActions, nil)
	if err != nil {
		t.Error(err)
	}
	expAct := &TPActions{
		dm: nil,
	}
	if !reflect.DeepEqual(rcv, expAct) {
		t.Errorf("Expected %v\n but received %v", expAct, rcv)
	}

	//Thresholds
	rcv, err = newTPExporter(utils.MetaThresholds, nil)
	if err != nil {
		t.Error(err)
	}
	expThd := &TPThresholds{
		dm: nil,
	}
	if !reflect.DeepEqual(rcv, expThd) {
		t.Errorf("Expected %v\n but received %v", expThd, rcv)
	}

	//Unsupported type
	_, err = newTPExporter("does not exist", nil)
	errExpect := "UNSUPPORTED_TPEXPORTER_TYPE:does not exist"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}
