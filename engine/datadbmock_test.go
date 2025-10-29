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

package engine

import (
	"errors"
	"testing"
	"time"

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

func TestDataDBMockRemoveResourceProfileDrv(t *testing.T) {
	tests := []struct {
		name                    string
		dbM                     *DataDBMock
		tnt                     string
		id                      string
		wantErr                 error
		setupRemoveResourceFunc func(tnt, id string) error
	}{
		{
			name:    "RemoveResourceProfileDrvF not set",
			dbM:     &DataDBMock{},
			tnt:     "tenant",
			id:      "profile",
			wantErr: utils.ErrNotImplemented,
		},
		{
			name: "RemoveResourceProfileDrvF set - returns no error",
			dbM:  &DataDBMock{},
			tnt:  "tenant",
			id:   "profile",
			setupRemoveResourceFunc: func(tnt, id string) error {
				if tnt == "tenant" && id == "profile" {
					return nil
				}
				return errors.New("unexpected input")
			},
			wantErr: nil,
		},
		{
			name: "RemoveResourceProfileDrvF set - returns an error",
			dbM:  &DataDBMock{},
			tnt:  "tenant2",
			id:   "profile2",
			setupRemoveResourceFunc: func(tnt, id string) error {
				return errors.New("failed to remove resource profile")
			},
			wantErr: errors.New("failed to remove resource profile"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupRemoveResourceFunc != nil {
				tt.dbM.RemoveResourceProfileDrvF = tt.setupRemoveResourceFunc
			}
			err := tt.dbM.RemoveResourceProfileDrv(tt.tnt, tt.id)
			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("RemoveResourceProfileDrv() error = %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && tt.wantErr != nil {
				t.Errorf("RemoveResourceProfileDrv() expected error but got none")
			}
		})
	}
}

func TestGetVersions(t *testing.T) {
	dbM := &DataDBMock{}
	itm := "item"
	vrs, err := dbM.GetVersions(itm)
	if vrs != nil {
		t.Errorf("Expected versions to be nil, got %v", vrs)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be ErrNotImplemented, got %v", err)
	}
}

func TestRemoveVersions(t *testing.T) {
	dbM := &DataDBMock{}
	var vrs Versions = nil
	err := dbM.RemoveVersions(vrs)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be ErrNotImplemented, got %v", err)
	}
}

func TestSelectDatabase(t *testing.T) {
	dbM := &DataDBMock{}
	dbName := "db"
	err := dbM.SelectDatabase(dbName)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be ErrNotImplemented, got %v", err)
	}
}

func TestGetStorageType(t *testing.T) {
	dbM := &DataDBMock{}
	result := dbM.GetStorageType()
	if result != utils.EmptyString {
		t.Errorf("Expected result to be utils.EmptyString, got %v", result)
	}
}

func TestIsDBEmpty(t *testing.T) {
	dbM := &DataDBMock{}
	resp, err := dbM.IsDBEmpty()
	if resp != false {
		t.Errorf("Expected resp to be false, got %v", resp)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be ErrNotImplemented, got %v", err)
	}
}

func TestHasDataDrv(t *testing.T) {
	dbM := &DataDBMock{}
	arg1 := "arg1"
	arg2 := "arg2"
	arg3 := "arg3"
	result, err := dbM.HasDataDrv(arg1, arg2, arg3)
	if result != false {
		t.Errorf("Expected result to be false, got %v", result)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be ErrNotImplemented, got %v", err)
	}
}

func TestGetRatingPlanDrv(t *testing.T) {
	dbM := &DataDBMock{}
	arg := "plan"
	result, err := dbM.GetRatingPlanDrv(arg)
	if result != nil {
		t.Errorf("Expected result to be nil, got %v", result)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be ErrNotImplemented, got %v", err)
	}
}

func TestSetRatingPlanDrv(t *testing.T) {
	dbM := &DataDBMock{}
	var plan *RatingPlan = nil
	err := dbM.SetRatingPlanDrv(plan)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be ErrNotImplemented, got %v", err)
	}
}

