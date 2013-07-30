/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"fmt"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/mediator"
	"io/ioutil"
	"net/http"
)

var (
	cfg     *config.CGRConfig // Share the configuration with the rest of the package
	storage engine.DataStorage
	medi    *mediator.Mediator
)

func fsCdrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	if fsCdr, err := new(FSCdr).New(body); err == nil {
		storage.SetCdr(fsCdr)
		if cfg.CDRSMediator == "internal" {
			errMedi := medi.MediateDBCDR(fsCdr, storage)
			if errMedi != nil {
				engine.Logger.Err(fmt.Sprintf("Could not run mediation on CDR: %s", errMedi.Error()))
			}
		} else {
			//TODO: use the connection to mediator
		}
	} else {
		engine.Logger.Err(fmt.Sprintf("Could not create CDR entry: %v", err))
	}
}

func genCdrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	if genCdr, err := new(GenCdr).New(body); err == nil {
		storage.SetCdr(genCdr)
		if cfg.CDRSMediator == "internal" {
			errMedi := medi.MediateDBCDR(genCdr, storage)
			if errMedi != nil {
				engine.Logger.Err(fmt.Sprintf("Could not run mediation on CDR: %s", errMedi.Error()))
			}
		} else {
			//TODO: use the connection to mediator
		}
	} else {
		engine.Logger.Err(fmt.Sprintf("Could not create CDR entry: %v", err))
	}
}

type CDRS struct{}

func New(s engine.DataStorage, m *mediator.Mediator, c *config.CGRConfig) *CDRS {
	storage = s
	medi = m
	cfg = c
	return &CDRS{}
}

func (cdrs *CDRS) StartCapturingCDRs() {
	if cfg.CDRSfsJSONEnabled {
		http.HandleFunc("/freeswitch_json", fsCdrHandler)
	}
	if cfg.CDRSgenJSONEnabled {
		http.HandleFunc("/generic_json", genCdrHandler)
	}
	http.ListenAndServe(cfg.CDRSListen, nil)
}
