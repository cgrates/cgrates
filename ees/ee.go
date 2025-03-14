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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type EventExporter interface {
	Cfg() *config.EventExporterCfg      // return the config
	Connect() error                     // called before exporting an event to make sure it is connected
	ExportEvent(any, string) error      // called on each event to be exported
	Close() error                       // called when the exporter needs to terminate
	GetMetrics() *utils.ExporterMetrics // called to get metrics
	PrepareMap(*utils.CGREvent) (any, error)
	PrepareOrderMap(*utils.OrderedNavigableMap) (any, error)
}

// NewEventExporter produces exporters
func NewEventExporter(cfg *config.EventExporterCfg, cgrCfg *config.CGRConfig, filterS *engine.FilterS,
	connMngr *engine.ConnManager) (ee EventExporter, err error) {
	timezone := utils.FirstNonEmpty(cfg.Timezone, cgrCfg.GeneralCfg().DefaultTimezone)
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}
	em := utils.NewExporterMetrics(cfg.MetricsResetSchedule, loc)

	switch cfg.Type {
	case utils.MetaFileCSV:
		return NewFileCSVee(cfg, cgrCfg, filterS, em)
	case utils.MetaFileFWV:
		return NewFileFWVee(cfg, cgrCfg, filterS, em)
	case utils.MetaHTTPPost:
		return NewHTTPPostEE(cfg, cgrCfg, filterS, em)
	case utils.MetaHTTPjsonMap:
		return NewHTTPjsonMapEE(cfg, cgrCfg, filterS, em)
	case utils.MetaNatsjsonMap:
		return NewNatsEE(cfg, cgrCfg.GeneralCfg().NodeID,
			cgrCfg.GeneralCfg().ConnectTimeout, em)
	case utils.MetaAMQPjsonMap:
		return NewAMQPee(cfg, em), nil
	case utils.MetaAMQPV1jsonMap:
		return NewAMQPv1EE(cfg, em), nil
	case utils.MetaS3jsonMap:
		return NewS3EE(cfg, em), nil
	case utils.MetaSQSjsonMap:
		return NewSQSee(cfg, em), nil
	case utils.MetaKafkajsonMap:
		return NewKafkaEE(cfg, em)
	case utils.MetaVirt:
		return NewVirtualEE(cfg, em), nil
	case utils.MetaElastic:
		return NewElasticEE(cfg, em)
	case utils.MetaSQL:
		return NewSQLEe(cfg, em)
	case utils.MetaLog:
		return NewLogEE(cfg, em), nil
	case utils.MetaRPC:
		return NewRpcEE(cfg, em, connMngr)
	default:
		return nil, fmt.Errorf("unsupported exporter type: <%s>", cfg.Type)
	}
}

func newConcReq(limit int) (c *concReq) {
	c = &concReq{limit: limit}
	if limit > 0 {
		c.reqs = make(chan struct{}, limit)
		for i := 0; i < limit; i++ {
			c.reqs <- struct{}{}
		}
	}
	return
}

type concReq struct {
	reqs  chan struct{}
	limit int
}

func (c *concReq) get() {
	if c.limit > 0 {
		<-c.reqs
	}
}
func (c *concReq) done() {
	if c.limit > 0 {
		c.reqs <- struct{}{}
	}
}

// composeHeaderTrailer will return the orderNM for *hdr or *trl
func composeHeaderTrailer(prfx string, fields []*config.FCTemplate, em utils.DataStorage, cfg *config.CGRConfig, fltS *engine.FilterS) (r *utils.OrderedNavigableMap, err error) {
	r = utils.NewOrderedNavigableMap()
	err = engine.NewExportRequest(map[string]utils.DataStorage{
		utils.MetaEM:  em,
		utils.MetaCfg: cfg.GetDataProvider(),
	}, cfg.GeneralCfg().DefaultTenant, fltS,
		map[string]*utils.OrderedNavigableMap{prfx: r}).SetFields(fields)
	return
}

