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
	"os"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
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
	cfg      *config.EventExporterCfg
	eClnt    *elasticsearch.Client
	dc       *utils.SafeMapStorage
	indxOpts esapi.IndexRequest // this variable is used only for storing the options from OptsMap
	reqs     *concReq
	clnOpts  elasticsearch.Config
	sync.RWMutex
	bytePreparing
}

// init will create all the necessary dependencies, including opening the file
func (eEe *ElasticEE) prepareOpts() (err error) {
	//parse opts
	eEe.indxOpts.Index = utils.CDRsTBL
	if eEe.Cfg().Opts.ElsIndex != nil {
		eEe.indxOpts.Index = *eEe.Cfg().Opts.ElsIndex
	}
	eEe.indxOpts.IfPrimaryTerm = eEe.Cfg().Opts.ElsIfPrimaryTerm
	eEe.indxOpts.IfSeqNo = eEe.Cfg().Opts.ElsIfSeqNo
	if eEe.Cfg().Opts.ElsOpType != nil {
		eEe.indxOpts.OpType = *eEe.Cfg().Opts.ElsOpType
	}
	if eEe.Cfg().Opts.ElsPipeline != nil {
		eEe.indxOpts.Pipeline = *eEe.Cfg().Opts.ElsPipeline
	}
	if eEe.Cfg().Opts.ElsRouting != nil {
		eEe.indxOpts.Routing = *eEe.Cfg().Opts.ElsRouting
	}
	if eEe.Cfg().Opts.ElsTimeout != nil {
		eEe.indxOpts.Timeout = *eEe.Cfg().Opts.ElsTimeout
	}
	eEe.indxOpts.Version = eEe.Cfg().Opts.ElsVersion
	if eEe.Cfg().Opts.ElsVersionType != nil {
		eEe.indxOpts.VersionType = *eEe.Cfg().Opts.ElsVersionType
	}
	if eEe.Cfg().Opts.ElsWaitForActiveShards != nil {
		eEe.indxOpts.WaitForActiveShards = *eEe.Cfg().Opts.ElsWaitForActiveShards
	}

	//client config
	if eEe.Cfg().Opts.ElsCloud != nil && *eEe.Cfg().Opts.ElsCloud {
		eEe.clnOpts.CloudID = eEe.Cfg().ExportPath
	} else {
		eEe.clnOpts.Addresses = strings.Split(eEe.Cfg().ExportPath, utils.InfieldSep)
	}
	if eEe.Cfg().Opts.ElsUsername != nil {
		eEe.clnOpts.Username = *eEe.Cfg().Opts.ElsUsername
	}
	if eEe.Cfg().Opts.ElsPassword != nil {
		eEe.clnOpts.Password = *eEe.Cfg().Opts.ElsPassword
	}
	if eEe.Cfg().Opts.ElsAPIKey != nil {
		eEe.clnOpts.APIKey = *eEe.Cfg().Opts.ElsAPIKey
	}
	if eEe.Cfg().Opts.ElsCACert != nil {
		var cacert []byte
		cacert, err = os.ReadFile(*eEe.Cfg().Opts.ElsCACert)
		if err != nil {
			return
		}
		eEe.clnOpts.CACert = cacert
	}
	if eEe.Cfg().Opts.ElsCertificateFingerprint != nil {
		eEe.clnOpts.CertificateFingerprint = *eEe.Cfg().Opts.ElsCertificateFingerprint
	}
	if eEe.Cfg().Opts.ElsServiceToken != nil {
		eEe.clnOpts.ServiceToken = *eEe.Cfg().Opts.ElsServiceToken
	}
	if eEe.Cfg().Opts.ElsDiscoverNodesOnStart != nil {
		eEe.clnOpts.DiscoverNodesOnStart = *eEe.Cfg().Opts.ElsDiscoverNodesOnStart
	}
	if eEe.Cfg().Opts.ElsDiscoverNodeInterval != nil {
		eEe.clnOpts.DiscoverNodesInterval = *eEe.Cfg().Opts.ElsDiscoverNodeInterval
	}
	if eEe.Cfg().Opts.ElsEnableDebugLogger != nil {
		eEe.clnOpts.EnableDebugLogger = *eEe.Cfg().Opts.ElsEnableDebugLogger
	}
	if eEe.Cfg().Opts.ElsCompressRequestBody != nil {
		eEe.clnOpts.CompressRequestBody = *eEe.Cfg().Opts.ElsCompressRequestBody
	}
	if eEe.Cfg().Opts.ElsRetryOnStatus != nil {
		eEe.clnOpts.RetryOnStatus = *eEe.Cfg().Opts.ElsRetryOnStatus
	}
	return
}

func (eEe *ElasticEE) Cfg() *config.EventExporterCfg { return eEe.cfg }

func (eEe *ElasticEE) Connect() (err error) {
	eEe.Lock()
	// create the client
	if eEe.eClnt != nil {
		return
	}
	eEe.eClnt, err = elasticsearch.NewClient(eEe.clnOpts)
	eEe.Unlock()
	return
}

// ExportEvent implements EventExporter
func (eEe *ElasticEE) ExportEvent(ev any, key string) (err error) {
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
		Index:               eEe.indxOpts.Index,
		DocumentID:          key,
		Body:                bytes.NewReader(ev.([]byte)),
		Refresh:             "true",
		IfPrimaryTerm:       eEe.indxOpts.IfPrimaryTerm,
		IfSeqNo:             eEe.indxOpts.IfSeqNo,
		OpType:              eEe.indxOpts.OpType,
		Pipeline:            eEe.indxOpts.Pipeline,
		Routing:             eEe.indxOpts.Routing,
		Timeout:             eEe.indxOpts.Timeout,
		Version:             eEe.indxOpts.Version,
		VersionType:         eEe.indxOpts.VersionType,
		WaitForActiveShards: eEe.indxOpts.WaitForActiveShards,
	}

	var resp *esapi.Response
	if resp, err = eReq.Do(context.Background(), eEe.eClnt); err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.IsError() {
		var e map[string]any
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
