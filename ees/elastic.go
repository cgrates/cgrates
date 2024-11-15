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

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
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
	if eEe.Cfg().Opts.Els.Index != nil {
		eEe.indxOpts.Index = *eEe.Cfg().Opts.Els.Index
	}
	eEe.indxOpts.IfPrimaryTerm = eEe.Cfg().Opts.Els.IfPrimaryTerm
	eEe.indxOpts.IfSeqNo = eEe.Cfg().Opts.Els.IfSeqNo
	if eEe.Cfg().Opts.Els.OpType != nil {
		eEe.indxOpts.OpType = *eEe.Cfg().Opts.Els.OpType
	}
	if eEe.Cfg().Opts.Els.Pipeline != nil {
		eEe.indxOpts.Pipeline = *eEe.Cfg().Opts.Els.Pipeline
	}
	if eEe.Cfg().Opts.Els.Routing != nil {
		eEe.indxOpts.Routing = *eEe.Cfg().Opts.Els.Routing
	}
	if eEe.Cfg().Opts.Els.Timeout != nil {
		eEe.indxOpts.Timeout = *eEe.Cfg().Opts.Els.Timeout
	}
	eEe.indxOpts.Version = eEe.Cfg().Opts.Els.Version
	if eEe.Cfg().Opts.Els.VersionType != nil {
		eEe.indxOpts.VersionType = *eEe.Cfg().Opts.Els.VersionType
	}
	if eEe.Cfg().Opts.Els.WaitForActiveShards != nil {
		eEe.indxOpts.WaitForActiveShards = *eEe.Cfg().Opts.Els.WaitForActiveShards
	}

	//client config
	if eEe.Cfg().Opts.Els.Cloud != nil && *eEe.Cfg().Opts.Els.Cloud {
		eEe.clnOpts.CloudID = eEe.Cfg().ExportPath
	} else {
		eEe.clnOpts.Addresses = strings.Split(eEe.Cfg().ExportPath, utils.InfieldSep)
	}
	if eEe.Cfg().Opts.Els.Username != nil {
		eEe.clnOpts.Username = *eEe.Cfg().Opts.Els.Username
	}
	if eEe.Cfg().Opts.Els.Password != nil {
		eEe.clnOpts.Password = *eEe.Cfg().Opts.Els.Password
	}
	if eEe.Cfg().Opts.Els.APIKey != nil {
		eEe.clnOpts.APIKey = *eEe.Cfg().Opts.Els.APIKey
	}
	if eEe.Cfg().Opts.RPC.CAPath != nil {
		var cacert []byte
		cacert, err = os.ReadFile(*eEe.Cfg().Opts.RPC.CAPath)
		if err != nil {
			return
		}
		eEe.clnOpts.CACert = cacert
	}
	if eEe.Cfg().Opts.Els.CertificateFingerprint != nil {
		eEe.clnOpts.CertificateFingerprint = *eEe.Cfg().Opts.Els.CertificateFingerprint
	}
	if eEe.Cfg().Opts.Els.ServiceToken != nil {
		eEe.clnOpts.ServiceToken = *eEe.Cfg().Opts.Els.ServiceToken
	}
	if eEe.Cfg().Opts.Els.DiscoverNodesOnStart != nil {
		eEe.clnOpts.DiscoverNodesOnStart = *eEe.Cfg().Opts.Els.DiscoverNodesOnStart
	}
	if eEe.Cfg().Opts.Els.DiscoverNodeInterval != nil {
		eEe.clnOpts.DiscoverNodesInterval = *eEe.Cfg().Opts.Els.DiscoverNodeInterval
	}
	if eEe.Cfg().Opts.Els.EnableDebugLogger != nil {
		eEe.clnOpts.EnableDebugLogger = *eEe.Cfg().Opts.Els.EnableDebugLogger
	}
	if loggerType := eEe.Cfg().Opts.Els.Logger; loggerType != nil {
		var logger elastictransport.Logger
		switch *loggerType {
		case utils.ElsJson:
			logger = &elastictransport.JSONLogger{Output: os.Stdout, EnableRequestBody: true, EnableResponseBody: true}
		case utils.ElsColor:
			logger = &elastictransport.ColorLogger{Output: os.Stdout, EnableRequestBody: true, EnableResponseBody: true}
		case utils.ElsText:
			logger = &elastictransport.TextLogger{Output: os.Stdout, EnableRequestBody: true, EnableResponseBody: true}
		default:
			return
		}
		eEe.clnOpts.Logger = logger
	}
	if eEe.Cfg().Opts.Els.CompressRequestBody != nil {
		eEe.clnOpts.CompressRequestBody = *eEe.Cfg().Opts.Els.CompressRequestBody
	}
	if eEe.Cfg().Opts.Els.RetryOnStatus != nil {
		eEe.clnOpts.RetryOnStatus = *eEe.Cfg().Opts.Els.RetryOnStatus
	}
	if eEe.Cfg().Opts.Els.MaxRetries != nil {
		eEe.clnOpts.MaxRetries = *eEe.Cfg().Opts.Els.MaxRetries
	}
	if eEe.Cfg().Opts.Els.DisableRetry != nil {
		eEe.clnOpts.DisableRetry = *eEe.Cfg().Opts.Els.DisableRetry
	}
	if eEe.Cfg().Opts.Els.CompressRequestBodyLevel != nil {
		eEe.clnOpts.CompressRequestBodyLevel = *eEe.Cfg().Opts.Els.CompressRequestBodyLevel
	}
	return
}

func (eEe *ElasticEE) Cfg() *config.EventExporterCfg { return eEe.cfg }

func (eEe *ElasticEE) Connect() (err error) {
	eEe.Lock()
	defer eEe.Unlock()
	// create the client
	if eEe.eClnt != nil {
		return
	}
	eEe.eClnt, err = elasticsearch.NewClient(eEe.clnOpts)
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
