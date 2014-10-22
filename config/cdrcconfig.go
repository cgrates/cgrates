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
	"fmt"
	"github.com/cgrates/cgrates/utils"
	"strings"
	"time"
)

func NewCdrcConfigFromCgrXmlCdrcCfg(id string, xmlCdrcCfg *CgrXmlCdrcCfg) (*CdrcConfig, error) {
	cdrcCfg := NewDefaultCdrcConfig()
	cdrcCfg.Id = id
	if xmlCdrcCfg.Enabled != nil {
		cdrcCfg.Enabled = *xmlCdrcCfg.Enabled
	}
	if xmlCdrcCfg.CdrsAddress != nil {
		cdrcCfg.CdrsAddress = *xmlCdrcCfg.CdrsAddress
	}
	if xmlCdrcCfg.CdrFormat != nil {
		cdrcCfg.CdrFormat = *xmlCdrcCfg.CdrFormat
	}
	if xmlCdrcCfg.FieldSeparator != nil {
		cdrcCfg.FieldSeparator = *xmlCdrcCfg.FieldSeparator
	}
	if xmlCdrcCfg.DataUsageMultiplyFactor != nil {
		cdrcCfg.DataUsageMultiplyFactor = *xmlCdrcCfg.DataUsageMultiplyFactor
	}
	if xmlCdrcCfg.RunDelay != nil {
		cdrcCfg.RunDelay = time.Duration(*xmlCdrcCfg.RunDelay) * time.Second
	}
	if xmlCdrcCfg.CdrInDir != nil {
		cdrcCfg.CdrInDir = *xmlCdrcCfg.CdrInDir
	}
	if xmlCdrcCfg.CdrOutDir != nil {
		cdrcCfg.CdrOutDir = *xmlCdrcCfg.CdrOutDir
	}
	if xmlCdrcCfg.CdrSourceId != nil {
		cdrcCfg.CdrSourceId = *xmlCdrcCfg.CdrSourceId
	}
	if len(xmlCdrcCfg.CdrFields) != 0 {
		cdrcCfg.CdrFields = nil // Reinit the fields, so we do not inherit from defaults here
	}
	for _, xmlCdrField := range xmlCdrcCfg.CdrFields {
		if cdrFld, err := NewCfgCdrFieldFromCgrXmlCfgCdrField(xmlCdrField, cdrcCfg.CdrFormat == utils.CDRE_FIXED_WIDTH); err != nil {
			return nil, err
		} else {
			cdrcCfg.CdrFields = append(cdrcCfg.CdrFields, cdrFld)
		}
	}
	return cdrcCfg, nil
}

