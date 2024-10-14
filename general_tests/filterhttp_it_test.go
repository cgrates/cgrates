//go:build integration
// +build integration

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

package general_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestFilterHTTP(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

		"data_db": {								
			"db_type": "*internal"
		},
		
		"stor_db": {
			"db_type": "*internal"
		},
		
		"attributes":{
			"enabled": true,
			"indexed_selects": false,
		},
		
		"rals": {
			"enabled": true,
		},
		
		"cdrs": {
			"enabled": true,
			"attributes_conns": ["*internal"],
			"rals_conns": ["*internal"]
		},
		
		"schedulers": {
			"enabled": true
		},
		
		"apiers": {
			"enabled": true,
			"scheduler_conns": ["*internal"]
		}
		
		}`

	type event struct {
		Req  map[string]any `json:"*req"`
		Opts map[string]any `json:"*opts"`
	}

	expEv := event{
		Req: map[string]any{
			utils.EventName:    "VariableTest",
			utils.ToR:          utils.MetaVoice,
			utils.AccountField: "1001",
		},
		Opts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/filters", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		var ev event

		if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		equal := reflect.DeepEqual(ev, expEv)

		fmt.Fprint(w, equal)
	})

	mux.HandleFunc("/attributes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		var ev event

		if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		equal := reflect.DeepEqual(ev, expEv)
		if !equal {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		fmt.Fprint(w, "Customer")

	})

	mux.HandleFunc("/filter", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		supplDest := []string{"1001", "1002", "1003", "1004"}
		str := r.URL.Query().Get("~*req.Destination")
		for _, sup := range supplDest {
			if sup == str {
				fmt.Fprint(w, "true")
				return
			}
		}
		fmt.Fprint(w, "false")

	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	tpFiles := map[string]string{
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4],ActivationInterval[5]
cgrates.org,FLTR_HTTP,*http#[` + srv.URL + `/filters],*any,,2023-07-29T15:00:00Z
cgrates.org,FLTR_DEST,*http#[` + srv.URL + `/filter],~*req.Destination,,2022-07-29T15:00:00Z
cgrates.org,FLTR_DST_1002,*string,~*req.Destination,1002,`,
		utils.AttributesCsv: `#Tenant,ID,Context,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ATTR_ACNT_1001,*any,FLTR_HTTP,,,*req.OfficeGroup,*http#[` + srv.URL + `/attributes],*attributes,false,10
cgrates.org,ATTR_DEST,*any,FLTR_DST_1002;FLTR_DEST,,,*req.Supplier,*constant,Supplier1,,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)

	t.Run("FilterHTTPFullEvent", func(t *testing.T) {

		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEvent",
			Event: map[string]any{
				utils.EventName:    "VariableTest",
				utils.ToR:          utils.MetaVoice,
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		}
		eRply := engine.AttrSProcessEventReply{
			MatchedProfiles: []string{"cgrates.org:ATTR_ACNT_1001"},
			AlteredFields:   []string{"*req.OfficeGroup"},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEvent",
				Event: map[string]any{
					utils.EventName:    "VariableTest",
					utils.AccountField: "1001",
					utils.ToR:          utils.MetaVoice,
					"OfficeGroup":      "Customer",
				},
				APIOpts: map[string]any{
					utils.OptsContext: utils.MetaSessionS,
				},
			},
		}

		var rplyEv engine.AttrSProcessEventReply
		if err := client.Call(context.Background(), utils.AttributeSv1ProcessEvent,
			ev, &rplyEv); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(eRply, rplyEv) {
			t.Errorf("Expecting: %s, received: %s",
				utils.ToJSON(eRply), utils.ToJSON(rplyEv))
		}
	})

	t.Run("FilterHTTPField", func(t *testing.T) {

		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeFilterHTTP",
			Event: map[string]any{
				"EventName":   "AddDestinationDetails",
				"Destination": "1002",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		}

		eRply := engine.AttrSProcessEventReply{
			MatchedProfiles: []string{"cgrates.org:ATTR_DEST"},
			AlteredFields:   []string{"*req.Supplier"},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeFilterHTTP",
				Event: map[string]any{
					"EventName":   "AddDestinationDetails",
					"Destination": "1002",
					"Supplier":    "Supplier1",
				},
				APIOpts: map[string]any{
					utils.OptsContext: utils.MetaSessionS,
				},
			},
		}

		var rplyEv engine.AttrSProcessEventReply
		if err := client.Call(context.Background(), utils.AttributeSv1ProcessEvent,
			ev, &rplyEv); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(eRply, rplyEv) {
			t.Errorf("Expecting: %+v, received: %+v",
				utils.ToJSON(eRply), utils.ToJSON(rplyEv))
		}

	})

}
