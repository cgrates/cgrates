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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
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
	opts     esapi.IndexRequest // this variable is used only for storing the options from OptsMap
	clntOpts elasticsearch.Config
	reqs     *concReq
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

	//client opts
	if eEe.Cfg().Opts.ElsCloud != nil && *eEe.Cfg().Opts.ElsCloud {
		eEe.clntOpts.CloudID = eEe.Cfg().ExportPath
	} else {
		eEe.clntOpts.Addresses = strings.Split(eEe.Cfg().ExportPath, utils.InfieldSep)
	}
	if eEe.Cfg().Opts.ElsUsername != nil {
		eEe.clntOpts.Username = *eEe.Cfg().Opts.ElsUsername
	}
	if eEe.Cfg().Opts.ElsPassword != nil {
		eEe.clntOpts.Password = *eEe.Cfg().Opts.ElsPassword
	}
	if eEe.Cfg().Opts.ElsAPIKey != nil {
		eEe.clntOpts.APIKey = *eEe.Cfg().Opts.ElsAPIKey
	}
	if eEe.Cfg().Opts.CAPath != nil {
		var cacert []byte
		cacert, err = os.ReadFile(*eEe.Cfg().Opts.CAPath)
		if err != nil {
			return
		}
		eEe.clntOpts.CACert = cacert
	}
	if eEe.Cfg().Opts.ElsCertificateFingerprint != nil {
		eEe.clntOpts.CertificateFingerprint = *eEe.Cfg().Opts.ElsCertificateFingerprint
	}
	if eEe.Cfg().Opts.ElsServiceToken != nil {
		eEe.clntOpts.ServiceToken = *eEe.Cfg().Opts.ElsServiceToken
	}
	if eEe.Cfg().Opts.ElsDiscoverNodesOnStart != nil {
		eEe.clntOpts.DiscoverNodesOnStart = *eEe.Cfg().Opts.ElsDiscoverNodesOnStart
	}
	if eEe.Cfg().Opts.ElsDiscoverNodeInterval != nil {
		eEe.clntOpts.DiscoverNodesInterval = *eEe.Cfg().Opts.ElsDiscoverNodeInterval
	}
	if eEe.Cfg().Opts.ElsEnableDebugLogger != nil {
		eEe.clntOpts.EnableDebugLogger = *eEe.Cfg().Opts.ElsEnableDebugLogger
	}
	if loggerType := eEe.Cfg().Opts.ElsLogger; loggerType != nil {
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
		eEe.clntOpts.Logger = logger
	}
	if eEe.Cfg().Opts.ElsCompressRequestBody != nil {
		eEe.clntOpts.CompressRequestBody = *eEe.Cfg().Opts.ElsCompressRequestBody
	}
	if eEe.Cfg().Opts.ElsRetryOnStatus != nil {
		eEe.clntOpts.RetryOnStatus = *eEe.Cfg().Opts.ElsRetryOnStatus
	}
	if eEe.Cfg().Opts.ElsMaxRetries != nil {
		eEe.clntOpts.MaxRetries = *eEe.Cfg().Opts.ElsMaxRetries
	}
	if eEe.Cfg().Opts.ElsDisableRetry != nil {
		eEe.clntOpts.DisableRetry = *eEe.Cfg().Opts.ElsDisableRetry
	}
	if eEe.Cfg().Opts.ElsCompressRequestBodyLevel != nil {
		eEe.clntOpts.CompressRequestBodyLevel = *eEe.Cfg().Opts.ElsCompressRequestBodyLevel
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
	eEe.eClnt, err = elasticsearch.NewClient(eEe.clntOpts)
	return
}

// ExportEvent implements EventExporter
func (eEe *ElasticEE) ExportEvent(ctx *context.Context, ev, extraData any) (err error) {
	eEe.reqs.get()
	eEe.RLock()
	defer func() {
		eEe.RUnlock()
		eEe.reqs.done()
	}()
	if eEe.eClnt == nil {
		return utils.ErrDisconnected
	}
	key := extraData.(string)
	eReq := esapi.IndexRequest{
		Index:               eEe.opts.Index,
		DocumentID:          key,
		Body:                bytes.NewReader(ev.([]byte)),
		Refresh:             "true",
		IfPrimaryTerm:       eEe.opts.IfPrimaryTerm,
		IfSeqNo:             eEe.opts.IfSeqNo,
		OpType:              eEe.opts.OpType,
		Pipeline:            eEe.opts.Pipeline,
		Routing:             eEe.opts.Routing,
		Timeout:             eEe.opts.Timeout,
		Version:             eEe.opts.Version,
		VersionType:         eEe.opts.VersionType,
		WaitForActiveShards: eEe.opts.WaitForActiveShards,
	}

	var resp *esapi.Response
	if resp, err = eReq.Do(ctx, eEe.eClnt); err != nil {
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

func (eEE *ElasticEE) ExtraData(ev *utils.CGREvent) any {
	return utils.ConcatenatedKey(
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaOriginID), utils.GenUUID()),
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaRunID), utils.MetaDefault),
	)
}
