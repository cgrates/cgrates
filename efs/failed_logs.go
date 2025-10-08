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

// FailedExportersLog is a failover poster for kafka logger type
type FailedExportersLog struct {
	lk             sync.RWMutex
	Path           string
	Opts           map[string]any // this is meta
	Format         string
	Events         []any
	FailedPostsDir string
	Module         string

	connMngr *engine.ConnManager
	cfg      *config.CGRConfig
}

// AddEvent adds one event
func (expEv *FailedExportersLog) AddEvent(ev any) {
	expEv.lk.Lock()
	defer expEv.lk.Unlock()
	expEv.Events = append(expEv.Events, ev)
}

// NewExportEventsFromFile returns ExportEvents from the file
// used only on replay failed post
func NewExportEventsFromFile(filePath string) (*FailedExportersLog, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if err := os.Remove(filePath); err != nil {
		return nil, err
	}
	var expEv FailedExportersLog
	dec := gob.NewDecoder(bytes.NewBuffer(content))
	if err := dec.Decode(&expEv); err != nil {
		return nil, err
	}
	return &expEv, nil
}

// ReplayFailedPosts tries to repost failed cdrs.
func (expEv *FailedExportersLog) ReplayFailedPosts(ctx *context.Context, attempts int, tnt string) error {
	expLogger := engine.NewExportLogger(ctx, tnt,
		expEv.connMngr, expEv.cfg)

	// Fall back to config values even if LogLevel and NodeID are always passed to
	// the opts (through the GetMeta method on the ExportLogger), just to be safe.
	if v, has := expEv.Opts[utils.NodeID]; has {
		expLogger.NodeID = utils.IfaceAsString(v)
	}
	if v, has := expEv.Opts[utils.Level]; has {
		lvl, err := utils.IfaceAsInt(v)
		if err != nil {
			return err
		}
		expLogger.LogLevel = lvl
	}

	for _, event := range expEv.Events {
		content, err := utils.ToUnescapedJSON(event)
		if err != nil {
			return err
		}
		if err := expLogger.Writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(utils.GenUUID()),
			Value: content,
		}); err != nil {
			var reply string
			// if there are any errors in kafka, we will post in FailedPostDirectory
			return expEv.connMngr.Call(ctx, expEv.cfg.LoggerCfg().EFsConns, utils.EfSv1ProcessEvent,
				&utils.ArgsFailedPosts{
					Tenant:    tnt,
					Path:      expLogger.Writer.Addr.String(),
					Event:     event,
					FailedDir: expLogger.FldPostDir,
					Module:    utils.Kafka,
					APIOpts:   expLogger.GetMeta(),
				}, &reply)
		}
	}
	return nil
}
