/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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
func NewDerivedCharger(runId, runFilters, reqTypeFld, dirFld, tenantFld, catFld, acntFld, subjFld, dstFld, sTimeFld, aTimeFld, durFld, supplFld, dCauseFld string) (dc *DerivedCharger, err error) {
	if len(runId) == 0 {
		return nil, errors.New("Empty run id field")
	}
	dc = &DerivedCharger{RunId: runId}
	dc.RunFilters = runFilters
	if strings.HasPrefix(dc.RunFilters, REGEXP_PREFIX) || strings.HasPrefix(dc.RunFilters, STATIC_VALUE_PREFIX) {
		if dc.rsrRunFilters, err = ParseRSRFields(dc.RunFilters, INFIELD_SEP); err != nil {
			return nil, err
		}
	}
	dc.ReqTypeField = reqTypeFld
	if strings.HasPrefix(dc.ReqTypeField, REGEXP_PREFIX) || strings.HasPrefix(dc.ReqTypeField, STATIC_VALUE_PREFIX) {
		if dc.rsrReqTypeField, err = NewRSRField(dc.ReqTypeField); err != nil {
			return nil, err
		}
	}
	dc.DirectionField = dirFld
	if strings.HasPrefix(dc.DirectionField, REGEXP_PREFIX) || strings.HasPrefix(dc.DirectionField, STATIC_VALUE_PREFIX) {
		if dc.rsrDirectionField, err = NewRSRField(dc.DirectionField); err != nil {
			return nil, err
		}
	}
	dc.TenantField = tenantFld
	if strings.HasPrefix(dc.TenantField, REGEXP_PREFIX) || strings.HasPrefix(dc.TenantField, STATIC_VALUE_PREFIX) {
		if dc.rsrTenantField, err = NewRSRField(dc.TenantField); err != nil {
			return nil, err
		}
	}
	dc.CategoryField = catFld
	if strings.HasPrefix(dc.CategoryField, REGEXP_PREFIX) || strings.HasPrefix(dc.CategoryField, STATIC_VALUE_PREFIX) {
		if dc.rsrCategoryField, err = NewRSRField(dc.CategoryField); err != nil {
			return nil, err
		}
	}
	dc.AccountField = acntFld
	if strings.HasPrefix(dc.AccountField, REGEXP_PREFIX) || strings.HasPrefix(dc.AccountField, STATIC_VALUE_PREFIX) {
		if dc.rsrAccountField, err = NewRSRField(dc.AccountField); err != nil {
			return nil, err
		}
	}
	dc.SubjectField = subjFld
	if strings.HasPrefix(dc.SubjectField, REGEXP_PREFIX) || strings.HasPrefix(dc.SubjectField, STATIC_VALUE_PREFIX) {
		if dc.rsrSubjectField, err = NewRSRField(dc.SubjectField); err != nil {
			return nil, err
		}
	}
	dc.DestinationField = dstFld
	if strings.HasPrefix(dc.DestinationField, REGEXP_PREFIX) || strings.HasPrefix(dc.DestinationField, STATIC_VALUE_PREFIX) {
		if dc.rsrDestinationField, err = NewRSRField(dc.DestinationField); err != nil {
			return nil, err
		}
	}
	dc.SetupTimeField = sTimeFld
	if strings.HasPrefix(dc.SetupTimeField, REGEXP_PREFIX) || strings.HasPrefix(dc.SetupTimeField, STATIC_VALUE_PREFIX) {
		if dc.rsrSetupTimeField, err = NewRSRField(dc.SetupTimeField); err != nil {
			return nil, err
		}
	}
	dc.AnswerTimeField = aTimeFld
	if strings.HasPrefix(dc.AnswerTimeField, REGEXP_PREFIX) || strings.HasPrefix(dc.AnswerTimeField, STATIC_VALUE_PREFIX) {
		if dc.rsrAnswerTimeField, err = NewRSRField(dc.AnswerTimeField); err != nil {
			return nil, err
		}
	}
	dc.UsageField = durFld
	if strings.HasPrefix(dc.UsageField, REGEXP_PREFIX) || strings.HasPrefix(dc.UsageField, STATIC_VALUE_PREFIX) {
		if dc.rsrUsageField, err = NewRSRField(dc.UsageField); err != nil {
			return nil, err
		}
	}
	dc.SupplierField = supplFld
	if strings.HasPrefix(dc.SupplierField, REGEXP_PREFIX) || strings.HasPrefix(dc.SupplierField, STATIC_VALUE_PREFIX) {
		if dc.rsrSupplierField, err = NewRSRField(dc.SupplierField); err != nil {
			return nil, err
		}
	}
	dc.DisconnectCauseField = dCauseFld
	if strings.HasPrefix(dc.DisconnectCauseField, REGEXP_PREFIX) || strings.HasPrefix(dc.DisconnectCauseField, STATIC_VALUE_PREFIX) {
		if dc.rsrDisconnectCauseField, err = NewRSRField(dc.DisconnectCauseField); err != nil {
			return nil, err
		}
	}
	return dc, nil
}

type DerivedCharger struct {
	RunId                   string      // Unique runId in the chain
	RunFilters              string      // Only run the charger if all the filters match
	ReqTypeField            string      // Field containing request type info, number in case of csv source, '^' as prefix in case of static values
	DirectionField          string      // Field containing direction info
	TenantField             string      // Field containing tenant info
	CategoryField           string      // Field containing tor info
	AccountField            string      // Field containing account information
	SubjectField            string      // Field containing subject information
	DestinationField        string      // Field containing destination information
	SetupTimeField          string      // Field containing setup time information
	AnswerTimeField         string      // Field containing answer time information
	UsageField              string      // Field containing usage information
	SupplierField           string      // Field containing supplier information
	DisconnectCauseField    string      // Field containing disconnect cause information
	rsrRunFilters           []*RSRField // Storage for compiled Regexp in case of RSRFields
	rsrReqTypeField         *RSRField
	rsrDirectionField       *RSRField
	rsrTenantField          *RSRField
	rsrCategoryField        *RSRField
	rsrAccountField         *RSRField
	rsrSubjectField         *RSRField
	rsrDestinationField     *RSRField
	rsrSetupTimeField       *RSRField
	rsrAnswerTimeField      *RSRField
	rsrUsageField           *RSRField
	rsrSupplierField        *RSRField
	rsrDisconnectCauseField *RSRField
}

func DerivedChargersKey(direction, tenant, category, account, subject string) string {
	return ConcatenatedKey(direction, tenant, category, account, subject)
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
	dcDf, _ := NewDerivedCharger(DEFAULT_RUNID, "", META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT,
		META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT)
	return append(dcs, dcDf), nil
}
