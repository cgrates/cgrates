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

package efs

import (
	"bytes"
	"encoding/gob"
	"os"
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/segmentio/kafka-go"
)

// FailedExportersLogg is a failover poster for kafka logger type
type FailedExportersLogg struct {
	lk             sync.RWMutex
	Path           string
	Opts           map[string]interface{} // this is meta
	Format         string
	Events         []interface{}
	FailedPostsDir string
	Module         string

	connMngr *engine.ConnManager
	cfg      *config.CGRConfig
}

// AddEvent adds one event
func (expEv *FailedExportersLogg) AddEvent(ev interface{}) {
	expEv.lk.Lock()
	expEv.Events = append(expEv.Events, ev)
	expEv.lk.Unlock()
}

// NewExportEventsFromFile returns ExportEvents from the file
// used only on replay failed post
func NewExportEventsFromFile(filePath string) (expEv *FailedExportersLogg, err error) {
	var fileContent []byte
	if fileContent, err = os.ReadFile(filePath); err != nil {
		return nil, err
	}
	if err = os.Remove(filePath); err != nil {
		return nil, err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(fileContent))
	// unmarshall it
	expEv = new(FailedExportersLogg)
	err = dec.Decode(&expEv)
	return
}

// ReplayFailedPosts tryies to post cdrs again
func (expEv *FailedExportersLogg) ReplayFailedPosts(ctx *context.Context, attempts int, tnt string) (err error) {
	nodeID := utils.IfaceAsString(expEv.Opts[utils.NodeID])
	logLvl, err := utils.IfaceAsInt(expEv.Opts[utils.Level])
	if err != nil {
		return
	}
	expLogger := engine.NewExportLogger(nodeID, tnt, logLvl,
		expEv.connMngr, expEv.cfg)
	for _, event := range expEv.Events {
		var content []byte
		if content, err = utils.ToUnescapedJSON(event); err != nil {
			return
		}
		if err = expLogger.Writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(utils.GenUUID()),
			Value: content,
		}); err != nil {
			var reply string
			// if there are any errors in kafka, we will post in FailedPostDirectory
			if err = expEv.connMngr.Call(ctx, expEv.cfg.LoggerCfg().EFsConns, utils.EfSv1ProcessEvent,
				&utils.ArgsFailedPosts{
					Tenant:    tnt,
					Path:      expLogger.Writer.Addr.String(),
					Event:     event,
					FailedDir: expLogger.FldPostDir,
					Module:    utils.Kafka,
					APIOpts:   expLogger.GetMeta(),
				}, &reply); err != nil {
				return err
			}
			return nil
		}
	}
	return err
}
