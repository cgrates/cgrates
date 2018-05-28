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

package utils

import (
	"errors"
	"strings"
)

// Wraps regexp compiling in case of rsr fields
func NewDerivedCharger(runId, runFilters, reqTypeFld, dirFld, tenantFld, catFld, acntFld, subjFld, dstFld, sTimeFld, pddFld, aTimeFld, durFld,
	supplFld, dCauseFld, preRatedFld, costFld string) (dc *DerivedCharger, err error) {
	if len(runId) == 0 {
		return nil, errors.New("Empty run id field")
	}
	dc = &DerivedCharger{RunID: runId}
	dc.RunFilters = runFilters
	if strings.HasPrefix(dc.RunFilters, REGEXP_PREFIX) || strings.HasPrefix(dc.RunFilters, STATIC_VALUE_PREFIX) {
		if dc.rsrRunFilters, err = ParseRSRFields(dc.RunFilters, INFIELD_SEP); err != nil {
			return nil, err
		}
	}
	dc.RequestTypeField = reqTypeFld
	if strings.HasPrefix(dc.RequestTypeField, REGEXP_PREFIX) || strings.HasPrefix(dc.RequestTypeField, STATIC_VALUE_PREFIX) {
		if dc.rsrRequestTypeField, err = NewRSRField(dc.RequestTypeField); err != nil {
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
	dc.PDDField = pddFld
	if strings.HasPrefix(dc.PDDField, REGEXP_PREFIX) || strings.HasPrefix(dc.PDDField, STATIC_VALUE_PREFIX) {
		if dc.rsrPddField, err = NewRSRField(dc.PDDField); err != nil {
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
	dc.PreRatedField = preRatedFld
	if strings.HasPrefix(dc.PreRatedField, REGEXP_PREFIX) || strings.HasPrefix(dc.PreRatedField, STATIC_VALUE_PREFIX) {
		if dc.rsrPreRatedField, err = NewRSRField(dc.PreRatedField); err != nil {
			return nil, err
		}
	}
	dc.CostField = costFld
	if strings.HasPrefix(dc.CostField, REGEXP_PREFIX) || strings.HasPrefix(dc.CostField, STATIC_VALUE_PREFIX) {
		if dc.rsrCostField, err = NewRSRField(dc.CostField); err != nil {
			return nil, err
		}
	}
	return dc, nil
}

type DerivedCharger struct {
	RunID                   string      // Unique runId in the chain
	RunFilters              string      // Only run the charger if all the filters match
	RequestTypeField        string      // Field containing request type info, number in case of csv source, '^' as prefix in case of static values
	DirectionField          string      // Field containing direction info
	TenantField             string      // Field containing tenant info
	CategoryField           string      // Field containing tor info
	AccountField            string      // Field containing account information
	SubjectField            string      // Field containing subject information
	DestinationField        string      // Field containing destination information
	SetupTimeField          string      // Field containing setup time information
	PDDField                string      // Field containing setup time information
	AnswerTimeField         string      // Field containing answer time information
	UsageField              string      // Field containing usage information
	SupplierField           string      // Field containing supplier information
	DisconnectCauseField    string      // Field containing disconnect cause information
	CostField               string      // Field containing cost information
	PreRatedField           string      // Field marking rated request in CDR
	rsrRunFilters           []*RSRField // Storage for compiled Regexp in case of RSRFields
	rsrRequestTypeField     *RSRField
	rsrDirectionField       *RSRField
	rsrTenantField          *RSRField
	rsrCategoryField        *RSRField
	rsrAccountField         *RSRField
	rsrSubjectField         *RSRField
	rsrDestinationField     *RSRField
	rsrSetupTimeField       *RSRField
	rsrPddField             *RSRField
	rsrAnswerTimeField      *RSRField
	rsrUsageField           *RSRField
	rsrSupplierField        *RSRField
	rsrDisconnectCauseField *RSRField
	rsrCostField            *RSRField
	rsrPreRatedField        *RSRField
}

func (dc *DerivedCharger) Equal(other *DerivedCharger) bool {
	return dc.RunID == other.RunID &&
		dc.RunFilters == other.RunFilters &&
		dc.RequestTypeField == other.RequestTypeField &&
		dc.DirectionField == other.DirectionField &&
		dc.TenantField == other.TenantField &&
		dc.CategoryField == other.CategoryField &&
		dc.AccountField == other.AccountField &&
		dc.SubjectField == other.SubjectField &&
		dc.DestinationField == other.DestinationField &&
		dc.SetupTimeField == other.SetupTimeField &&
		dc.PDDField == other.PDDField &&
		dc.AnswerTimeField == other.AnswerTimeField &&
		dc.UsageField == other.UsageField &&
		dc.SupplierField == other.SupplierField &&
		dc.DisconnectCauseField == other.DisconnectCauseField &&
		dc.CostField == other.CostField &&
		dc.PreRatedField == other.PreRatedField
}

func DerivedChargersKey(direction, tenant, category, account, subject string) string {
	return ConcatenatedKey(direction, tenant, category, account, subject)
}

type DerivedChargers struct {
	DestinationIDs StringMap
	Chargers       []*DerivedCharger
}

// Precheck that RunId is unique
func (dcs *DerivedChargers) Append(dc *DerivedCharger) (*DerivedChargers, error) {
	if dc.RunID == DEFAULT_RUNID {
		return nil, errors.New("Reserved RunId")
	}
	for _, dcLocal := range dcs.Chargers {
		if dcLocal.RunID == dc.RunID {
			return nil, errors.New("Duplicated RunId")
		}
	}
	dcs.Chargers = append(dcs.Chargers, dc)
	return dcs, nil
}

func (dcs *DerivedChargers) AppendDefaultRun() (*DerivedChargers, error) {
	dcDf, _ := NewDerivedCharger(DEFAULT_RUNID, "", META_DEFAULT, META_DEFAULT,
		META_DEFAULT, META_DEFAULT, META_DEFAULT,
		META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT,
		META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT, META_DEFAULT)
	dcs.Chargers = append(dcs.Chargers, dcDf)
	return dcs, nil
}

func (dcs *DerivedChargers) Equal(other *DerivedChargers) bool {
	dcs.DestinationIDs.Equal(other.DestinationIDs)
	for i, dc := range dcs.Chargers {
		if !dc.Equal(other.Chargers[i]) {
			return false
		}
	}
	return true
}