func TestRemoveRatingPlanDrv(t *testing.T) {
	dbM := &DataDBMock{}
	key := "key"
	err := dbM.RemoveRatingPlanDrv(key)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be ErrNotImplemented, got %v", err)
	}
}

func TestGetRatingProfileDrv(t *testing.T) {
	dbM := &DataDBMock{}
	key := "profile"
	result, err := dbM.GetRatingProfileDrv(key)
	if result != nil {
		t.Errorf("Expected result to be nil, got %v", result)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be ErrNotImplemented, got %v", err)
	}
}

func TestSetRatingProfileDrv(t *testing.T) {
	dbM := &DataDBMock{}
	var profile *RatingProfile = nil
	err := dbM.SetRatingProfileDrv(profile)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be ErrNotImplemented, got %v", err)
	}
}

func TestGetDestinationDrv(t *testing.T) {
	dbM := &DataDBMock{}
	param1 := "param1"
	param2 := "param2"
	result, err := dbM.GetDestinationDrv(param1, param2)
	if result != nil {
		t.Errorf("Expected result to be nil, got %v", result)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be ErrNotImplemented, got %v", err)
	}
}

func TestSetDestinationDrv(t *testing.T) {
	dbM := &DataDBMock{}
	var dest *Destination = nil
	key := "key"
	err := dbM.SetDestinationDrv(dest, key)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be ErrNotImplemented, got %v", err)
	}
}

func TestRemoveDestinationDrv(t *testing.T) {
	dbM := &DataDBMock{}
	param1 := "param1"
	param2 := "param2"
	err := dbM.RemoveDestinationDrv(param1, param2)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be ErrNotImplemented, got %v", err)
	}
}

func TestRemoveReverseDestinationDrv(t *testing.T) {
	dbM := &DataDBMock{}
	param1 := "param1"
	param2 := "param2"
	param3 := "param3"
	err := dbM.RemoveReverseDestinationDrv(param1, param2, param3)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be ErrNotImplemented, got %v", err)
	}
}

func TestSetThresholdProfileDrvDefaultBehavior(t *testing.T) {
	dbM := &DataDBMock{}
	err := dbM.SetThresholdProfileDrv(&ThresholdProfile{})
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestSetThresholdProfileDrvCustomBehavior(t *testing.T) {
	expectedError := errors.New("custom error")
	dbM := &DataDBMock{
		SetThresholdProfileDrvF: func(tp *ThresholdProfile) error {
			return expectedError
		},
	}
	err := dbM.SetThresholdProfileDrv(&ThresholdProfile{})
	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}
}

func TestRemThresholdProfileDrv(t *testing.T) {
	tests := []struct {
		name        string
		mockFunc    func(tenant, id string) error
		tenant      string
		id          string
		expectError error
	}{
		{
			name:        "Default behavior",
			mockFunc:    nil,
			tenant:      "cgrates.org",
			id:          "id1",
			expectError: utils.ErrNotImplemented,
		},
		{
			name: "Custom behavior",
			mockFunc: func(tenant, id string) error {
				if tenant == "cgrates.org" && id == "id1" {
					return errors.New("custom error")
				}
				return nil
			},
			tenant:      "cgrates.org",
			id:          "id1",
			expectError: errors.New("custom error"),
		},
		{
			name: "Custom behavior with different input",
			mockFunc: func(tenant, id string) error {
				if tenant == "tenant2" && id == "id2" {
					return errors.New("another custom error")
				}
				return nil
			},
			tenant:      "tenant2",
			id:          "id2",
			expectError: errors.New("another custom error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbM := &DataDBMock{
				RemThresholdProfileDrvF: tt.mockFunc,
			}
			err := dbM.RemThresholdProfileDrv(tt.tenant, tt.id)
			if err == nil && tt.expectError != nil {
				t.Errorf("expected error %v, got nil", tt.expectError)
			} else if err != nil && tt.expectError == nil {
				t.Errorf("expected no error, got %v", err)
			} else if err != nil && err.Error() != tt.expectError.Error() {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
		})
	}
}

