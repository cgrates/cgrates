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

func NewPosterJSONMapEE(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc utils.MapStorage) (pstrJSON *PosterJSONMapEE, err error) {
	pstrJSON = &PosterJSONMapEE{
		id:      cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg:  cgrCfg,
		cfgIdx:  cfgIdx,
		filterS: filterS,
		dc:      dc,
	}
	switch cgrCfg.EEsCfg().Exporters[cfgIdx].Type {
	case utils.MetaAMQPjsonMap:
		pstrJSON.poster = engine.NewAMQPPoster(cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
			cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts, cgrCfg.EEsCfg().Exporters[cfgIdx].Opts)
	case utils.MetaAMQPV1jsonMap:
		pstrJSON.poster = engine.NewAMQPv1Poster(cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
			cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts, cgrCfg.EEsCfg().Exporters[cfgIdx].Opts)
	case utils.MetaSQSjsonMap:
		pstrJSON.poster = engine.NewSQSPoster(cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
			cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts, cgrCfg.EEsCfg().Exporters[cfgIdx].Opts)
	case utils.MetaKafkajsonMap:
		pstrJSON.poster = engine.NewKafkaPoster(cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
			cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts, cgrCfg.EEsCfg().Exporters[cfgIdx].Opts)
	case utils.MetaS3jsonMap:
		pstrJSON.poster = engine.NewS3Poster(cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
			cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts, cgrCfg.EEsCfg().Exporters[cfgIdx].Opts)
	}
	return
}

// PosterJSONMapEE implements EventExporter interface for .csv files
type PosterJSONMapEE struct {
	id      string
	cgrCfg  *config.CGRConfig
	cfgIdx  int // index of config instance within ERsCfg.Readers
	filterS *engine.FilterS
	poster  engine.Poster
	dc      utils.MapStorage
	sync.RWMutex
}

// ID returns the identificator of this exporter
func (pstrEE *PosterJSONMapEE) ID() string {
	return pstrEE.id
}

// OnEvicted implements EventExporter, doing the cleanup before exit
func (pstrEE *PosterJSONMapEE) OnEvicted(string, interface{}) {
	pstrEE.poster.Close()
	return
}

// ExportEvent implements EventExporter
func (pstrEE *PosterJSONMapEE) ExportEvent(cgrEv *utils.CGREvent) (err error) {
	pstrEE.Lock()
	defer func() {
		if err != nil {
			pstrEE.dc[utils.NegativeExports].(utils.StringSet).Add(cgrEv.ID)
		} else {
			pstrEE.dc[utils.PositiveExports].(utils.StringSet).Add(cgrEv.ID)
		}
		pstrEE.Unlock()
	}()

	pstrEE.dc[utils.NumberOfEvents] = pstrEE.dc[utils.NumberOfEvents].(int64) + 1

	valMp := make(map[string]interface{})
	if len(pstrEE.cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].ContentFields()) == 0 {
		valMp = cgrEv.Event
	} else {
		eeReq := NewEventExporterRequest(utils.MapStorage(cgrEv.Event), pstrEE.dc, cgrEv.Opts,
			pstrEE.cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].Tenant,
			pstrEE.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(pstrEE.cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].Timezone,
				pstrEE.cgrCfg.GeneralCfg().DefaultTimezone), pstrEE.filterS)

		if err = eeReq.SetFields(pstrEE.cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].ContentFields()); err != nil {
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
	}
	updateEEMetrics(pstrEE.dc, cgrEv.Event, utils.FirstNonEmpty(pstrEE.cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].Timezone,
		pstrEE.cgrCfg.GeneralCfg().DefaultTimezone))
	cgrID := utils.FirstNonEmpty(engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.CGRID), utils.GenUUID())
	runID := utils.FirstNonEmpty(engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.RunID), utils.MetaDefault)
	var body []byte
	if body, err = json.Marshal(valMp); err != nil {
		return
	}
	if err = pstrEE.poster.Post(body, utils.ConcatenatedKey(cgrID, runID)); err != nil &&
		pstrEE.cgrCfg.GeneralCfg().FailedPostsDir != utils.MetaNone {
		engine.AddFailedPost(pstrEE.cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].ExportPath,
			pstrEE.cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].Type, utils.EventExporterS, body,
			pstrEE.cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].Opts)
	}
	return
}

func (pstrEE *PosterJSONMapEE) GetMetrics() utils.MapStorage {
	return pstrEE.dc.Clone()
}
