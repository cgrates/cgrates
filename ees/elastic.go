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
	opts := e.cfg.Opts.Els

	// Parse index request options.
	e.indexReqOpts.Index = utils.CDRsTBL
	if opts.Index != nil {
		e.indexReqOpts.Index = *opts.Index
	}
	e.indexReqOpts.IfPrimaryTerm = opts.IfPrimaryTerm
	e.indexReqOpts.IfSeqNo = opts.IfSeqNo
	if opts.OpType != nil {
		e.indexReqOpts.OpType = *opts.OpType
	}
	if opts.Pipeline != nil {
		e.indexReqOpts.Pipeline = *opts.Pipeline
	}
	if opts.Routing != nil {
		e.indexReqOpts.Routing = *opts.Routing
	}
	if opts.Timeout != nil {
		e.indexReqOpts.Timeout = *opts.Timeout
	}
	e.indexReqOpts.Version = opts.Version
	if opts.VersionType != nil {
		e.indexReqOpts.VersionType = *opts.VersionType
	}
	if opts.WaitForActiveShards != nil {
		e.indexReqOpts.WaitForActiveShards = *opts.WaitForActiveShards
	}

	// Parse client config options.
	if opts.Cloud != nil && *opts.Cloud {
		e.clientCfg.CloudID = e.Cfg().ExportPath
	} else {
		e.clientCfg.Addresses = strings.Split(e.Cfg().ExportPath, utils.InfieldSep)
	}
	if opts.Username != nil {
		e.clientCfg.Username = *opts.Username
	}
	if opts.Password != nil {
		e.clientCfg.Password = *opts.Password
	}
	if opts.APIKey != nil {
		e.clientCfg.APIKey = *opts.APIKey
	}
	if e.Cfg().Opts.RPC.CAPath != nil {
		cacert, err := os.ReadFile(*e.Cfg().Opts.RPC.CAPath)
		if err != nil {
			return err
		}
		e.clientCfg.CACert = cacert
	}
	if opts.CertificateFingerprint != nil {
		e.clientCfg.CertificateFingerprint = *opts.CertificateFingerprint
	}
	if opts.ServiceToken != nil {
		e.clientCfg.ServiceToken = *opts.ServiceToken
	}
	if opts.DiscoverNodesOnStart != nil {
		e.clientCfg.DiscoverNodesOnStart = *opts.DiscoverNodesOnStart
	}
	if opts.DiscoverNodeInterval != nil {
		e.clientCfg.DiscoverNodesInterval = *opts.DiscoverNodeInterval
	}
	if opts.EnableDebugLogger != nil {
		e.clientCfg.EnableDebugLogger = *opts.EnableDebugLogger
	}
	if loggerType := opts.Logger; loggerType != nil {
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
	if opts.CompressRequestBody != nil {
		e.clientCfg.CompressRequestBody = *opts.CompressRequestBody
	}
	if opts.RetryOnStatus != nil {
		e.clientCfg.RetryOnStatus = *opts.RetryOnStatus
	}
	if opts.MaxRetries != nil {
		e.clientCfg.MaxRetries = *opts.MaxRetries
	}
	if opts.DisableRetry != nil {
		e.clientCfg.DisableRetry = *opts.DisableRetry
	}
	if opts.CompressRequestBodyLevel != nil {
		e.clientCfg.CompressRequestBodyLevel = *opts.CompressRequestBodyLevel
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
func (e *ElasticEE) ExportEvent(ev any, key string) error {
	e.reqs.get()
	e.mu.RLock()
	defer func() {
		e.mu.RUnlock()
		e.reqs.done()
	}()
	if e.client == nil {
		return utils.ErrDisconnected
	}
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

	resp, err := req.Do(context.Background(), e.client)
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