func TestGetThresholdDrv(t *testing.T) {
	tests := []struct {
		name        string
		mockFunc    func(tenant, id string) (*Threshold, error)
		tenant      string
		id          string
		expectValue *Threshold
		expectError error
	}{
		{
			name:        "Default behavior",
			mockFunc:    nil,
			tenant:      "cgrates.org",
			id:          "id1",
			expectValue: nil,
			expectError: utils.ErrNotImplemented,
		},
		{
			name: "Custom behavior",
			mockFunc: func(tenant, id string) (*Threshold, error) {
				if tenant == "cgrates.org" && id == "id1" {
					return &Threshold{}, nil
				}
				return nil, errors.New("custom error")
			},
			tenant:      "cgrates.org",
			id:          "id1",
			expectValue: &Threshold{},
			expectError: nil,
		},
		{
			name: "Custom behavior with error",
			mockFunc: func(tenant, id string) (*Threshold, error) {
				if tenant == "tenant2" && id == "id2" {
					return nil, errors.New("another custom error")
				}
				return nil, nil
			},
			tenant:      "tenant2",
			id:          "id2",
			expectValue: nil,
			expectError: errors.New("another custom error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbM := &DataDBMock{
				GetThresholdDrvF: tt.mockFunc,
			}
			_, err := dbM.GetThresholdDrv(tt.tenant, tt.id)
			if err == nil && tt.expectError != nil {
				t.Errorf("expected error %v, got nil", tt.expectError)
			} else if err != nil && tt.expectError == nil {
				t.Errorf("expected no error, got %v", err)
			} else if err != nil && err.Error() != tt.expectError.Error() {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
		})
	}
}

func TestGetFilterDrv(t *testing.T) {
	tests := []struct {
		name        string
		mockFunc    func(tnt, id string) (*Filter, error)
		tnt         string
		id          string
		expectValue *Filter
		expectError error
	}{
		{
			name:        "Default behavior",
			mockFunc:    nil,
			tnt:         "cgrates.org",
			id:          "id1",
			expectValue: nil,
			expectError: utils.ErrNotImplemented,
		},
		{
			name: "Custom behavior",
			mockFunc: func(tnt, id string) (*Filter, error) {
				if tnt == "cgrates.org" && id == "id1" {
					return &Filter{}, nil
				}
				return nil, errors.New("custom error")
			},
			tnt:         "cgrates.org",
			id:          "id1",
			expectValue: &Filter{},
			expectError: nil,
		},
		{
			name: "Custom behavior with error",
			mockFunc: func(tnt, id string) (*Filter, error) {
				if tnt == "tenant2" && id == "id2" {
					return nil, errors.New("another custom error")
				}
				return nil, nil
			},
			tnt:         "tenant2",
			id:          "id2",
			expectValue: nil,
			expectError: errors.New("another custom error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbM := &DataDBMock{
				GetFilterDrvF: tt.mockFunc,
			}
			_, err := dbM.GetFilterDrv(tt.tnt, tt.id)
			if err == nil && tt.expectError != nil {
				t.Errorf("expected error %v, got nil", tt.expectError)
			} else if err != nil && tt.expectError == nil {
				t.Errorf("expected no error, got %v", err)
			} else if err != nil && err.Error() != tt.expectError.Error() {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
		})
	}
}

