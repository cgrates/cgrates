/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package utils

import (
	"errors"
	"strings"
)

// Wraps regexp compiling in case of rsr fields
func NewDerivedCharger(runId, reqTypeFld, dirFld, tenantFld, totFld, acntFld, subjFld, dstFld, sTimeFld, aTimeFld, durFld string) (dc *DerivedCharger, err error) {
	if len(runId) == 0 {
		return nil, errors.New("Empty run id field")
	}
	dc = &DerivedCharger{RunId: runId}
	dc.ReqTypeField = reqTypeFld
	if strings.HasPrefix(dc.ReqTypeField, REGEXP_PREFIX) {
		if dc.rsrReqTypeField, err = NewRSRField(dc.ReqTypeField); err != nil {
			return nil, err
		}
	}
	dc.DirectionField = dirFld
	if strings.HasPrefix(dc.DirectionField, REGEXP_PREFIX) {
		if dc.rsrDirectionField, err = NewRSRField(dc.DirectionField); err != nil {
			return nil, err
		}
	}
	dc.TenantField = tenantFld
	if strings.HasPrefix(dc.TenantField, REGEXP_PREFIX) {
		if dc.rsrTenantField, err = NewRSRField(dc.TenantField); err != nil {
			return nil, err
		}
	}
	dc.TorField = totFld
	if strings.HasPrefix(dc.TorField, REGEXP_PREFIX) {
		if dc.rsrTorField, err = NewRSRField(dc.TorField); err != nil {
			return nil, err
		}
	}
	dc.AccountField = acntFld
	if strings.HasPrefix(dc.AccountField, REGEXP_PREFIX) {
		if dc.rsrAccountField, err = NewRSRField(dc.AccountField); err != nil {
			return nil, err
		}
	}
	dc.SubjectField = subjFld
	if strings.HasPrefix(dc.SubjectField, REGEXP_PREFIX) {
		if dc.rsrSubjectField, err = NewRSRField(dc.SubjectField); err != nil {
			return nil, err
		}
	}
	dc.DestinationField = dstFld
	if strings.HasPrefix(dc.DestinationField, REGEXP_PREFIX) {
		if dc.rsrDestinationField, err = NewRSRField(dc.DestinationField); err != nil {
			return nil, err
		}
	}
	dc.SetupTimeField = sTimeFld
	if strings.HasPrefix(dc.SetupTimeField, REGEXP_PREFIX) {
		if dc.rsrSetupTimeField, err = NewRSRField(dc.SetupTimeField); err != nil {
			return nil, err
		}
	}
	dc.AnswerTimeField = aTimeFld
	if strings.HasPrefix(dc.AnswerTimeField, REGEXP_PREFIX) {
		if dc.rsrAnswerTimeField, err = NewRSRField(dc.AnswerTimeField); err != nil {
			return nil, err
		}
	}
	dc.DurationField = durFld
	if strings.HasPrefix(dc.DurationField, REGEXP_PREFIX) {
		if dc.rsrDurationField, err = NewRSRField(dc.DurationField); err != nil {
			return nil, err
		}
	}
	return dc, nil
}

type DerivedCharger struct {
	RunId               string    // Unique runId in the chain
	ReqTypeField        string    // Field containing request type info, number in case of csv source, '^' as prefix in case of static values
	DirectionField      string    // Field containing direction info
	TenantField         string    // Field containing tenant info
	TorField            string    // Field containing tor info
	AccountField        string    // Field containing account information
	SubjectField        string    // Field containing subject information
	DestinationField    string    // Field containing destination information
	SetupTimeField      string    // Field containing setup time information
	AnswerTimeField     string    // Field containing answer time information
	DurationField       string    // Field containing duration information
	rsrReqTypeField     *RSRField // Storage for compiled Regexp in case of RSRFields
	rsrDirectionField   *RSRField
	rsrTenantField      *RSRField
	rsrTorField         *RSRField
	rsrAccountField     *RSRField
	rsrSubjectField     *RSRField
	rsrDestinationField *RSRField
	rsrSetupTimeField   *RSRField
	rsrAnswerTimeField  *RSRField
	rsrDurationField    *RSRField
}

func DerivedChargersKey(tenant, tor, direction, account, subject string) string {
	return ConcatenatedKey(tenant, tor, direction, account, subject)
}

type DerivedChargers []*DerivedCharger

// Precheck that RunId is unique
func (dcs DerivedChargers) Append(dc *DerivedCharger) (DerivedChargers, error) {
	if dc.RunId == DEFAULT_RUNID {
		return nil, errors.New("Reserved RunId")
	}
	for _, dcLocal := range dcs {
		if dcLocal.RunId == dc.RunId {
			return nil, errors.New("Duplicated RunId")
		}
	}
	return append(dcs, dc), nil
}

func (dcs DerivedChargers) AppendDefaultRun() (DerivedChargers, error) {
	dcDf, _ := NewDerivedCharger(DEFAULT_RUNID, META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT,
		META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT)
	return append(dcs, dcDf), nil
}
