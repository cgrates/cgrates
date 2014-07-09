/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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

package config

import (
	"code.google.com/p/goconf/conf"
	"errors"
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// Adds support for slice values in config
func ConfigSlice(cfgVal string) ([]string, error) {
	cfgValStrs := strings.Split(cfgVal, ",") // If need arrises, we can make the separator configurable
	for idx, elm := range cfgValStrs {
		//if elm == "" { //One empty element is presented when splitting empty string
		//	return nil, errors.New("Empty values in config slice")
		//}
		cfgValStrs[idx] = strings.TrimSpace(elm) // By default spaces are not removed so we do it here to avoid unpredicted results in config
	}
	return cfgValStrs, nil
}

func ParseRSRFields(configVal string) ([]*utils.RSRField, error) { //ToDo: Unify it with the Parser inside RSRField
	cfgValStrs := strings.Split(configVal, string(utils.CSV_SEP))
	if len(cfgValStrs) == 1 && cfgValStrs[0] == "" { // Prevents returning iterable with empty value
		return []*utils.RSRField{}, nil
	}
	rsrFields := make([]*utils.RSRField, len(cfgValStrs))
	for idx, cfgValStr := range cfgValStrs {
		if rsrField, err := utils.NewRSRField(cfgValStr); err != nil {
			return nil, err
		} else {
			rsrFields[idx] = rsrField
		}
	}
	return rsrFields, nil
}

// Parse the configuration file and returns utils.DerivedChargers instance if no errors
func ParseCfgDerivedCharging(c *conf.ConfigFile) (dcs utils.DerivedChargers, err error) {
	var runIds, runFilters, reqTypeFlds, directionFlds, tenantFlds, torFlds, acntFlds, subjFlds, dstFlds, sTimeFlds, aTimeFlds, durFlds []string
	cfgVal, _ := c.GetString("derived_charging", "run_ids")
	if runIds, err = ConfigSlice(cfgVal); err != nil {
		return nil, err
	}
	cfgVal, _ = c.GetString("derived_charging", "run_filters")
	if runFilters, err = ConfigSlice(cfgVal); err != nil {
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
	cfgVal, _ = c.GetString("derived_charging", "category_fields")
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
	cfgVal, _ = c.GetString("derived_charging", "usage_fields")
	if durFlds, err = ConfigSlice(cfgVal); err != nil {
		return nil, err
	}
	// We need all to be the same length
	if len(runFilters) != len(runIds) ||
		len(reqTypeFlds) != len(runIds) ||
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
	dcs = make(utils.DerivedChargers, 0)
	if len(runIds) == 1 && len(runIds[0]) == 0 { // Avoid iterating on empty runid
		return dcs, nil
	}
	for runIdx, runId := range runIds {
		dc, err := utils.NewDerivedCharger(runId, runFilters[runIdx], reqTypeFlds[runIdx], directionFlds[runIdx], tenantFlds[runIdx], torFlds[runIdx],
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

func ParseCdrcCdrFields(torFld, accIdFld, reqtypeFld, directionFld, tenantFld, categoryFld, acntFld, subjectFld, destFld,
	setupTimeFld, answerTimeFld, durFld, extraFlds string) (map[string]*utils.RSRField, error) {
	cdrcCdrFlds := make(map[string]*utils.RSRField)
	if len(extraFlds) != 0 {
		if sepExtraFlds, err := ConfigSlice(extraFlds); err != nil {
			return nil, err
		} else {
			for _, fldStr := range sepExtraFlds {
				// extra fields defined as: <label_extrafield_1>:<index_extrafield_1>
				if spltLbl := strings.Split(fldStr, utils.CONCATENATED_KEY_SEP); len(spltLbl) != 2 {
					return nil, fmt.Errorf("Wrong format for cdrc.extra_fields: %s", fldStr)
				} else {
					if rsrFld, err := utils.NewRSRField(spltLbl[1]); err != nil {
						return nil, err
					} else {
						cdrcCdrFlds[spltLbl[0]] = rsrFld
					}
				}
			}
		}
	}
	for fldTag, fldVal := range map[string]string{utils.TOR: torFld, utils.ACCID: accIdFld, utils.REQTYPE: reqtypeFld, utils.DIRECTION: directionFld, utils.TENANT: tenantFld,
		utils.CATEGORY: categoryFld, utils.ACCOUNT: acntFld, utils.SUBJECT: subjectFld, utils.DESTINATION: destFld, utils.SETUP_TIME: setupTimeFld,
		utils.ANSWER_TIME: answerTimeFld, utils.USAGE: durFld} {
		if len(fldVal) != 0 {
			if rsrFld, err := utils.NewRSRField(fldVal); err != nil {
				return nil, err
			} else {
				cdrcCdrFlds[fldTag] = rsrFld
			}
		}
	}
	return cdrcCdrFlds, nil
}