func TestSetFilterDrv(t *testing.T) {
	dbM := &DataDBMock{}
	filter := &Filter{}
	err := dbM.SetFilterDrv(filter)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestRemoveFilterDrv(t *testing.T) {
	dbM := &DataDBMock{}
	tenant := "cgrates.org"
	id := "id"
	err := dbM.RemoveFilterDrv(tenant, id)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestSetRouteProfileDrv(t *testing.T) {
	dbM := &DataDBMock{}
	routeProfile := &RouteProfile{}
	err := dbM.SetRouteProfileDrv(routeProfile)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestDataDBMockSetReverseDestinationDrv(t *testing.T) {
	dbMock := &DataDBMock{}
	expectedError := utils.ErrNotImplemented
	err := dbMock.SetReverseDestinationDrv("Set", []string{"val1", "val2"}, "value")
	if err != expectedError {
		t.Errorf("expected error %v, but got %v", expectedError, err)
	}
}

func TestDataDBMockGetReverseDestinationDrv(t *testing.T) {
	dbMock := &DataDBMock{}
	expectedError := utils.ErrNotImplemented
	result, err := dbMock.GetReverseDestinationDrv("Reverse", "Destination")
	if err != expectedError {
		t.Errorf("expected error %v, but got %v", expectedError, err)
	}
	if result != nil {
		t.Errorf("expected result to be nil, but got %v", result)
	}
}

func TestDataDBMockGetActionsDrv(t *testing.T) {
	dbMock := &DataDBMock{}
	expectedError := utils.ErrNotImplemented
	result, err := dbMock.GetActionsDrv("GetActions")
	if err != expectedError {
		t.Errorf("expected error %v, but got %v", expectedError, err)
	}
	if result != nil {
		t.Errorf("expected result to be nil, but got %v", result)
	}
}

func TestDataDBMockSetActionsDrv(t *testing.T) {
	dbMock := &DataDBMock{}
	expectedError := utils.ErrNotImplemented
	err := dbMock.SetActionsDrv("Actions", Actions{})
	if err != expectedError {
		t.Errorf("expected error %v, but got %v", expectedError, err)
	}
}

func TestRemoveAttributeProfileDrvNotImplemented(t *testing.T) {
	dbM := &DataDBMock{}
	err := dbM.RemoveAttributeProfileDrv("Profile", "Drv")
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestSetAttributeProfileDrvNotImplemented(t *testing.T) {
	dbM := &DataDBMock{}
	Profile := &AttributeProfile{}
	err := dbM.SetAttributeProfileDrv(Profile)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetAttributeProfileDrvNotImplemented(t *testing.T) {
	dbM := &DataDBMock{}
	profile, err := dbM.GetAttributeProfileDrv("Profile", "Drv")
	if profile != nil {
		t.Errorf("expected profile to be nil, but got %v", profile)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetThresholdProfileDrvNotImplemented(t *testing.T) {
	dbM := &DataDBMock{
		GetThresholdProfileDrvF: nil,
	}
	profile, err := dbM.GetThresholdProfileDrv("Tenant", "ID")
	if profile != nil {
		t.Errorf("expected profile to be nil, but got %v", profile)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestRemStatQueueDrvNotImplemented(t *testing.T) {
	dbM := &DataDBMock{}
	err := dbM.RemStatQueueDrv("Tenant", "ID")
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestSetStatQueueDrvNotImplemented(t *testing.T) {
	dbM := &DataDBMock{}
	dummySSQ := &StoredStatQueue{}
	dummySQ := &StatQueue{}
	err := dbM.SetStatQueueDrv(dummySSQ, dummySQ)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetStatQueueDrvNotImplemented(t *testing.T) {
	dbM := &DataDBMock{}
	sq, err := dbM.GetStatQueueDrv("Tenant", "ID")
	if sq != nil {
		t.Errorf("expected StatQueue to be nil, but got %v", sq)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestRemoveTrendDrvNotImplemented(t *testing.T) {
	dbM := &DataDBMock{}
	err := dbM.RemoveTrendDrv("dummyTenant", "dummyID")
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestSetTrendDrvNotImplemented(t *testing.T) {
	dbM := &DataDBMock{}
	tTrend := &Trend{}
	err := dbM.SetTrendDrv(tTrend)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetTrendDrvNotImplemented(t *testing.T) {
	dbM := &DataDBMock{}
	trend, err := dbM.GetTrendDrv("Tenant", "ID")
	if trend != nil {
		t.Errorf("expected Trend to be nil, but got %v", trend)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestRemTrendProfileDrv(t *testing.T) {
	dbM := &DataDBMock{
		RemTrendProfileDrvF: nil,
	}
	err := dbM.RemTrendProfileDrv("Tenant", "ID")
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
	dbM.RemTrendProfileDrvF = func(tenant, id string) error {
		return utils.ErrNotImplemented
	}
	err = dbM.RemTrendProfileDrv("Tenant", "ID")
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetTrendProfileDrv(t *testing.T) {
	dbM := &DataDBMock{}
	tenant := "Tenant"
	id := "ID"
	dummyTrendProfile := &TrendProfile{}
	dbM.GetTrendProfileDrvF = nil
	sg, err := dbM.GetTrendProfileDrv(tenant, id)
	if sg != nil {
		t.Errorf("Expected TrendProfile to be nil, but got %v", sg)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
	dbM.GetTrendProfileDrvF = func(tenant, id string) (*TrendProfile, error) {
		return dummyTrendProfile, nil
	}
	sg, err = dbM.GetTrendProfileDrv(tenant, id)
	if err == nil {
		t.Errorf("NOT_IMPLEMENTED")
	}
}

func TestRemoveResourceDrv(t *testing.T) {
	dbM := &DataDBMock{}
	err := dbM.RemoveResourceDrv("Tenant", "ID")
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestSetTimingDrv(t *testing.T) {
	dbM := &DataDBMock{}
	dummyTiming := &utils.TPTiming{}
	err := dbM.SetTimingDrv(dummyTiming)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetLoadHistory(t *testing.T) {
	dbM := &DataDBMock{}
	instances, err := dbM.GetLoadHistory(0, false, "Param")
	if instances != nil {
		t.Errorf("expected instances to be nil, but got %v", instances)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestRemoveActionsDrv(t *testing.T) {
	dbMock := &DataDBMock{}
	err := dbMock.RemoveActionsDrv("Drv")
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetSharedGroupDrv(t *testing.T) {
	dbMock := &DataDBMock{}
	result, err := dbMock.GetSharedGroupDrv("Shared")
	if result != nil {
		t.Errorf("expected result to be nil, got %v", result)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestSetSharedGroupDrv(t *testing.T) {
	dbMock := &DataDBMock{}
	sharedGroup := &SharedGroup{}
	err := dbMock.SetSharedGroupDrv(sharedGroup)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestRemoveSharedGroupDrv(t *testing.T) {
	dbMock := &DataDBMock{}
	id := "ID"
	err := dbMock.RemoveSharedGroupDrv(id)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetActionTriggersDrv(t *testing.T) {
	dbMock := &DataDBMock{}
	id := "ID"
	triggers, err := dbMock.GetActionTriggersDrv(id)
	if triggers != nil {
		t.Errorf("expected nil ActionTriggers, got %v", triggers)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestSetActionTriggersDrv(t *testing.T) {
	dbMock := &DataDBMock{}
	id := "ID"
	var triggers ActionTriggers
	err := dbMock.SetActionTriggersDrv(id, triggers)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestRemoveActionTriggersDrv(t *testing.T) {
	dbMock := &DataDBMock{}
	id := "ID"
	err := dbMock.RemoveActionTriggersDrv(id)
	if err != utils.ErrNotImplemented {
		t.Errorf("expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetActionPlanDrv(t *testing.T) {
	{
		dbMock := &DataDBMock{}
		key := "Key"
		result, err := dbMock.GetActionPlanDrv(key)
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		if err != utils.ErrNotImplemented {
			t.Errorf("expected error %v, got %v", utils.ErrNotImplemented, err)
		}
	}
	{
		expectedActionPlan := &ActionPlan{}
		customError := error(nil)
		dbMock := &DataDBMock{
			GetActionPlanDrvF: func(key string) (*ActionPlan, error) {
				if key != "Key" {
					t.Errorf("expected key 'Key', got %s", key)
				}
				return expectedActionPlan, customError
			},
		}
		result, err := dbMock.GetActionPlanDrv("Key")
		if result != expectedActionPlan {
			t.Errorf("expected %v, got %v", expectedActionPlan, result)
		}
		if err != customError {
			t.Errorf("expected %v, got %v", customError, err)
		}
	}
}

func TestSetActionPlanDrv(t *testing.T) {
	{
		dbMock := &DataDBMock{}
		key := "Key"
		actionPlan := &ActionPlan{}

		err := dbMock.SetActionPlanDrv(key, actionPlan)

		if err != utils.ErrNotImplemented {
			t.Errorf("expected error %v, got %v", utils.ErrNotImplemented, err)
		}
	}

	{
		expectedError := error(nil)
		dbMock := &DataDBMock{
			SetActionPlanDrvF: func(key string, ap *ActionPlan) error {
				if key != "Key" {
					t.Errorf("expected key 'Key', got %s", key)
				}
				if ap == nil {
					t.Error("expected non-nil ActionPlan")
				}
				return expectedError
			},
		}
		actionPlan := &ActionPlan{}
		err := dbMock.SetActionPlanDrv("Key", actionPlan)

		if err == expectedError {
			t.Errorf("expected error %v, got %v", expectedError, err)
		}
	}
}

func TestGetAllActionPlansDrv(t *testing.T) {
	{
		dbMock := &DataDBMock{}
		actionPlans, err := dbMock.GetAllActionPlansDrv()
		if actionPlans != nil {
			t.Errorf("expected nil, got %v", actionPlans)
		}
		if err != utils.ErrNotImplemented {
			t.Errorf("expected error %v, got %v", utils.ErrNotImplemented, err)
		}
	}
	{
		expectedActionPlans := map[string]*ActionPlan{}
		expectedError := error(nil)
		dbMock := &DataDBMock{}
		actionPlans, err := dbMock.GetAllActionPlansDrv()
		if len(actionPlans) != len(expectedActionPlans) {
			t.Errorf("expected actionPlans length %d, got %d", len(expectedActionPlans), len(actionPlans))
		}
		for key, expectedPlan := range expectedActionPlans {
			if plan, exists := actionPlans[key]; !exists || plan != expectedPlan {
				t.Errorf("expected action plan for key %s to be %v, got %v", key, expectedPlan, plan)
			}
		}
		if err == expectedError {
			t.Errorf("expected error %v, got %v", expectedError, err)
		}
	}
}

func TestPopTask(t *testing.T) {
	mock := &DataDBMock{}
	task, err := mock.PopTask()
	if task != nil {
		t.Errorf("Expected task to be nil, got %v", task)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestPushTask(t *testing.T) {
	mock := &DataDBMock{}
	task := &Task{}
	err := mock.PushTask(task)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestRemoveIndexesDrv(t *testing.T) {
	mock := &DataDBMock{}
	err := mock.RemoveIndexesDrv("Type", "Context", "Key")
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestAddLoadHistory(t *testing.T) {
	mock := &DataDBMock{}
	loadInstance := &utils.LoadInstance{}
	err := mock.AddLoadHistory(loadInstance, 42, "test")
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestRemoveTimingDrv(t *testing.T) {
	mock := &DataDBMock{}
	err := mock.RemoveTimingDrv("Param")
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetTimingDrv(t *testing.T) {
	mock := &DataDBMock{}
	result, err := mock.GetTimingDrv("Param")
	if result != nil {
		t.Errorf("Expected result to be nil, got %v", result)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetResourceDrv(t *testing.T) {
	mock := &DataDBMock{}
	result, err := mock.GetResourceDrv("Param1", "Param2")
	if result != nil {
		t.Errorf("Expected result to be nil, got %v", result)
	}
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestRemAccountActionPlansDrv(t *testing.T) {
	mock := &DataDBMock{}
	err := mock.RemAccountActionPlansDrv("AccountID")
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestRemoveRankingDrv(t *testing.T) {
	dbMock := &DataDBMock{}

	err := dbMock.RemoveRankingDrv("Rank1", "Rank2")

	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v but got %v", utils.ErrNotImplemented, err)
	}

	err = dbMock.RemoveRankingDrv("Rank1", "Rank2")

	err = dbMock.RemoveRankingDrv("Rank3", "Rank4")
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v but got %v", utils.ErrNotImplemented, err)
	}
}

func TestSetRankingDrv(t *testing.T) {
	dbMock := &DataDBMock{}

	err := dbMock.SetRankingDrv(&Ranking{ID: "Ranking"})

	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v but got %v", utils.ErrNotImplemented, err)
	}

	err = dbMock.SetRankingDrv(&Ranking{ID: "Ranking"})

	err = dbMock.SetRankingDrv(&Ranking{ID: "Ranking2"})
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v but got %v", utils.ErrNotImplemented, err)
	}
}

func TestRemRankingProfileDrv(t *testing.T) {
	dbMock := &DataDBMock{}

	err := dbMock.RemRankingProfileDrv("cgrates.org", "rankingID1")

	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v, but got %v", utils.ErrNotImplemented, err)
	}

	dbMock.RemRankingProfileDrvF = func(tenant string, id string) error {
		if tenant == "cgrates.org" && id == "rankingID1" {
			return nil
		}
		return utils.ErrNotImplemented
	}

	err = dbMock.RemRankingProfileDrv("cgrates.org", "rankingID1")

	if err != nil {
		t.Errorf("Expected no error but got %v", err)
	}

	err = dbMock.RemRankingProfileDrv("cgrates.org", "rankingID2")
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v, but got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetAccountActionPlansDrv(t *testing.T) {
	dbMock := &DataDBMock{}
	acntID := "ID"
	apIDs, err := dbMock.GetAccountActionPlansDrv(acntID)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v but got %v", utils.ErrNotImplemented, err)
	}
	if apIDs != nil {
		t.Errorf("Expected apIDs to be nil but got %v", apIDs)
	}
}

func TestAccountSetActionPlansDrv(t *testing.T) {
	dbMock := &DataDBMock{}
	acntID := "ID"
	apIDs := []string{"plan1", "plan2"}
	err := dbMock.SetAccountActionPlansDrv(acntID, apIDs)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v but got %v", utils.ErrNotImplemented, err)
	}
}

func TestSetResourceDrv(t *testing.T) {
	dbMock := &DataDBMock{}
	resource := &Resource{
		Tenant: "cgrates.org",
		ID:     "ID",
		Usages: map[string]*ResourceUsage{
			"usage1": {
				Tenant:     "cgrates.org",
				ID:         "ID",
				ExpiryTime: time.Now().Add(24 * time.Hour),
				Units:      100.0,
			},
		},
		TTLIdx: []string{"resource1", "resource2"},
		ttl:    nil,
		tUsage: nil,
		dirty:  nil,
		rPrf:   nil,
	}

	err := dbMock.SetResourceDrv(resource)

	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v but got %v", utils.ErrNotImplemented, err)
	}

	dbMock.SetResourceDrvF = func(r *Resource) error {
		if r.ID == "ID" {
			return nil
		}
		return utils.ErrNotImplemented
	}

	err = dbMock.SetResourceDrv(resource)

	if err != nil {
		t.Errorf("Expected no error but got %v", err)
	}

	tResource := &Resource{
		Tenant: "cgrates.org",
		ID:     "ID1",
		Usages: nil,
	}

	err = dbMock.SetResourceDrv(tResource)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v but got %v", utils.ErrNotImplemented, err)
	}
}
