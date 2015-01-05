/*
Real-time Charging System for Telecom & ISP environments
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

import (
	"github.com/cgrates/cgrates/utils"
	"reflect"
	"testing"
)

var cgrJsonCfg CgrJsonCfg

func TestNewCgrJsonCfgFromFile(t *testing.T) {
	var err error
	if cgrJsonCfg, err = NewCgrJsonCfgFromFile("cgrates_sample_cfg.json"); err != nil {
		t.Error(err.Error())
	}
}

func TestGeneralJsonCfg(t *testing.T) {
	eGCfg := &GeneralJsonCfg{
		Http_skip_tls_veify: utils.BoolPointer(false),
		Rounding_decimals:   utils.IntPointer(10),
		Dbdata_encoding:     utils.StringPointer("msgpack"),
		Tpexport_dir:        utils.StringPointer("/var/log/cgrates/tpe"),
		Default_reqtype:     utils.StringPointer("rated"),
		Default_category:    utils.StringPointer("call"),
		Default_tenant:      utils.StringPointer("cgrates.org"),
		Default_subject:     utils.StringPointer("cgrates")}
	if gCfg, err := cgrJsonCfg.GeneralJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eGCfg, gCfg) {
		t.Error("Received: ", gCfg)
	}
}

/*
func TestGetValInterface(t *testing.T) {
	if valIface, err := cgrJsonCfg.getValInterface("WRONG_SECTION", "", "default_reqtype"); err != nil {
		t.Error("Error: ", err)
	} else if valIface != nil {
		t.Error("Received: ", valIface)
	}
	if valIface, err := cgrJsonCfg.getValInterface(GENERAL_JSN, "", "default_reqtype"); err != nil {
		t.Error("Error: ", err)
	} else if valIface != interface{}("rated") {
		t.Error("Received: ", valIface)
	}
	expectContentFields := []interface{}{
		map[string]interface{}{
			"tag":       "CgrId",
			"type":      "cdrfield",
			"value":     "cgrid",
			"width":     40,
			"mandatory": true},
		map[string]interface{}{
			"tag":   "RunId",
			"type":  "cdrfield",
			"value": "mediation_runid",
			"width": 20},
		map[string]interface{}{
			"tag":   "Tor",
			"type":  "cdrfield",
			"value": "tor",
			"width": 6},
		map[string]interface{}{
			"tag":   "AccId",
			"type":  "cdrfield",
			"value": "accid",
			"width": 36},
		map[string]interface{}{
			"tag":   "ReqType",
			"type":  "cdrfield",
			"value": "reqtype",
			"width": 13},
		map[string]interface{}{
			"tag":   "Direction",
			"type":  "cdrfield",
			"value": "direction",
			"width": 4},
		map[string]interface{}{
			"tag":   "Tenant",
			"type":  "cdrfield",
			"value": "tenant",
			"width": 24},
		map[string]interface{}{
			"tag":   "Category",
			"type":  "cdrfield",
			"value": "category",
			"width": 10},
		map[string]interface{}{
			"tag":   "Account",
			"type":  "cdrfield",
			"value": "account",
			"width": 24},
		map[string]interface{}{
			"tag":   "Subject",
			"type":  "cdrfield",
			"value": "subject",
			"width": 24},
		map[string]interface{}{
			"tag":   "Destination",
			"type":  "cdrfield",
			"value": "destination",
			"width": 24},
		map[string]interface{}{
			"tag":    "SetupTime",
			"type":   "cdrfield",
			"value":  "setup_time",
			"layout": "2006-01-02T15:04:05Z07:00",
			"width":  30},
		map[string]interface{}{
			"tag":    "AnswerTime",
			"type":   "cdrfield",
			"value":  "answer_time",
			"layout": "2006-01-02T15:04:05Z07:00",
			"width":  30},
		map[string]interface{}{
			"tag":   "Usage",
			"type":  "cdrfield",
			"value": "usage",
			"width": 30},
		map[string]interface{}{
			"tag":   "Cost",
			"type":  "cdrfield",
			"value": "cost",
			"width": 24}}
	if valIface, err := cgrJsonCfg.getValInterface(CDRE_JSN, "CDRE-FW1", CONTENT_FIELDS_JSN); err != nil {
		t.Error("Error: ", err)
	} else if !reflect.DeepEqual(expectContentFields, valIface) {
		t.Errorf("Received: <%T>", valIface)
	}
}
*/
