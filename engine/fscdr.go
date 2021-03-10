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
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

const (
	// Freswitch event property names
	FS_CDR_MAP            = "variables"
	FS_UUID               = "uuid" // -Unique ID for this call leg
	FS_CALL_DEST_NR       = "dialed_extension"
	FS_PARK_TIME          = "start_epoch"
	FS_SETUP_TIME         = "start_epoch"
	FS_ANSWER_TIME        = "answer_epoch"
	FS_HANGUP_TIME        = "end_epoch"
	FS_DURATION           = "billsec"
	FS_USERNAME           = "user_name"
	FS_CDR_SOURCE         = "freeswitch_json"
	FS_SIP_REQUSER        = "sip_req_user" // Apps like FusionPBX do not set dialed_extension, alternative being destination_number but that comes in customer profile, not in vars
	FS_PROGRESS_MEDIAMSEC = "progress_mediamsec"
	FS_PROGRESSMS         = "progressmsec"
	FsUsername            = "username"
	FsIPv4                = "FreeSWITCH-IPv4"
)

func NewFSCdr(body io.Reader, cgrCfg *config.CGRConfig) (*FSCdr, error) {
	fsCdr := &FSCdr{cgrCfg: cgrCfg, vars: make(map[string]string)}
	var err error
	if err = json.NewDecoder(body).Decode(&fsCdr.body); err != nil {
		return nil, err
	}
	if variables, ok := fsCdr.body[FS_CDR_MAP]; ok {
		if variables, ok := variables.(map[string]interface{}); ok {
			for k, v := range variables {
				fsCdr.vars[k] = v.(string)
			}
		}
	}
	return fsCdr, nil
}

type FSCdr struct {
	cgrCfg *config.CGRConfig
	vars   map[string]string
	body   map[string]interface{} // keeps the loaded body for extra field search
}

func (fsCdr FSCdr) getCGRID() string {
	return utils.Sha1(fsCdr.vars[FS_UUID],
		utils.FirstNonEmpty(fsCdr.vars[utils.CGROriginHost], fsCdr.vars[FsIPv4]))
}

func (fsCdr FSCdr) getExtraFields() map[string]string {
	extraFields := make(map[string]string, len(fsCdr.cgrCfg.CdrsCfg().ExtraFields))
	const dynprefix string = utils.MetaDynReq + utils.NestingSep
	for _, field := range fsCdr.cgrCfg.CdrsCfg().ExtraFields {
		if !strings.HasPrefix(field.Rules, dynprefix) {
			continue
		}
		attrName := field.AttrName()[5:]
		origFieldVal, foundInVars := fsCdr.vars[attrName]
		if !foundInVars {
			origFieldVal = fsCdr.searchExtraField(attrName, fsCdr.body)
		}
		if parsed, err := field.ParseValue(origFieldVal); err == nil {
			extraFields[attrName] = parsed
		}
	}
	return extraFields
}

func (fsCdr FSCdr) searchExtraField(field string, body map[string]interface{}) (result string) {
	for key, value := range body {
		if key == field {
			return utils.IfaceAsString(value)
		}
		switch v := value.(type) {
		case map[string]interface{}:
			if result = fsCdr.searchExtraField(field, v); len(result) != 0 {
				return
			}
		case []interface{}:
			for _, item := range v {
				if otherMap, ok := item.(map[string]interface{}); ok {
					if result = fsCdr.searchExtraField(field, otherMap); len(result) != 0 {
						return
					}
				}
			}
		}
	}
	return
}

// firstDefined will return first defined or search for dfltFld
func (fsCdr FSCdr) firstDefined(fldNames []string, dfltFld string) (val string) {
	var has bool
	for _, fldName := range fldNames {
		if val, has = fsCdr.vars[fldName]; has {
			return
		}
	}
	return fsCdr.searchExtraField(dfltFld, fsCdr.body)
}

func (fsCdr FSCdr) AsCDR(timezone string) (storCdr *CDR, err error) {
	storCdr = &CDR{
		CGRID:       fsCdr.getCGRID(),
		RunID:       fsCdr.vars["cgr_runid"],
		OriginHost:  utils.FirstNonEmpty(fsCdr.vars[utils.CGROriginHost], fsCdr.vars[FsIPv4]),
		Source:      FS_CDR_SOURCE,
		OriginID:    fsCdr.vars[FS_UUID],
		ToR:         utils.MetaVoice,
		RequestType: utils.FirstNonEmpty(fsCdr.vars[utils.CGRReqType], fsCdr.cgrCfg.GeneralCfg().DefaultReqType),
		Tenant:      utils.FirstNonEmpty(fsCdr.vars[utils.CGRTenant], fsCdr.cgrCfg.GeneralCfg().DefaultTenant),
		Category:    utils.FirstNonEmpty(fsCdr.vars[utils.CGRCategory], fsCdr.cgrCfg.GeneralCfg().DefaultCategory),
		Account:     fsCdr.firstDefined([]string{utils.CGRAccount, FS_USERNAME}, FsUsername),
		Subject:     fsCdr.firstDefined([]string{utils.CGRSubject, utils.CGRAccount, FS_USERNAME}, FsUsername),
		Destination: utils.FirstNonEmpty(fsCdr.vars[utils.CGRDestination], fsCdr.vars[FS_CALL_DEST_NR], fsCdr.vars[FS_SIP_REQUSER]),
		ExtraFields: fsCdr.getExtraFields(),
		ExtraInfo:   fsCdr.vars["cgr_extrainfo"],
		CostSource:  fsCdr.vars["cgr_costsource"],
		Cost:        -1,
	}
	if orderID, hasIt := fsCdr.vars["cgr_orderid"]; hasIt {
		if storCdr.OrderID, err = strconv.ParseInt(orderID, 10, 64); err != nil {
			return nil, err
		}
	}
	if setupTime, hasIt := fsCdr.vars[FS_SETUP_TIME]; hasIt {
		if storCdr.SetupTime, err = utils.ParseTimeDetectLayout(setupTime, timezone); err != nil {
			return nil, err
		} // Not interested to process errors, should do them if necessary in a previous step
	}
	if answerTime, hasIt := fsCdr.vars[FS_ANSWER_TIME]; hasIt {
		if storCdr.AnswerTime, err = utils.ParseTimeDetectLayout(answerTime, timezone); err != nil {
			return nil, err
		}
	}
	if usage, hasIt := fsCdr.vars[FS_DURATION]; hasIt {
		if storCdr.Usage, err = utils.ParseDurationWithSecs(usage); err != nil {
			return nil, err
		}
	}
	if partial, hasIt := fsCdr.vars["cgr_partial"]; hasIt {
		if storCdr.Partial, err = strconv.ParseBool(partial); err != nil {
			return nil, err
		}
	}
	if preRated, hasIt := fsCdr.vars["cgr_prerated"]; hasIt {
		if storCdr.PreRated, err = strconv.ParseBool(preRated); err != nil {
			return nil, err
		}
	}
	return
}
