/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package ees

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/esapi"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	elasticsearch "github.com/elastic/go-elasticsearch"
)

func NewElasticEE(cfg *config.EventExporterCfg, dc *utils.SafeMapStorage) (eEe *ElasticEE, err error) {
	eEe = &ElasticEE{
		cfg:  cfg,
		dc:   dc,
		reqs: newConcReq(cfg.ConcurrentRequests),
	}
	err = eEe.prepareOpts()
	return
}

// ElasticEE implements EventExporter interface for ElasticSearch export
type ElasticEE struct {
	cfg   *config.EventExporterCfg
	eClnt *elasticsearch.Client
	dc    *utils.SafeMapStorage
	opts  esapi.IndexRequest // this variable is used only for storing the options from OptsMap
	reqs  *concReq
	sync.RWMutex
	bytePreparing
}

// init will create all the necessary dependencies, including opening the file
func (eEe *ElasticEE) prepareOpts() (err error) {
	//parse opts
	eEe.opts.Index = utils.CDRsTBL
	if val, has := eEe.Cfg().Opts[utils.ElsIndex]; has {
		eEe.opts.Index = utils.IfaceAsString(val)
	}
	if val, has := eEe.Cfg().Opts[utils.ElsIfPrimaryTerm]; has {
		var intVal int64
		if intVal, err = utils.IfaceAsTInt64(val); err != nil {
			return
		}
		eEe.opts.IfPrimaryTerm = utils.IntPointer(int(intVal))
	}
	if val, has := eEe.Cfg().Opts[utils.ElsIfSeqNo]; has {
		var intVal int64
		if intVal, err = utils.IfaceAsTInt64(val); err != nil {
			return
		}
		eEe.opts.IfSeqNo = utils.IntPointer(int(intVal))
	}
	if val, has := eEe.Cfg().Opts[utils.ElsOpType]; has {
		eEe.opts.OpType = utils.IfaceAsString(val)
	}
	if val, has := eEe.Cfg().Opts[utils.ElsPipeline]; has {
		eEe.opts.Pipeline = utils.IfaceAsString(val)
	}
	if val, has := eEe.Cfg().Opts[utils.ElsRouting]; has {
		eEe.opts.Routing = utils.IfaceAsString(val)
	}
	if val, has := eEe.Cfg().Opts[utils.ElsTimeout]; has {
		if eEe.opts.Timeout, err = utils.IfaceAsDuration(val); err != nil {
			return
		}
	}
	if val, has := eEe.Cfg().Opts[utils.ElsVersionLow]; has {
		var intVal int64
		if intVal, err = utils.IfaceAsTInt64(val); err != nil {
			return
		}
		eEe.opts.Version = utils.IntPointer(int(intVal))
	}
	if val, has := eEe.Cfg().Opts[utils.ElsVersionType]; has {
		eEe.opts.VersionType = utils.IfaceAsString(val)
	}
	if val, has := eEe.Cfg().Opts[utils.ElsWaitForActiveShards]; has {
		eEe.opts.WaitForActiveShards = utils.IfaceAsString(val)
	}
	return
}

func (eEe *ElasticEE) Cfg() *config.EventExporterCfg { return eEe.cfg }

func (eEe *ElasticEE) Connect() (err error) {
	eEe.Lock()
	// create the client
	if eEe.eClnt == nil {
		eEe.eClnt, err = elasticsearch.NewClient(
			elasticsearch.Config{Addresses: strings.Split(eEe.Cfg().ExportPath, utils.InfieldSep)},
		)
	}
	eEe.Unlock()
	return
}

// ExportEvent implements EventExporter
func (eEe *ElasticEE) ExportEvent(ev interface{}, key string) (err error) {
	eEe.reqs.get()
	eEe.RLock()
	defer func() {
		eEe.RUnlock()
		eEe.reqs.done()
	}()
	if eEe.eClnt == nil {
		return utils.ErrDisconnected
	}
	eReq := esapi.IndexRequest{
		Index:               eEe.opts.Index,
		DocumentID:          key,
		Body:                bytes.NewReader(ev.([]byte)),
		Refresh:             "true",
		IfPrimaryTerm:       eEe.opts.IfPrimaryTerm,
		IfSeqNo:             eEe.opts.IfSeqNo,
		OpType:              eEe.opts.OpType,
		Parent:              eEe.opts.Parent,
		Pipeline:            eEe.opts.Pipeline,
		Routing:             eEe.opts.Routing,
		Timeout:             eEe.opts.Timeout,
		Version:             eEe.opts.Version,
		VersionType:         eEe.opts.VersionType,
		WaitForActiveShards: eEe.opts.WaitForActiveShards,
	}

	var resp *esapi.Response
	if resp, err = eReq.Do(context.Background(), eEe.eClnt); err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.IsError() {
		var e map[string]interface{}
		if err = json.NewDecoder(resp.Body).Decode(&e); err != nil {
			return
		}
		utils.Logger.Warning(fmt.Sprintf("<%s> Exporter with id: <%s> received error: <%+v> when indexing document",
			utils.EEs, eEe.Cfg().ID, e))
	}
	return
}

func (eEe *ElasticEE) Close() (_ error) {
	eEe.Lock()
	eEe.eClnt = nil
	eEe.Unlock()
	return
}

func (eEe *ElasticEE) GetMetrics() *utils.SafeMapStorage { return eEe.dc }
