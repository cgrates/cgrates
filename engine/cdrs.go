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
func storeAndMediate(storedCdr *StoredCdr) error {
	if !cfg.CDRSStoreDisable {
		if err := storage.SetCdr(storedCdr); err != nil {
			return err
		}
	}
	if stats != nil {
		go func(storedCdr *StoredCdr) {
			if err := stats.AppendCDR(storedCdr, nil); err != nil {
				Logger.Err(fmt.Sprintf("<CDRS> Could not append cdr to stats: %s", err.Error()))
			}
		}(storedCdr)
	}
	if cfg.CDRSCdrReplication != nil {
		replicateCdr(storedCdr, cfg.CDRSCdrReplication)
	}
	if cfg.CDRSMediator == utils.INTERNAL {
		go func(storedCdr *StoredCdr) {
			if err := medi.RateCdr(storedCdr, true); err != nil {
				Logger.Err(fmt.Sprintf("<CDRS> Could not run mediation on CDR: %s", err.Error()))
			}
		}(storedCdr)
	}
	return nil
}

// ToDo: Add websocket support
func replicateCdr(cdr *StoredCdr, replCfgs []*config.CdrReplicationCfg) error {
	for _, rplCfg := range replCfgs {
		switch rplCfg.Transport {
		case utils.META_HTTP_POST:
			httpClient := new(http.Client)
			errChan := make(chan error)
			go func(cdr *StoredCdr, rplCfg *config.CdrReplicationCfg, errChan chan error) {
				if _, err := httpClient.PostForm(fmt.Sprintf("http://%s/cdr_post", rplCfg.Server), cdr.AsHttpForm()); err != nil {
					Logger.Err(fmt.Sprintf("<CDRReplicator> Replicating CDR: %+v, got error: %s", cdr, err.Error()))
					errChan <- err
				}
				errChan <- nil
			}(cdr, rplCfg, errChan)
			if rplCfg.Synchronous { // Synchronize here
				<-errChan
			}
		}
	}
	return nil
}

// Handler for generic cgr cdr http
func cgrCdrHandler(w http.ResponseWriter, r *http.Request) {
	cgrCdr, err := NewCgrCdrFromHttpReq(r)
	if err != nil {
		Logger.Err(fmt.Sprintf("<CDRS> Could not create CDR entry: %s", err.Error()))
	}
	if err := storeAndMediate(cgrCdr.AsStoredCdr()); err != nil {
		Logger.Err(fmt.Sprintf("<CDRS> Errors when storing CDR entry: %s", err.Error()))
	}
}

// Handler for fs http
func fsCdrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	fsCdr, err := NewFSCdr(body)
	if err != nil {
		Logger.Err(fmt.Sprintf("<CDRS> Could not create CDR entry: %s", err.Error()))
	}
	if err := storeAndMediate(fsCdr.AsStoredCdr()); err != nil {
		Logger.Err(fmt.Sprintf("<CDRS> Errors when storing CDR entry: %s", err.Error()))
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
				Logger.Err(fmt.Sprintf("<CDRS> Errors connecting to CDRS stats service : %s", err.Error()))
			}
		}
	} else {
		// disable stats for cdrs
		stats = nil
	}
	return &CDRS{}
}

func (cdrs *CDRS) RegisterHanlersToServer(server *Server) {
	server.RegisterHttpFunc("/cdr_post", cgrCdrHandler)
	server.RegisterHttpFunc("/freeswitch_json", fsCdrHandler)
}

// Used to internally process CDR
func (cdrs *CDRS) ProcessCdr(cdr *StoredCdr) error {
	return storeAndMediate(cdr)
}
