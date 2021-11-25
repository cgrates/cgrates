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
	if eEe.Cfg().Opts.ElsIndex != nil {
		eEe.opts.Index = *eEe.Cfg().Opts.ElsIndex
	}
	eEe.opts.IfPrimaryTerm = eEe.Cfg().Opts.ElsIfPrimaryTerm
	eEe.opts.IfSeqNo = eEe.Cfg().Opts.ElsIfSeqNo
	if eEe.Cfg().Opts.ElsOpType != nil {
		eEe.opts.OpType = *eEe.Cfg().Opts.ElsOpType
	}
	if eEe.Cfg().Opts.ElsPipeline != nil {
		eEe.opts.Pipeline = *eEe.Cfg().Opts.ElsPipeline
	}
	if eEe.Cfg().Opts.ElsRouting != nil {
		eEe.opts.Routing = *eEe.Cfg().Opts.ElsRouting
	}
	if eEe.Cfg().Opts.ElsTimeout != nil {
		eEe.opts.Timeout = *eEe.Cfg().Opts.ElsTimeout
	}
	eEe.opts.Version = eEe.Cfg().Opts.ElsVersion
	if eEe.Cfg().Opts.ElsVersionType != nil {
		eEe.opts.VersionType = *eEe.Cfg().Opts.ElsVersionType
	}
	if eEe.Cfg().Opts.ElsWaitForActiveShards != nil {
		eEe.opts.WaitForActiveShards = *eEe.Cfg().Opts.ElsWaitForActiveShards
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
