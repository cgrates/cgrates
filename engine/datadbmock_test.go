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

package engine

import (
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestSetResourceProfileDrv(t *testing.T) {
	tests := []struct {
		name        string
		setFunction func(*ResourceProfile) error
		expectedErr error
	}{
		{
			name:        "No Function Set",
			setFunction: nil,
			expectedErr: utils.ErrNotImplemented,
		},
		{
			name: "Function Set",
			setFunction: func(rp *ResourceProfile) error {
				return utils.ErrNotImplemented
			},
			expectedErr: utils.ErrNotImplemented,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbM := &DataDBMock{
				SetResourceProfileDrvF: tt.setFunction,
			}
			rp := &ResourceProfile{}
			err := dbM.SetResourceProfileDrv(rp)
			if err != tt.expectedErr {
				t.Errorf("expected error %v, but got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestRemoveSessionsBackupDrv(t *testing.T) {
	dbMock := &DataDBMock{}
	err := dbMock.RemoveSessionsBackupDrv("node1", "tenant1", "cgrid1")
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetSessionsBackupDrv(t *testing.T) {
	dbM := &DataDBMock{}
	sessions, err := dbM.GetSessionsBackupDrv("nodeID", "tenant")
	if sessions != nil {
		t.Errorf("expected nil, but got %v", sessions)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestSetBackupSessionsDrv(t *testing.T) {
	dbM := &DataDBMock{}
	nodeID := "nodeID"
	tnt := "tenant"
	storedSessions := []*StoredSession{}
	err := dbM.SetBackupSessionsDrv(nodeID, tnt, storedSessions)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestRemoveRatingProfileDrv(t *testing.T) {
	dbM := &DataDBMock{}
	err := dbM.RemoveRatingProfileDrv("ID")
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestSetVersions(t *testing.T) {
	dbM := &DataDBMock{}
	sampleVersions := Versions{}
	err := dbM.SetVersions(sampleVersions, true)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestRemoveDispatcherHostDrv(t *testing.T) {
	dbM := &DataDBMock{}
	err := dbM.RemoveDispatcherHostDrv("NodeID", "Host")
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestSetDispatcherHostDrv(t *testing.T) {
	dbM := &DataDBMock{}
	host := &DispatcherHost{}
	err := dbM.SetDispatcherHostDrv(host)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetDispatcherHostDrv(t *testing.T) {
	dbM := &DataDBMock{}
	result, err := dbM.GetDispatcherHostDrv("arg1", "arg2")
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
	if result != nil {
		t.Errorf("expected result to be nil, but got %v", result)
	}
}

func TestRemoveLoadIDsDrv(t *testing.T) {
	dbM := &DataDBMock{}
	err := dbM.RemoveLoadIDsDrv()
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetItemLoadIDsDrv(t *testing.T) {
	dbM := &DataDBMock{}
	itemIDPrefix := "Prefix"
	loadIDs, err := dbM.GetItemLoadIDsDrv(itemIDPrefix)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
	if loadIDs != nil {
		t.Errorf("expected loadIDs to be nil, but got %v", loadIDs)
	}
}

func TestRemoveDispatcherProfileDrv(t *testing.T) {
	dbM := &DataDBMock{}
	arg1 := "Arg1"
	arg2 := "Arg2"
	err := dbM.RemoveDispatcherProfileDrv(arg1, arg2)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestSetDispatcherProfileDrv(t *testing.T) {
	dbM := &DataDBMock{}
	profile := &DispatcherProfile{}
	err := dbM.SetDispatcherProfileDrv(profile)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetDispatcherProfileDrv(t *testing.T) {
	dbM := &DataDBMock{}
	param1 := "param1"
	param2 := "param2"
	profile, err := dbM.GetDispatcherProfileDrv(param1, param2)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
	if profile != nil {
		t.Errorf("expected profile to be nil, but got %v", profile)
	}
}

func TestRemoveChargerProfileDrv(t *testing.T) {
	dbM := &DataDBMock{}
	param1 := "param1"
	param2 := "param2"
	err := dbM.RemoveChargerProfileDrv(param1, param2)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestSetChargerProfileDrv(t *testing.T) {
	dbM := &DataDBMock{}
	var profile *ChargerProfile = nil
	err := dbM.SetChargerProfileDrv(profile)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestFlush(t *testing.T) {
	dbM := &DataDBMock{}
	param := "testParam"
	err := dbM.Flush(param)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}
