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

package cdrs

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

const (
	// Freswitch event property names
	FS_CDR_MAP      = "variables"
	FS_DIRECTION    = "direction"
	FS_SUBJECT      = "cgr_subject"
	FS_ACCOUNT      = "cgr_account"
	FS_DESTINATION  = "cgr_destination"
	FS_REQTYPE      = "cgr_reqtype" //prepaid or postpaid
	FS_CATEGORY     = "cgr_category"
	FS_UUID         = "uuid" // -Unique ID for this call leg
	FS_CSTMID       = "cgr_tenant"
	FS_CALL_DEST_NR = "dialed_extension"
	FS_PARK_TIME    = "start_epoch"
	FS_SETUP_TIME   = "start_epoch"
	FS_ANSWER_TIME  = "answer_epoch"
	FS_HANGUP_TIME  = "end_epoch"
	FS_DURATION     = "billsec"
	FS_USERNAME     = "user_name"
	FS_IP           = "sip_local_network_addr"
	FS_CDR_SOURCE   = "freeswitch_json"
	FS_SIP_REQUSER  = "sip_req_user" // Apps like FusionPBX do not set dialed_extension, alternative being destination_number but that comes in customer profile, not in vars
)

func NewFSCdr(body []byte) (*FSCdr, error) {
	fsCdr := new(FSCdr)
	fsCdr.vars = make(map[string]string)
	var err error
	if err = json.Unmarshal(body, &fsCdr.body); err == nil {
		if variables, ok := fsCdr.body[FS_CDR_MAP]; ok {
			if variables, ok := variables.(map[string]interface{}); ok {
				for k, v := range variables {
					fsCdr.vars[k] = v.(string)
				}
			}
			return fsCdr, nil
		}
	}
	return nil, err
}

type FSCdr struct {
	vars map[string]string
	body map[string]interface{} // keeps the loaded body for extra field search
}

func (fsCdr FSCdr) getCgrId() string {
	setupTime, _ := utils.ParseTimeDetectLayout(fsCdr.vars[FS_SETUP_TIME])
	return utils.Sha1(fsCdr.vars[FS_UUID], setupTime.UTC().String())
}

func (fsCdr FSCdr) getExtraFields() map[string]string {
	extraFields := make(map[string]string, len(cfg.CDRSExtraFields))
	for _, field := range cfg.CDRSExtraFields {
		origFieldVal, foundInVars := fsCdr.vars[field.Id]
		if strings.HasPrefix(field.Id, utils.STATIC_VALUE_PREFIX) { // Support for static values injected in the CDRS. it will show up as {^value:value}
			foundInVars = true
		}
		if !foundInVars {
			origFieldVal = fsCdr.searchExtraField(field.Id, fsCdr.body)
		}
		extraFields[field.Id] = field.ParseValue(origFieldVal)
	}
	return extraFields
}

func (fsCdr FSCdr) searchExtraField(field string, body map[string]interface{}) (result string) {
	for key, value := range body {
		switch v := value.(type) {
		case string:
			if key == field {
				return v
			}
		case map[string]interface{}:
			if result = fsCdr.searchExtraField(field, v); result != "" {
				return
			}
		case []interface{}:
			for _, item := range v {
				if otherMap, ok := item.(map[string]interface{}); ok {
					if result = fsCdr.searchExtraField(field, otherMap); result != "" {
						return
					}
				} else {
					engine.Logger.Warning(fmt.Sprintf("Slice with no maps: %v", reflect.TypeOf(item)))
				}
			}
		default:
			engine.Logger.Warning(fmt.Sprintf("Unexpected type: %v", reflect.TypeOf(v)))
		}
	}
	return
}

func (fsCdr FSCdr) AsStoredCdr() *utils.StoredCdr {
	storCdr := new(utils.StoredCdr)
	storCdr.CgrId = fsCdr.getCgrId()
	storCdr.TOR = utils.VOICE
	storCdr.AccId = fsCdr.vars[FS_UUID]
	storCdr.CdrHost = fsCdr.vars[FS_IP]
	storCdr.CdrSource = FS_CDR_SOURCE
	storCdr.ReqType = utils.FirstNonEmpty(fsCdr.vars[FS_REQTYPE], cfg.DefaultReqType)
	storCdr.Direction = "*out"
	storCdr.Tenant = utils.FirstNonEmpty(fsCdr.vars[FS_CSTMID], cfg.DefaultTenant)
	storCdr.Category = utils.FirstNonEmpty(fsCdr.vars[FS_CATEGORY], cfg.DefaultCategory)
	storCdr.Account = utils.FirstNonEmpty(fsCdr.vars[FS_ACCOUNT], fsCdr.vars[FS_USERNAME])
	storCdr.Subject = utils.FirstNonEmpty(fsCdr.vars[FS_SUBJECT], fsCdr.vars[FS_USERNAME])
	storCdr.Destination = utils.FirstNonEmpty(fsCdr.vars[FS_DESTINATION], fsCdr.vars[FS_CALL_DEST_NR], fsCdr.vars[FS_SIP_REQUSER])
	storCdr.SetupTime, _ = utils.ParseTimeDetectLayout(fsCdr.vars[FS_SETUP_TIME]) // Not interested to process errors, should do them if necessary in a previous step
	storCdr.AnswerTime, _ = utils.ParseTimeDetectLayout(fsCdr.vars[FS_ANSWER_TIME])
	storCdr.Usage, _ = utils.ParseDurationWithSecs(fsCdr.vars[FS_DURATION])
	storCdr.ExtraFields = fsCdr.getExtraFields()
	storCdr.Cost = -1
	return storCdr
}
