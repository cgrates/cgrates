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

func NewElasticEE(cfg *config.EventExporterCfg, dc *utils.SafeMapStorage) (*ElasticEE, error) {
	el := &ElasticEE{
		cfg:  cfg,
		dc:   dc,
		reqs: newConcReq(cfg.ConcurrentRequests),
	}
	if err := el.prepareOpts(); err != nil {
		return nil, err
	}
	return el, nil
}

// ElasticEE implements EventExporter interface for ElasticSearch export.
type ElasticEE struct {
	mu   sync.RWMutex
	cfg  *config.EventExporterCfg
	dc   *utils.SafeMapStorage
	reqs *concReq
	bytePreparing

	client    *elasticsearch.Client
	clientCfg elasticsearch.Config

	// indexReqOpts is used to store IndexRequest options for convenience
	// and does not represent the IndexRequest itself.
	indexReqOpts esapi.IndexRequest
}

// init will create all the necessary dependencies, including opening the file
func (e *ElasticEE) prepareOpts() error {
	opts := e.cfg.Opts

	// Parse index request options.
	e.indexReqOpts.Index = utils.CDRsTBL
	if opts.ElsIndex != nil {
		e.indexReqOpts.Index = *opts.ElsIndex
	}
	e.indexReqOpts.IfPrimaryTerm = opts.ElsIfPrimaryTerm
	e.indexReqOpts.IfSeqNo = opts.ElsIfSeqNo
	if opts.ElsOpType != nil {
		e.indexReqOpts.OpType = *opts.ElsOpType
	}
	if opts.ElsPipeline != nil {
		e.indexReqOpts.Pipeline = *opts.ElsPipeline
	}
	if opts.ElsRouting != nil {
		e.indexReqOpts.Routing = *opts.ElsRouting
	}
	if opts.ElsTimeout != nil {
		e.indexReqOpts.Timeout = *opts.ElsTimeout
	}
	e.indexReqOpts.Version = opts.ElsVersion
	if opts.ElsVersionType != nil {
		e.indexReqOpts.VersionType = *opts.ElsVersionType
	}
	if opts.ElsWaitForActiveShards != nil {
		e.indexReqOpts.WaitForActiveShards = *opts.ElsWaitForActiveShards
	}

	// Parse client config options.
	if opts.ElsCloud != nil && *opts.ElsCloud {
		e.clientCfg.CloudID = e.Cfg().ExportPath
	} else {
		e.clientCfg.Addresses = strings.Split(e.Cfg().ExportPath, utils.InfieldSep)
	}
	if opts.ElsUsername != nil {
		e.clientCfg.Username = *opts.ElsUsername
	}
	if opts.ElsPassword != nil {
		e.clientCfg.Password = *opts.ElsPassword
	}
	if opts.ElsAPIKey != nil {
		e.clientCfg.APIKey = *opts.ElsAPIKey
	}
	if opts.CAPath != nil {
		cacert, err := os.ReadFile(*opts.CAPath)
		if err != nil {
			return err
		}
		e.clientCfg.CACert = cacert
	}
	if opts.ElsCertificateFingerprint != nil {
		e.clientCfg.CertificateFingerprint = *opts.ElsCertificateFingerprint
	}
	if opts.ElsServiceToken != nil {
		e.clientCfg.ServiceToken = *opts.ElsServiceToken
	}
	if opts.ElsDiscoverNodesOnStart != nil {
		e.clientCfg.DiscoverNodesOnStart = *opts.ElsDiscoverNodesOnStart
	}
	if opts.ElsDiscoverNodeInterval != nil {
		e.clientCfg.DiscoverNodesInterval = *opts.ElsDiscoverNodeInterval
	}
	if opts.ElsEnableDebugLogger != nil {
		e.clientCfg.EnableDebugLogger = *opts.ElsEnableDebugLogger
	}
	if loggerType := opts.ElsLogger; loggerType != nil {
		var logger elastictransport.Logger
		switch *loggerType {
		case utils.ElsJson:
			logger = &elastictransport.JSONLogger{
				Output:             os.Stdout,
				EnableRequestBody:  true,
				EnableResponseBody: true,
			}
		case utils.ElsColor:
			logger = &elastictransport.ColorLogger{
				Output:             os.Stdout,
				EnableRequestBody:  true,
				EnableResponseBody: true,
			}
		case utils.ElsText:
			logger = &elastictransport.TextLogger{
				Output:             os.Stdout,
				EnableRequestBody:  true,
				EnableResponseBody: true,
			}
		default:
			return fmt.Errorf("invalid logger type: %q", *loggerType)
		}
		e.clientCfg.Logger = logger
	}
	if opts.ElsCompressRequestBody != nil {
		e.clientCfg.CompressRequestBody = *opts.ElsCompressRequestBody
	}
	if opts.ElsRetryOnStatus != nil {
		e.clientCfg.RetryOnStatus = *opts.ElsRetryOnStatus
	}
	if opts.ElsMaxRetries != nil {
		e.clientCfg.MaxRetries = *opts.ElsMaxRetries
	}
	if opts.ElsDisableRetry != nil {
		e.clientCfg.DisableRetry = *opts.ElsDisableRetry
	}
	if opts.ElsCompressRequestBodyLevel != nil {
		e.clientCfg.CompressRequestBodyLevel = *opts.ElsCompressRequestBodyLevel
	}
	return nil
}

func (e *ElasticEE) Cfg() *config.EventExporterCfg { return e.cfg }

func (e *ElasticEE) Connect() (err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.client != nil { // check if connection is cached
		return
	}
	e.client, err = elasticsearch.NewClient(e.clientCfg)
	return
}

// ExportEvent implements EventExporter
func (e *ElasticEE) ExportEvent(ctx *context.Context, ev, extraData any) error {
	e.reqs.get()
	e.mu.RLock()
	defer func() {
		e.mu.RUnlock()
		e.reqs.done()
	}()
	if e.client == nil {
		return utils.ErrDisconnected
	}
	key := extraData.(string)
	req := esapi.IndexRequest{
		DocumentID:          key,
		Body:                bytes.NewReader(ev.([]byte)),
		Refresh:             "true",
		Index:               e.indexReqOpts.Index,
		IfPrimaryTerm:       e.indexReqOpts.IfPrimaryTerm,
		IfSeqNo:             e.indexReqOpts.IfSeqNo,
		OpType:              e.indexReqOpts.OpType,
		Pipeline:            e.indexReqOpts.Pipeline,
		Routing:             e.indexReqOpts.Routing,
		Timeout:             e.indexReqOpts.Timeout,
		Version:             e.indexReqOpts.Version,
		VersionType:         e.indexReqOpts.VersionType,
		WaitForActiveShards: e.indexReqOpts.WaitForActiveShards,
	}

	resp, err := req.Do(ctx, e.client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.IsError() {
		var errResp map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return err
		}
		utils.Logger.Warning(fmt.Sprintf(
			"<%s> exporter %q: failed to index document: %+v",
			utils.EEs, e.Cfg().ID, errResp))
	}
	return nil
}

func (e *ElasticEE) Close() error {
	e.mu.Lock()
	e.client = nil
	e.mu.Unlock()
	return nil
}

func (e *ElasticEE) GetMetrics() *utils.SafeMapStorage { return e.dc }

func (eEE *ElasticEE) ExtraData(ev *utils.CGREvent) any {
	return utils.ConcatenatedKey(
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaOriginID), utils.GenUUID()),
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaRunID), utils.MetaDefault),
	)
}