func NewDefaultCdrcConfig() *CdrcConfig {
	torTag, accIdTag, reqTypeTag, dirTag, tenantTag, categTag, acntTag, subjTag, dstTag, sTimeTag, aTimeTag, usageTag := utils.TOR,
		utils.ACCID, utils.REQTYPE, utils.DIRECTION, utils.TENANT, utils.CATEGORY, utils.ACCOUNT, utils.SUBJECT, utils.DESTINATION, utils.SETUP_TIME, utils.ANSWER_TIME, utils.USAGE
	torFld, _ := NewCfgCdrFieldWithDefaults(false, []*utils.RSRField{&utils.RSRField{Id: "2"}}, nil, nil, &torTag, nil, nil, nil, nil, nil, nil)
	accIdFld, _ := NewCfgCdrFieldWithDefaults(false, []*utils.RSRField{&utils.RSRField{Id: "3"}}, nil, nil, &accIdTag, nil, nil, nil, nil, nil, nil)
	reqTypeFld, _ := NewCfgCdrFieldWithDefaults(false, []*utils.RSRField{&utils.RSRField{Id: "4"}}, nil, nil, &reqTypeTag, nil, nil, nil, nil, nil, nil)
	directionFld, _ := NewCfgCdrFieldWithDefaults(false, []*utils.RSRField{&utils.RSRField{Id: "5"}}, nil, nil, &dirTag, nil, nil, nil, nil, nil, nil)
	tenantFld, _ := NewCfgCdrFieldWithDefaults(false, []*utils.RSRField{&utils.RSRField{Id: "6"}}, nil, nil, &tenantTag, nil, nil, nil, nil, nil, nil)
	categoryFld, _ := NewCfgCdrFieldWithDefaults(false, []*utils.RSRField{&utils.RSRField{Id: "7"}}, nil, nil, &categTag, nil, nil, nil, nil, nil, nil)
	acntFld, _ := NewCfgCdrFieldWithDefaults(false, []*utils.RSRField{&utils.RSRField{Id: "8"}}, nil, nil, &acntTag, nil, nil, nil, nil, nil, nil)
	subjFld, _ := NewCfgCdrFieldWithDefaults(false, []*utils.RSRField{&utils.RSRField{Id: "9"}}, nil, nil, &subjTag, nil, nil, nil, nil, nil, nil)
	dstFld, _ := NewCfgCdrFieldWithDefaults(false, []*utils.RSRField{&utils.RSRField{Id: "10"}}, nil, nil, &dstTag, nil, nil, nil, nil, nil, nil)
	setupTimeFld, _ := NewCfgCdrFieldWithDefaults(false, []*utils.RSRField{&utils.RSRField{Id: "11"}}, nil, nil, &sTimeTag, nil, nil, nil, nil, nil, nil)
	answerTimeFld, _ := NewCfgCdrFieldWithDefaults(false, []*utils.RSRField{&utils.RSRField{Id: "12"}}, nil, nil, &aTimeTag, nil, nil, nil, nil, nil, nil)
	usageFld, _ := NewCfgCdrFieldWithDefaults(false, []*utils.RSRField{&utils.RSRField{Id: "13"}}, nil, nil, &usageTag, nil, nil, nil, nil, nil, nil)
	cdrcCfg := &CdrcConfig{
		Id:                      utils.META_DEFAULT,
		Enabled:                 false,
		CdrsAddress:             "",
		CdrFormat:               utils.CSV,
		FieldSeparator:          utils.FIELDS_SEP,
		DataUsageMultiplyFactor: 1,
		RunDelay:                time.Duration(0),
		CdrInDir:                "/var/log/cgrates/cdrc/in",
		CdrOutDir:               "/var/log/cgrates/cdrc/out",
		CdrSourceId:             utils.CSV,
		CdrFields:               []*CfgCdrField{torFld, accIdFld, reqTypeFld, directionFld, tenantFld, categoryFld, acntFld, subjFld, dstFld, setupTimeFld, answerTimeFld, usageFld},
	}
	return cdrcCfg
}

