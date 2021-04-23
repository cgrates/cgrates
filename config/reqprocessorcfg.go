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

package config

import "github.com/cgrates/cgrates/utils"

// RequestProcessor is the request processor configuration
type RequestProcessor struct {
	ID            string
	Tenant        RSRParsers
	Filters       []string
	Flags         utils.FlagsWithParams
	Timezone      string
	RequestFields []*FCTemplate
	ReplyFields   []*FCTemplate
}

func (rp *RequestProcessor) loadFromJSONCfg(jsnCfg *ReqProcessorJsnCfg, sep string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.ID != nil {
		rp.ID = *jsnCfg.ID
	}
	if jsnCfg.Filters != nil {
		rp.Filters = utils.CloneStringSlice(*jsnCfg.Filters)
	}
	if jsnCfg.Flags != nil {
		rp.Flags = utils.FlagsWithParamsFromSlice(*jsnCfg.Flags)
	}
	if jsnCfg.Timezone != nil {
		rp.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Tenant != nil {
		if rp.Tenant, err = NewRSRParsers(*jsnCfg.Tenant, sep); err != nil {
			return
		}
	}
	if jsnCfg.Request_fields != nil {
		if rp.RequestFields, err = FCTemplatesFromFCTemplatesJSONCfg(*jsnCfg.Request_fields, sep); err != nil {
			return
		}
	}
	if jsnCfg.Reply_fields != nil {
		if rp.ReplyFields, err = FCTemplatesFromFCTemplatesJSONCfg(*jsnCfg.Reply_fields, sep); err != nil {
			return
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (rp *RequestProcessor) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.IDCfg:       rp.ID,
		utils.FiltersCfg:  utils.CloneStringSlice(rp.Filters),
		utils.FlagsCfg:    rp.Flags.SliceFlags(),
		utils.TimezoneCfg: rp.Timezone,
	}
	if rp.Tenant != nil {
		initialMP[utils.TenantCfg] = rp.Tenant.GetRule(separator)
	}
	if rp.RequestFields != nil {
		requestFields := make([]map[string]interface{}, len(rp.RequestFields))
		for i, item := range rp.RequestFields {
			requestFields[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.RequestFieldsCfg] = requestFields
	}
	if rp.ReplyFields != nil {
		replyFields := make([]map[string]interface{}, len(rp.ReplyFields))
		for i, item := range rp.ReplyFields {
			replyFields[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.ReplyFieldsCfg] = replyFields
	}
	return
}

// Clone returns a deep copy of APIBanCfg
func (rp RequestProcessor) Clone() (cln *RequestProcessor) {
	cln = &RequestProcessor{
		ID:       rp.ID,
		Tenant:   rp.Tenant.Clone(),
		Flags:    rp.Flags.Clone(),
		Timezone: rp.Timezone,
	}
	if rp.Filters != nil {
		cln.Filters = utils.CloneStringSlice(rp.Filters)
	}
	if rp.RequestFields != nil {
		cln.RequestFields = make([]*FCTemplate, len(rp.RequestFields))
		for i, rf := range rp.RequestFields {
			cln.RequestFields[i] = rf.Clone()
		}
	}
	if rp.ReplyFields != nil {
		cln.ReplyFields = make([]*FCTemplate, len(rp.ReplyFields))
		for i, rf := range rp.ReplyFields {
			cln.ReplyFields[i] = rf.Clone()
		}
	}
	return
}

type ReqProcessorJsnCfg struct {
	ID             *string
	Filters        *[]string
	Tenant         *string
	Timezone       *string
	Flags          *[]string
	Request_fields *[]*FcTemplateJsonCfg
	Reply_fields   *[]*FcTemplateJsonCfg
}

func diffReqProcessorJsnCfg(d *ReqProcessorJsnCfg, v1, v2 *RequestProcessor, separator string) *ReqProcessorJsnCfg {
	if d == nil {
		d = new(ReqProcessorJsnCfg)
	}
	if v1.ID != v2.ID {
		d.ID = utils.StringPointer(v2.ID)
	}
	tnt1 := v1.Tenant.GetRule(separator)
	tnt2 := v2.Tenant.GetRule(separator)
	if tnt1 != tnt2 {
		d.Tenant = utils.StringPointer(tnt2)
	}
	if !utils.SliceStringEqual(v1.Filters, v2.Filters) {
		d.Filters = utils.SliceStringPointer(utils.CloneStringSlice(v2.Filters))
	}
	flag1 := v1.Flags.SliceFlags()
	flag2 := v2.Flags.SliceFlags()
	if !utils.SliceStringEqual(flag1, flag2) {
		d.Flags = utils.SliceStringPointer(flag2)
	}
	if v1.Timezone != v2.Timezone {
		d.Timezone = utils.StringPointer(v2.Timezone)
	}
	var req []*FcTemplateJsonCfg
	if d.Request_fields != nil {
		req = *d.Request_fields
	}
	req = diffFcTemplateJsonCfg(req, v1.RequestFields, v2.RequestFields, separator)
	d.Request_fields = &req

	var rply []*FcTemplateJsonCfg
	if d.Reply_fields != nil {
		rply = *d.Reply_fields
	}
	rply = diffFcTemplateJsonCfg(rply, v1.ReplyFields, v2.ReplyFields, separator)
	d.Reply_fields = &rply
	return d
}
func getReqProcessorJsnCfg(d []*ReqProcessorJsnCfg, id string) (*ReqProcessorJsnCfg, int) {
	for i, v := range d {
		if v.ID != nil && *v.ID == id {
			return v, i
		}
	}
	return nil, -1
}

func getRequestProcessor(d []*RequestProcessor, id string) *RequestProcessor {
	for _, v := range d {
		if v.ID == id {
			return v
		}
	}
	return new(RequestProcessor)
}

func diffReqProcessorsJsnCfg(d *[]*ReqProcessorJsnCfg, v1, v2 []*RequestProcessor, separator string) *[]*ReqProcessorJsnCfg {
	if d == nil || *d == nil {
		d = &[]*ReqProcessorJsnCfg{}
	}
	for _, val := range v2 {
		dv, i := getReqProcessorJsnCfg(*d, val.ID)
		dv = diffReqProcessorJsnCfg(dv, getRequestProcessor(v1, val.ID), val, separator)
		if i == -1 {
			*d = append(*d, dv)
		} else {
			(*d)[i] = dv
		}
	}
	return d
}

func appendRequestProcessors(to []*RequestProcessor, from *[]*ReqProcessorJsnCfg, separator string) (_ []*RequestProcessor, err error) {
	if from == nil {
		return to, nil
	}
	for _, rpJsn := range *from {
		rp := new(RequestProcessor)
		var haveID bool
		if rpJsn.ID != nil {
			for _, rpTo := range to {
				if rpTo.ID == *rpJsn.ID {
					rp = rpTo // Will load data into the one set
					haveID = true
					break
				}
			}
		}
		if err = rp.loadFromJSONCfg(rpJsn, separator); err != nil {
			return
		}
		if !haveID {
			to = append(to, rp)
		}
	}
	return to, nil
}

func equalsRequestProcessors(v1, v2 []*RequestProcessor) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v2 {
		if v1[i].ID != v2[i].ID ||
			!utils.SliceStringEqual(v1[i].Tenant.AsStringSlice(), v2[i].Tenant.AsStringSlice()) ||
			!utils.SliceStringEqual(v1[i].Filters, v2[i].Filters) ||
			!utils.SliceStringEqual(v1[i].Flags.SliceFlags(), v2[i].Flags.SliceFlags()) ||
			v1[i].Timezone != v2[i].Timezone ||
			!fcTemplatesEqual(v1[i].RequestFields, v2[i].RequestFields) ||
			!fcTemplatesEqual(v1[i].ReplyFields, v2[i].ReplyFields) {
			return false
		}
	}
	return true
}
