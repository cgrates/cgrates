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
	"io/ioutil"
	"net/http"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/mediator"
	"github.com/cgrates/cgrates/utils"
)

var (
	cfg     *config.CGRConfig // Share the configuration with the rest of the package
	storage engine.CdrStorage
	medi    *mediator.Mediator
)

// Returns error if not able to properly store the CDR, mediation is async since we can always recover offline
func storeAndMediate(storedCdr *utils.StoredCdr) error {
	if err := storage.SetCdr(storedCdr); err != nil {
		return err
	}
	if cfg.CDRSMediator == utils.INTERNAL {
		go func() {
			if err := medi.RateCdr(storedCdr); err != nil {
				engine.Logger.Err(fmt.Sprintf("Could not run mediation on CDR: %s", err.Error()))
			}
		}()
	}
	return nil
}

// Handler for generic cgr cdr http
func cgrCdrHandler(w http.ResponseWriter, r *http.Request) {
	cgrCdr, err := utils.NewCgrCdrFromHttpReq(r)
	if err != nil {
		engine.Logger.Err(fmt.Sprintf("Could not create CDR entry: %s", err.Error()))
	}
	if err := storeAndMediate(cgrCdr.AsStoredCdr()); err != nil {
		engine.Logger.Err(fmt.Sprintf("Errors when storing CDR entry: %s", err.Error()))
	}
}

// Handler for fs http
func fsCdrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	fsCdr, err := NewFSCdr(body)
	if err != nil {
		engine.Logger.Err(fmt.Sprintf("Could not create CDR entry: %s", err.Error()))
	}
	if err := storeAndMediate(fsCdr.AsStoredCdr()); err != nil {
		engine.Logger.Err(fmt.Sprintf("Errors when storing CDR entry: %s", err.Error()))
	}
}

type CDRS struct{}

func New(s engine.CdrStorage, m *mediator.Mediator, c *config.CGRConfig) *CDRS {
	storage = s
	medi = m
	cfg = c
	return &CDRS{}
}

func (cdrs *CDRS) RegisterHanlersToServer(server *engine.Server) {
	server.RegisterHttpFunc("/cgr", cgrCdrHandler)
	server.RegisterHttpFunc("/freeswitch_json", fsCdrHandler)
}

// Used to internally process CDR
func (cdrs *CDRS) ProcessCdr(cdr *utils.StoredCdr) error {
	return storeAndMediate(cdr)
}
