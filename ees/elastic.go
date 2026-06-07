/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package ees

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// ElasticEE implements EventExporter interface for ElasticSearch export.
type ElasticEE struct {
	mu   sync.RWMutex
	cfg  *config.EventExporterCfg
	em   *utils.ExporterMetrics
	reqs *concReq

	client    *elastictransport.Client
	clientCfg elastictransport.Config

	docURL   string // host left empty, the transport fills it in
	rawQuery string
}

func NewElasticEE(cfg *config.EventExporterCfg, em *utils.ExporterMetrics) (*ElasticEE, error) {
	el := &ElasticEE{
		cfg:  cfg,
		em:   em,
		reqs: newConcReq(cfg.ConcurrentRequests),
	}
	if err := el.parseClientOpts(); err != nil {
		return nil, err
	}
	el.parseRequestOpts()
	return el, nil
}

func (e *ElasticEE) parseClientOpts() error {
	opts := e.cfg.Opts
	if opts.ElsUsername != nil {
		e.clientCfg.Username = *opts.ElsUsername
	}
	if opts.ElsPassword != nil {
		e.clientCfg.Password = *opts.ElsPassword
	}
	if opts.ElsAPIKey != nil {
		e.clientCfg.APIKey = *opts.ElsAPIKey
	}
	if opts.ElsCAPath != nil {
		cacert, err := os.ReadFile(*opts.ElsCAPath)
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
	if opts.ElsDiscoverNodeInterval != nil {
		e.clientCfg.DiscoverNodesInterval = *opts.ElsDiscoverNodeInterval
	}
	if opts.ElsEnableDebugLogger != nil {
		e.clientCfg.EnableDebugLogger = *opts.ElsEnableDebugLogger
	}
	e.clientCfg.Logger = elasticLogger(opts.ElsLogger)
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

func (e *ElasticEE) parseRequestOpts() {
	opts := e.cfg.Opts
	indexName := utils.CDRsTBL
	if opts.ElsIndex != nil {
		indexName = *opts.ElsIndex
	}
	e.docURL = "http:///" + indexName + "/_doc/"

	q := make(url.Values)
	if opts.ElsRefresh != nil {
		q.Set("refresh", *opts.ElsRefresh)
	}
	if opts.ElsOpType != nil {
		q.Set("op_type", *opts.ElsOpType)
	}
	if opts.ElsPipeline != nil {
		q.Set("pipeline", *opts.ElsPipeline)
	}
	if opts.ElsRouting != nil {
		q.Set("routing", *opts.ElsRouting)
	}
	if opts.ElsTimeout != nil {
		q.Set("timeout", (*opts.ElsTimeout).String())
	}
	if opts.ElsWaitForActiveShards != nil {
		q.Set("wait_for_active_shards", *opts.ElsWaitForActiveShards)
	}
	e.rawQuery = q.Encode()
}

func elasticLogger(loggerType *string) elastictransport.Logger {
	if loggerType == nil {
		return nil
	}
	switch *loggerType {
	case utils.ElsJson:
		return &elastictransport.JSONLogger{Output: os.Stdout, EnableRequestBody: true, EnableResponseBody: true}
	case utils.ElsColor:
		return &elastictransport.ColorLogger{Output: os.Stdout, EnableRequestBody: true, EnableResponseBody: true}
	case utils.ElsText:
		return &elastictransport.TextLogger{Output: os.Stdout, EnableRequestBody: true, EnableResponseBody: true}
	}
	return nil
}

func (e *ElasticEE) Cfg() *config.EventExporterCfg { return e.cfg }

func (e *ElasticEE) Connect() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.client != nil { // check if connection is cached
		return nil
	}
	addrs := strings.Split(e.Cfg().ExportPath, utils.InfieldSep)
	urls := make([]*url.URL, len(addrs))
	for i, addr := range addrs {
		u, err := url.Parse(addr)
		if err != nil {
			return err
		}
		urls[i] = u
	}
	e.clientCfg.URLs = urls
	client, err := elastictransport.New(e.clientCfg)
	if err != nil {
		return err
	}
	if opts := e.cfg.Opts; opts.ElsDiscoverNodesOnStart != nil && *opts.ElsDiscoverNodesOnStart {
		if err = client.DiscoverNodes(); err != nil {
			return err
		}
	}
	e.client = client
	return nil
}

// ExportEvent implements EventExporter
func (e *ElasticEE) ExportEvent(ctx *context.Context, event, extraData any) error {
	e.reqs.get()
	e.mu.RLock()
	defer func() {
		e.mu.RUnlock()
		e.reqs.done()
	}()
	if e.client == nil {
		return utils.ErrDisconnected
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	key := extraData.(string)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		e.docURL+url.PathEscape(key), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.URL.RawQuery = e.rawQuery

	resp, err := e.client.Perform(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusMultipleChoices {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("elasticsearch index failed: %s: %s", resp.Status, respBody)
	}
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

func (e *ElasticEE) PrepareMap(cgrEv *utils.CGREvent) (any, error) {
	return cgrEv.Event, nil
}

func (e *ElasticEE) PrepareOrderMap(onm *utils.OrderedNavigableMap) (any, error) {
	return onm.AsMap(), nil
}

func (e *ElasticEE) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.client = nil
	return nil
}

func (e *ElasticEE) GetMetrics() *utils.ExporterMetrics { return e.em }

func (e *ElasticEE) ExtraData(ev *utils.CGREvent) any {
	return utils.ConcatenatedKey(
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaOriginID), utils.GenUUID()),
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaRunID), utils.MetaDefault),
	)
}
