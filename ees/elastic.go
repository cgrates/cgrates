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
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/optype"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/versiontype"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
)

// ElasticEE implements EventExporter interface for ElasticSearch export.
type ElasticEE struct {
	mu   sync.RWMutex
	cfg  *config.EventExporterCfg
	dc   *utils.SafeMapStorage
	reqs *concReq

	client    *elasticsearch.TypedClient
	clientCfg elasticsearch.Config
}

func NewElasticEE(cfg *config.EventExporterCfg, dc *utils.SafeMapStorage) (*ElasticEE, error) {
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

func (e *ElasticEE) parseClientOpts() error {
	opts := e.cfg.Opts
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
	e.client, err = elasticsearch.NewTypedClient(e.clientCfg)
	return
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

	// Build and send index request.
	key := extraData.(string)
	opts := e.cfg.Opts
	indexName := utils.CDRsTBL
	if opts.ElsIndex != nil {
		indexName = *opts.ElsIndex
	}
	req := e.client.Index(indexName).
		Id(key).
		Request(event).
		Refresh(refresh.True)

	if opts.ElsIfPrimaryTerm != nil {
		req.IfPrimaryTerm(strconv.Itoa(*opts.ElsIfPrimaryTerm))
	}
	if opts.ElsIfSeqNo != nil {
		req.IfSeqNo(strconv.Itoa(*opts.ElsIfSeqNo))
	}
	if opts.ElsOpType != nil {
		req.OpType(optype.OpType{Name: *opts.ElsOpType})
	}
	if opts.ElsPipeline != nil {
		req.Pipeline(*opts.ElsPipeline)
	}
	if opts.ElsRouting != nil {
		req.Routing(*opts.ElsRouting)
	}
	if opts.ElsTimeout != nil {
		req.Timeout((*opts.ElsTimeout).String())
	}
	if opts.ElsVersion != nil {
		req.Version(strconv.Itoa(*opts.ElsVersion))
	}
	if opts.ElsVersionType != nil {
		req.VersionType(versiontype.VersionType{Name: *opts.ElsVersionType})
	}
	if opts.ElsWaitForActiveShards != nil {
		req.WaitForActiveShards(*opts.ElsWaitForActiveShards)
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

func (e *ElasticEE) GetMetrics() *utils.SafeMapStorage { return e.dc }

func (eEE *ElasticEE) ExtraData(ev *utils.CGREvent) any {
	return utils.ConcatenatedKey(
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaOriginID), utils.GenUUID()),
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaRunID), utils.MetaDefault),
	)
}
