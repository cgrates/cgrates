/*
Rating system designed to be used in VoIP Carriers World
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

package config

import (
	"code.google.com/p/goconf/conf"
	"errors"
	"github.com/cgrates/cgrates/utils"
	"strings"
)

// Wraps regexp compiling in case of rsr fields
func NewDerivedCharger(runId, reqTypeFld, dirFld, tenantFld, totFld, acntFld, subjFld, dstFld, sTimeFld, aTimeFld, durFld string) (dc *DerivedCharger, err error) {
	if len(runId) == 0 {
		return nil, errors.New("Empty run id field")
	}
	dc = &DerivedCharger{RunId: runId}
	dc.ReqTypeField = reqTypeFld
	if strings.HasPrefix(dc.ReqTypeField, utils.REGEXP_PREFIX) {
		if dc.rsrReqTypeField, err = utils.NewRSRField(dc.ReqTypeField); err != nil {
			return nil, err
		}
	}
	dc.DirectionField = dirFld
	if strings.HasPrefix(dc.DirectionField, utils.REGEXP_PREFIX) {
		if dc.rsrDirectionField, err = utils.NewRSRField(dc.DirectionField); err != nil {
			return nil, err
		}
	}
	dc.TenantField = tenantFld
	if strings.HasPrefix(dc.TenantField, utils.REGEXP_PREFIX) {
		if dc.rsrTenantField, err = utils.NewRSRField(dc.TenantField); err != nil {
			return nil, err
		}
	}
	dc.TorField = totFld
	if strings.HasPrefix(dc.TorField, utils.REGEXP_PREFIX) {
		if dc.rsrTorField, err = utils.NewRSRField(dc.TorField); err != nil {
			return nil, err
		}
	}
	dc.AccountField = acntFld
	if strings.HasPrefix(dc.AccountField, utils.REGEXP_PREFIX) {
		if dc.rsrAccountField, err = utils.NewRSRField(dc.AccountField); err != nil {
			return nil, err
		}
	}
	dc.SubjectField = subjFld
	if strings.HasPrefix(dc.SubjectField, utils.REGEXP_PREFIX) {
		if dc.rsrSubjectField, err = utils.NewRSRField(dc.SubjectField); err != nil {
			return nil, err
		}
	}
	dc.DestinationField = dstFld
	if strings.HasPrefix(dc.DestinationField, utils.REGEXP_PREFIX) {
		if dc.rsrDestinationField, err = utils.NewRSRField(dc.DestinationField); err != nil {
			return nil, err
		}
	}
	dc.SetupTimeField = sTimeFld
	if strings.HasPrefix(dc.SetupTimeField, utils.REGEXP_PREFIX) {
		if dc.rsrSetupTimeField, err = utils.NewRSRField(dc.SetupTimeField); err != nil {
			return nil, err
		}
	}
	dc.AnswerTimeField = aTimeFld
	if strings.HasPrefix(dc.AnswerTimeField, utils.REGEXP_PREFIX) {
		if dc.rsrAnswerTimeField, err = utils.NewRSRField(dc.AnswerTimeField); err != nil {
			return nil, err
		}
	}
	dc.DurationField = durFld
	if strings.HasPrefix(dc.DurationField, utils.REGEXP_PREFIX) {
		if dc.rsrDurationField, err = utils.NewRSRField(dc.DurationField); err != nil {
			return nil, err
		}
	}
	return dc, nil
}

type DerivedCharger struct {
	RunId               string          // Unique runId in the chain
	ReqTypeField        string          // Field containing request type info, number in case of csv source, '^' as prefix in case of static values
	DirectionField      string          // Field containing direction info
	TenantField         string          // Field containing tenant info
	TorField            string          // Field containing tor info
	AccountField        string          // Field containing account information
	SubjectField        string          // Field containing subject information
	DestinationField    string          // Field containing destination information
	SetupTimeField      string          // Field containing setup time information
	AnswerTimeField     string          // Field containing answer time information
	DurationField       string          // Field containing duration information
	rsrReqTypeField     *utils.RSRField // Storage for compiled Regexp in case of RSRFields
	rsrDirectionField   *utils.RSRField
	rsrTenantField      *utils.RSRField
	rsrTorField         *utils.RSRField
	rsrAccountField     *utils.RSRField
	rsrSubjectField     *utils.RSRField
	rsrDestinationField *utils.RSRField
	rsrSetupTimeField   *utils.RSRField
	rsrAnswerTimeField  *utils.RSRField
	rsrDurationField    *utils.RSRField
}

type DerivedChargers []*DerivedCharger

// Precheck that RunId is unique
func (dcs DerivedChargers) Append(dc *DerivedCharger) (DerivedChargers, error) {
	if dc.RunId == utils.DEFAULT_RUNID {
		return nil, errors.New("Reserved RunId")
	}
	for _, dcLocal := range dcs {
		if dcLocal.RunId == dc.RunId {
			return nil, errors.New("Duplicated RunId")
		}
	}
	return append(dcs, dc), nil
}

// Parse the configuration file and returns DerivedChargers instance if no errors
func ParseCfgDerivedCharging(c *conf.ConfigFile) (dcs DerivedChargers, err error) {
	var runIds, reqTypeFlds, directionFlds, tenantFlds, torFlds, acntFlds, subjFlds, dstFlds, sTimeFlds, aTimeFlds, durFlds []string
	cfgVal, _ := c.GetString("derived_charging", "run_ids")
	if runIds, err = ConfigSlice(cfgVal); err != nil {
		return nil, err
	}
	cfgVal, _ = c.GetString("derived_charging", "reqtype_fields")
	if reqTypeFlds, err = ConfigSlice(cfgVal); err != nil {
		return nil, err
	}
	cfgVal, _ = c.GetString("derived_charging", "direction_fields")
	if directionFlds, err = ConfigSlice(cfgVal); err != nil {
		return nil, err
	}
	cfgVal, _ = c.GetString("derived_charging", "tenant_fields")
	if tenantFlds, err = ConfigSlice(cfgVal); err != nil {
		return nil, err
	}
	cfgVal, _ = c.GetString("derived_charging", "tor_fields")
	if torFlds, err = ConfigSlice(cfgVal); err != nil {
		return nil, err
	}
	cfgVal, _ = c.GetString("derived_charging", "account_fields")
	if acntFlds, err = ConfigSlice(cfgVal); err != nil {
		return nil, err
	}
	cfgVal, _ = c.GetString("derived_charging", "subject_fields")
	if subjFlds, err = ConfigSlice(cfgVal); err != nil {
		return nil, err
	}
	cfgVal, _ = c.GetString("derived_charging", "destination_fields")
	if dstFlds, err = ConfigSlice(cfgVal); err != nil {
		return nil, err
	}
	cfgVal, _ = c.GetString("derived_charging", "setup_time_fields")
	if sTimeFlds, err = ConfigSlice(cfgVal); err != nil {
		return nil, err
	}
	cfgVal, _ = c.GetString("derived_charging", "answer_time_fields")
	if aTimeFlds, err = ConfigSlice(cfgVal); err != nil {
		return nil, err
	}
	cfgVal, _ = c.GetString("derived_charging", "duration_fields")
	if durFlds, err = ConfigSlice(cfgVal); err != nil {
		return nil, err
	}
	// We need all to be the same length
	if len(reqTypeFlds) != len(runIds) ||
		len(directionFlds) != len(runIds) ||
		len(tenantFlds) != len(runIds) ||
		len(torFlds) != len(runIds) ||
		len(acntFlds) != len(runIds) ||
		len(subjFlds) != len(runIds) ||
		len(dstFlds) != len(runIds) ||
		len(sTimeFlds) != len(runIds) ||
		len(aTimeFlds) != len(runIds) ||
		len(durFlds) != len(runIds) {
		return nil, errors.New("<ConfigSanity> Inconsistent fields length in derivated_charging section")
	}
	// Create the individual chargers and append them to the final instance
	dcs = make(DerivedChargers, 0)
	for runIdx, runId := range runIds {
		dc, err := NewDerivedCharger(runId, reqTypeFlds[runIdx], directionFlds[runIdx], tenantFlds[runIdx], torFlds[runIdx],
			acntFlds[runIdx], subjFlds[runIdx], dstFlds[runIdx], sTimeFlds[runIdx], aTimeFlds[runIdx], durFlds[runIdx])
		if err != nil {
			return nil, err
		}
		if dcs, err = dcs.Append(dc); err != nil {
			return nil, err
		}
	}
	return dcs, nil
}