func updateEEMetrics(em *utils.ExporterMetrics, cgrID string, ev engine.MapEvent, hasError bool, timezone string) {
	em.Lock()
	defer em.Unlock()
	if hasError {
		em.MapStorage[utils.NegativeExports].(utils.StringSet).Add(cgrID)
	} else {
		em.MapStorage[utils.PositiveExports].(utils.StringSet).Add(cgrID)
	}
	if aTime, err := ev.GetTime(utils.AnswerTime, timezone); err == nil {
		if _, has := em.MapStorage[utils.FirstEventATime]; !has {
			em.MapStorage[utils.FirstEventATime] = time.Time{}
		}
		if _, has := em.MapStorage[utils.LastEventATime]; !has {
			em.MapStorage[utils.LastEventATime] = time.Time{}
		}
		if em.MapStorage[utils.FirstEventATime].(time.Time).IsZero() ||
			aTime.Before(em.MapStorage[utils.FirstEventATime].(time.Time)) {
			em.MapStorage[utils.FirstEventATime] = aTime
		}
		if aTime.After(em.MapStorage[utils.LastEventATime].(time.Time)) {
			em.MapStorage[utils.LastEventATime] = aTime
		}
	}
	if oID, err := ev.GetTInt64(utils.OrderID); err == nil {
		if _, has := em.MapStorage[utils.FirstExpOrderID]; !has {
			em.MapStorage[utils.FirstExpOrderID] = int64(0)
		}
		if _, has := em.MapStorage[utils.LastExpOrderID]; !has {
			em.MapStorage[utils.LastExpOrderID] = int64(0)
		}
		if em.MapStorage[utils.FirstExpOrderID].(int64) == 0 ||
			em.MapStorage[utils.FirstExpOrderID].(int64) > oID {
			em.MapStorage[utils.FirstExpOrderID] = oID
		}
		if em.MapStorage[utils.LastExpOrderID].(int64) < oID {
			em.MapStorage[utils.LastExpOrderID] = oID
		}
	}
	if cost, err := ev.GetFloat64(utils.Cost); err == nil {
		if _, has := em.MapStorage[utils.TotalCost]; !has {
			em.MapStorage[utils.TotalCost] = float64(0.0)
		}
		em.MapStorage[utils.TotalCost] = em.MapStorage[utils.TotalCost].(float64) + cost
	}
	if tor, err := ev.GetString(utils.ToR); err == nil {
		if usage, err := ev.GetDuration(utils.Usage); err == nil {
			switch tor {
			case utils.MetaVoice:
				if _, has := em.MapStorage[utils.TotalDuration]; !has {
					em.MapStorage[utils.TotalDuration] = time.Duration(0)
				}
				em.MapStorage[utils.TotalDuration] = em.MapStorage[utils.TotalDuration].(time.Duration) + usage
			case utils.MetaSMS:
				if _, has := em.MapStorage[utils.TotalSMSUsage]; !has {
					em.MapStorage[utils.TotalSMSUsage] = time.Duration(0)
				}
				em.MapStorage[utils.TotalSMSUsage] = em.MapStorage[utils.TotalSMSUsage].(time.Duration) + usage
			case utils.MetaMMS:
				if _, has := em.MapStorage[utils.TotalMMSUsage]; !has {
					em.MapStorage[utils.TotalMMSUsage] = time.Duration(0)
				}
				em.MapStorage[utils.TotalMMSUsage] = em.MapStorage[utils.TotalMMSUsage].(time.Duration) + usage
			case utils.MetaGeneric:
				if _, has := em.MapStorage[utils.TotalGenericUsage]; !has {
					em.MapStorage[utils.TotalGenericUsage] = time.Duration(0)
				}
				em.MapStorage[utils.TotalGenericUsage] = em.MapStorage[utils.TotalGenericUsage].(time.Duration) + usage
			case utils.MetaData:
				if _, has := em.MapStorage[utils.TotalDataUsage]; !has {
					em.MapStorage[utils.TotalDataUsage] = time.Duration(0)
				}
				em.MapStorage[utils.TotalDataUsage] = em.MapStorage[utils.TotalDataUsage].(time.Duration) + usage
			}
		}
	}
}

type bytePreparing struct{}

func (bytePreparing) PrepareMap(mp *utils.CGREvent) (any, error) {
	return json.Marshal(mp.Event)
}
func (bytePreparing) PrepareOrderMap(mp *utils.OrderedNavigableMap) (any, error) {
	valMp := make(map[string]any)
	for el := mp.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		nmIt, _ := mp.Field(path)
		path = path[:len(path)-1] // remove the last index
		valMp[strings.Join(path, utils.NestingSep)] = nmIt.String()
	}
	return json.Marshal(valMp)
}

type slicePreparing struct{}

func (slicePreparing) PrepareMap(mp *utils.CGREvent) (any, error) {
	csvRecord := make([]string, 0, len(mp.Event))
	for _, val := range mp.Event {
		csvRecord = append(csvRecord, utils.IfaceAsString(val))
	}
	return csvRecord, nil
}
func (slicePreparing) PrepareOrderMap(mp *utils.OrderedNavigableMap) (any, error) {
	return mp.OrderedFieldsAsStrings(), nil
}