func NewCdrcConfigFromFileParams(c *conf.ConfigFile) (*CdrcConfig, error) {
	var err error
	cdrcCfg := NewDefaultCdrcConfig()
	if hasOpt := c.HasOption("cdrc", "enabled"); hasOpt {
		cdrcCfg.Enabled, _ = c.GetBool("cdrc", "enabled")
	}
	if hasOpt := c.HasOption("cdrc", "cdrs"); hasOpt {
		cdrcCfg.CdrsAddress, _ = c.GetString("cdrc", "cdrs")
	}
	if hasOpt := c.HasOption("cdrc", "cdr_format"); hasOpt {
		cdrcCfg.CdrFormat, _ = c.GetString("cdrc", "cdr_format")
	}
	if hasOpt := c.HasOption("cdrc", "field_separator"); hasOpt {
		cdrcCfg.FieldSeparator, _ = c.GetString("cdrc", "field_separator")
	}
	if hasOpt := c.HasOption("cdrc", "data_usage_multiply_factor"); hasOpt {
		mf, _ := c.GetInt("cdrc", "data_usage_multiply_factor")
		cdrcCfg.DataUsageMultiplyFactor = int64(mf)
	}
	if hasOpt := c.HasOption("cdrc", "run_delay"); hasOpt {
		durStr, _ := c.GetString("cdrc", "run_delay")
		if cdrcCfg.RunDelay, err = utils.ParseDurationWithSecs(durStr); err != nil {
			return nil, err
		}
	}
	if hasOpt := c.HasOption("cdrc", "cdr_in_dir"); hasOpt {
		cdrcCfg.CdrInDir, _ = c.GetString("cdrc", "cdr_in_dir")
	}
	if hasOpt := c.HasOption("cdrc", "cdr_out_dir"); hasOpt {
		cdrcCfg.CdrOutDir, _ = c.GetString("cdrc", "cdr_out_dir")
	}
	if hasOpt := c.HasOption("cdrc", "cdr_source_id"); hasOpt {
		cdrcCfg.CdrSourceId, _ = c.GetString("cdrc", "cdr_source_id")
	}
	// Parse CdrFields
	torFld, _ := c.GetString("cdrc", "tor_field")
	accIdFld, _ := c.GetString("cdrc", "accid_field")
	reqtypeFld, _ := c.GetString("cdrc", "reqtype_field")
	directionFld, _ := c.GetString("cdrc", "direction_field")
	tenantFld, _ := c.GetString("cdrc", "tenant_field")
	categoryFld, _ := c.GetString("cdrc", "category_field")
	acntFld, _ := c.GetString("cdrc", "account_field")
	subjectFld, _ := c.GetString("cdrc", "subject_field")
	destFld, _ := c.GetString("cdrc", "destination_field")
	setupTimeFld, _ := c.GetString("cdrc", "setup_time_field")
	answerTimeFld, _ := c.GetString("cdrc", "answer_time_field")
	durFld, _ := c.GetString("cdrc", "usage_field")
	newVals := false
	for _, fldData := range [][]string{ // Need to keep fields order
		[]string{utils.TOR, torFld}, []string{utils.ACCID, accIdFld}, []string{utils.REQTYPE, reqtypeFld}, []string{utils.DIRECTION, directionFld},
		[]string{utils.TENANT, tenantFld}, []string{utils.CATEGORY, categoryFld}, []string{utils.ACCOUNT, acntFld}, []string{utils.SUBJECT, subjectFld},
		[]string{utils.DESTINATION, destFld}, []string{utils.SETUP_TIME, setupTimeFld}, []string{utils.ANSWER_TIME, answerTimeFld}, []string{utils.USAGE, durFld}} {
		if len(fldData[1]) != 0 {
			if rsrFlds, err := utils.ParseRSRFields(fldData[1], utils.INFIELD_SEP); err != nil {
				return nil, err
			} else if len(rsrFlds) > 0 {
				if !newVals { // Default values there, reset them since we have at least one new
					cdrcCfg.CdrFields = nil
					newVals = true
				}
				if cdrcFld, err := NewCfgCdrFieldWithDefaults(false, rsrFlds, nil, nil, &fldData[0], nil, nil, nil, nil, nil, nil); err != nil {
					return nil, err
				} else {
					cdrcCfg.CdrFields = append(cdrcCfg.CdrFields, cdrcFld)
				}
			}
		}
	}
	extraFlds, _ := c.GetString("cdrc", "extra_fields")
	if len(extraFlds) != 0 {
		if sepExtraFlds, err := ConfigSlice(extraFlds); err != nil {
			return nil, err
		} else {
			for _, fldStr := range sepExtraFlds {
				// extra fields defined as: <label_extrafield_1>:<index_extrafield_1>
				if spltLbl := strings.Split(fldStr, utils.CONCATENATED_KEY_SEP); len(spltLbl) != 2 {
					return nil, fmt.Errorf("Wrong format for cdrc.extra_fields: %s", fldStr)
				} else {
					if rsrFlds, err := utils.ParseRSRFields(spltLbl[1], utils.INFIELD_SEP); err != nil {
						return nil, err
					} else if len(rsrFlds) > 0 {
						if cdrcFld, err := NewCfgCdrFieldWithDefaults(false, rsrFlds, nil, nil, &spltLbl[0], nil, nil, nil, nil, nil, nil); err != nil {
							return nil, err
						} else {
							cdrcCfg.CdrFields = append(cdrcCfg.CdrFields, cdrcFld)
						}
					}
				}
			}
		}
	}
	return cdrcCfg, nil
}

type CdrcConfig struct {
	Id                      string         // Configuration label
	Enabled                 bool           // Enable/Disable the profile
	CdrsAddress             string         // The address where CDRs can be reached
	CdrFormat               string         // The type of CDR file to process <csv>
	FieldSeparator          string         // The separator to use when reading csvs
	DataUsageMultiplyFactor int64          // Conversion factor for data usage
	RunDelay                time.Duration  // Delay between runs, 0 for inotify driven requests
	CdrInDir                string         // Folder to process CDRs from
	CdrOutDir               string         // Folder to move processed CDRs to
	CdrSourceId             string         // Source identifier for the processed CDRs
	CdrFields               []*CfgCdrField // List of fields to be processed
}
