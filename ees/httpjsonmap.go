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
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewHTTPJsonMapEe(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc utils.MapStorage) (httpJSON *HTTPJsonMapEe, err error) {
	dc[utils.ExporterID] = cgrCfg.EEsCfg().Exporters[cfgIdx].ID
	httpJSON = &HTTPJsonMapEe{
		id:      cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg:  cgrCfg,
		cfgIdx:  cfgIdx,
		filterS: filterS,
		dc:      dc,
	}
	if cgrCfg.EEsCfg().Exporters[cfgIdx].Type == utils.MetaHTTPjsonMap {
		httpJSON.httpPoster, err = engine.NewHTTPPoster(cgrCfg.GeneralCfg().HttpSkipTlsVerify,
			cgrCfg.GeneralCfg().ReplyTimeout, cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
			utils.PosterTransportContentTypes[cgrCfg.EEsCfg().Exporters[cfgIdx].Type], cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts)
	}
	return
}

// HTTPJsonMapEe implements EventExporter interface for .csv files
type HTTPJsonMapEe struct {
	id         string
	cgrCfg     *config.CGRConfig
	cfgIdx     int // index of config instance within ERsCfg.Readers
	filterS    *engine.FilterS
	httpPoster *engine.HTTPPoster
	dc         utils.MapStorage
	sync.RWMutex
}

// ID returns the identificator of this exporter
func (httpJson *HTTPJsonMapEe) ID() string {
	return httpJson.id
}

// OnEvicted implements EventExporter, doing the cleanup before exit
func (httpJson *HTTPJsonMapEe) OnEvicted(string, interface{}) {
	return
}

// ExportEvent implements EventExporter
func (httpJson *HTTPJsonMapEe) ExportEvent(cgrEv *utils.CGREvent) (err error) {
	httpJson.Lock()
	defer func() {
		if err != nil {
			httpJson.dc[utils.NegativeExports].(utils.StringSet).Add(cgrEv.ID)
		} else {
			httpJson.dc[utils.PositiveExports].(utils.StringSet).Add(cgrEv.ID)
		}
		httpJson.Unlock()
	}()

	httpJson.dc[utils.NumberOfEvents] = httpJson.dc[utils.NumberOfEvents].(int) + 1

	valMp := make(map[string]string)
	eeReq := NewEventExporterRequest(utils.MapStorage(cgrEv.Event), httpJson.dc,
		cgrEv.Tenant, httpJson.cgrCfg.GeneralCfg().DefaultTimezone, httpJson.filterS)

	if err = eeReq.SetFields(httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].ContentFields()); err != nil {
		return
	}
	for el := eeReq.cnt.GetFirstElement(); el != nil; el = el.Next() {
		var nmIt utils.NMInterface
		if nmIt, err = eeReq.cnt.Field(el.Value); err != nil {
			return
		}
		itm, isNMItem := nmIt.(*config.NMItem)
		if !isNMItem {
			err = fmt.Errorf("cannot encode reply value: %s, err: not NMItems", utils.ToJSON(el.Value))
			return
		}
		if itm == nil {
			continue // all attributes, not writable to diameter packet
		}
		valMp[strings.Join(itm.Path, utils.NestingSep)] = utils.IfaceAsString(itm.Data)
	}
	updateEEMetrics(httpJson.dc, cgrEv.Event, httpJson.cgrCfg.GeneralCfg().DefaultTimezone)
	cgrID := utils.FirstNonEmpty(engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.CGRID), utils.GenUUID())
	runID := utils.FirstNonEmpty(engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.RunID), utils.MetaDefault)
	var body []byte
	if body, err = json.Marshal(valMp); err != nil {
		return
	}
	err = httpJson.post(body, utils.ConcatenatedKey(cgrID, runID))
	return
}

func (httpJson *HTTPJsonMapEe) post(body []byte, key string) (err error) {
	switch httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].Type {
	case utils.MetaHTTPjsonMap:
		err = httpJson.httpPoster.Post(body, utils.EmptyString)
	case utils.MetaAMQPjsonMap:
		err = engine.PostersCache.PostAMQP(httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].ExportPath,
			httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].Attempts, body)
	case utils.MetaAMQPV1jsonMap:
		err = engine.PostersCache.PostAMQPv1(httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].ExportPath,
			httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].Attempts, body)
	case utils.MetaSQSjsonMap:
		err = engine.PostersCache.PostSQS(httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].ExportPath,
			httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].Attempts, body)
	case utils.MetaKafkajsonMap:
		err = engine.PostersCache.PostKafka(httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].ExportPath,
			httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].Attempts, body, key)
	case utils.MetaS3jsonMap:
		err = engine.PostersCache.PostS3(httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].ExportPath,
			httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].Attempts, body, key)
	}
	if err != nil && httpJson.cgrCfg.GeneralCfg().FailedPostsDir != utils.META_NONE {
		engine.AddFailedPost(httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].ExportPath,
			httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].Type, utils.EventExporterS, body)
	}
	return
}

func (httpJson *HTTPJsonMapEe) GetMetrics() utils.MapStorage {
	return httpJson.dc.Clone()
}
