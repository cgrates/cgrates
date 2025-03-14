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
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/optype"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
)

// ElasticEE implements EventExporter interface for ElasticSearch export.
type ElasticEE struct {
	mu   sync.RWMutex
	cfg  *config.EventExporterCfg
	dc   *utils.ExporterMetrics
	reqs *concReq

	client    *elasticsearch.TypedClient
	clientCfg elasticsearch.Config
}

func NewElasticEE(cfg *config.EventExporterCfg, dc *utils.ExporterMetrics) (*ElasticEE, error) {
	el := &ElasticEE{
		cfg:  cfg,
		dc:   dc,
		reqs: newConcReq(cfg.ConcurrentRequests),
	}
	if err := el.parseClientOpts(); err != nil {
		return nil, err
	}
	return el, nil
}

// init will create all the necessary dependencies, including opening the file
func (e *ElasticEE) parseClientOpts() error {
	opts := e.cfg.Opts.Els
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
	if opts.CAPath != nil {
		cacert, err := os.ReadFile(*opts.CAPath)
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
	e.client, err = elasticsearch.NewTypedClient(e.clientCfg)
	return
}

// ExportEvent implements EventExporter
func (e *ElasticEE) ExportEvent(event any, key string) error {
	e.reqs.get()
	e.mu.RLock()
	defer func() {
		e.mu.RUnlock()
		e.reqs.done()
	}()
	if e.client == nil {
		return utils.ErrDisconnected
	}

	// Build and send index request.
	opts := e.cfg.Opts.Els
	indexName := utils.CDRsTBL
	if opts.Index != nil {
		indexName = *opts.Index
	}
	req := e.client.Index(indexName).
		Id(key).
		Request(event)

	if opts.Refresh != nil {
		req.Refresh(refresh.Refresh{Name: *opts.Refresh})
	}
	if opts.OpType != nil {
		req.OpType(optype.OpType{Name: *opts.OpType})
	}
	if opts.Pipeline != nil {
		req.Pipeline(*opts.Pipeline)
	}
	if opts.Routing != nil {
		req.Routing(*opts.Routing)
	}
	if opts.Timeout != nil {
		req.Timeout((*opts.Timeout).String())
	}
	if opts.WaitForActiveShards != nil {
		req.WaitForActiveShards(*opts.WaitForActiveShards)
	}
	_, err := req.Do(context.TODO())
	return err
}

func (e *ElasticEE) PrepareMap(cgrEv *utils.CGREvent) (any, error) {
	return cgrEv.Event, nil
}

func (e *ElasticEE) PrepareOrderMap(onm *utils.OrderedNavigableMap) (any, error) {
	preparedMap := make(map[string]any)
	for el := onm.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		item, err := onm.Field(path)
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> exporter %q: failed to retrieve field at path %q",
				utils.EEs, e.cfg.ID, path))
			continue
		}
		path = path[:len(path)-1] // remove the last index
		preparedMap[strings.Join(path, utils.NestingSep)] = item.String()
	}
	return preparedMap, nil
}

func (e *ElasticEE) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.client = nil
	return nil
}

func (e *ElasticEE) GetMetrics() *utils.ExporterMetrics { return e.dc }
