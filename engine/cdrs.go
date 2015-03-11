/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	cfg     *config.CGRConfig // Share the configuration with the rest of the package
	storage CdrStorage
	medi    *Mediator
	stats   StatsInterface
)

// Returns error if not able to properly store the CDR, mediation is async since we can always recover offline
func storeAndMediate(storedCdr *utils.StoredCdr) error {
	if !cfg.CDRSStoreDisable {
		if err := storage.SetCdr(storedCdr); err != nil {
			return err
		}
	}
	if stats != nil {
		go func() {
			if err := stats.AppendCDR(storedCdr, nil); err != nil {
				Logger.Err(fmt.Sprintf("Could not append cdr to stats: %s", err.Error()))
			}
		}()
	}
	if cfg.CDRSMediator == utils.INTERNAL {
		go func() {
			if err := medi.RateCdr(storedCdr, true); err != nil {
				Logger.Err(fmt.Sprintf("Could not run mediation on CDR: %s", err.Error()))
			}
		}()
	}
	return nil
}

// Handler for generic cgr cdr http
func cgrCdrHandler(w http.ResponseWriter, r *http.Request) {
	cgrCdr, err := utils.NewCgrCdrFromHttpReq(r)
	if err != nil {
		Logger.Err(fmt.Sprintf("Could not create CDR entry: %s", err.Error()))
	}
	if err := storeAndMediate(cgrCdr.AsStoredCdr()); err != nil {
		Logger.Err(fmt.Sprintf("Errors when storing CDR entry: %s", err.Error()))
	}
}

// Handler for fs http
func fsCdrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	fsCdr, err := NewFSCdr(body)
	if err != nil {
		Logger.Err(fmt.Sprintf("Could not create CDR entry: %s", err.Error()))
	}
	if err := storeAndMediate(fsCdr.AsStoredCdr()); err != nil {
		Logger.Err(fmt.Sprintf("Errors when storing CDR entry: %s", err.Error()))
	}
}

type CDRS struct{}

func NewCdrS(s CdrStorage, m *Mediator, st *Stats, c *config.CGRConfig) *CDRS {
	storage = s
	medi = m
	cfg = c
	stats = st
	if cfg.CDRSStats != "" {
		if cfg.CDRSStats != utils.INTERNAL {
			if s, err := NewProxyStats(cfg.CDRSStats); err == nil {
				stats = s
			} else {
				Logger.Err(fmt.Sprintf("Errors connecting to CDRS stats service : %s", err.Error()))
			}
		}
	} else {
		// disable stats for cdrs
		stats = nil
	}
	return &CDRS{}
}

func (cdrs *CDRS) RegisterHanlersToServer(server *Server) {
	server.RegisterHttpFunc("/cgr", cgrCdrHandler)
	server.RegisterHttpFunc("/freeswitch_json", fsCdrHandler)
}

// Used to internally process CDR
func (cdrs *CDRS) ProcessCdr(cdr *utils.StoredCdr) error {
	return storeAndMediate(cdr)
}
