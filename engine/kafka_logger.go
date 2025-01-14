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
along with this program.  If not, see <http://.gnu.org/licenses/>
*/

package engine

import (
	"fmt"
	"log/syslog"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/segmentio/kafka-go"
)

// Logs to kafka
type ExportLogger struct {
	sync.Mutex
	efsConns []string
	connMgr  *ConnManager
	ctx      *context.Context

	LogLevel   int
	FldPostDir string
	Writer     *kafka.Writer
	NodeID     string
	Tenant     string
}

// NewExportLogger will export loggers to kafka
func NewExportLogger(ctx *context.Context, tenant string, connMgr *ConnManager,
	cfg *config.CGRConfig) *ExportLogger {
	return &ExportLogger{
		ctx:        ctx,
		efsConns:   cfg.LoggerCfg().EFsConns,
		connMgr:    connMgr,
		LogLevel:   cfg.LoggerCfg().Level,
		FldPostDir: cfg.LoggerCfg().Opts.FailedPostsDir,
		NodeID:     cfg.GeneralCfg().NodeID,
		Tenant:     tenant,
		Writer: &kafka.Writer{
			Addr:        kafka.TCP(cfg.LoggerCfg().Opts.KafkaConn),
			Topic:       cfg.LoggerCfg().Opts.KafkaTopic,
			MaxAttempts: cfg.LoggerCfg().Opts.KafkaAttempts,
		},
	}
}

func (el *ExportLogger) Close() (err error) {
	if el.Writer != nil {
		err = el.Writer.Close()
		el.Writer = nil
	}
	return
}

func (el *ExportLogger) call(m string, level int) (err error) {
	eventExport := &utils.CGREvent{
		Tenant: el.Tenant,
		Event: map[string]any{
			utils.NodeID:    el.NodeID,
			utils.Message:   m,
			utils.Severity:  level,
			utils.Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	// event will be exported through kafka as json format
	var content []byte
	if content, err = utils.ToUnescapedJSON(eventExport); err != nil {
		return
	}
	if err = el.Writer.WriteMessages(el.ctx, kafka.Message{
		Key:   []byte(utils.GenUUID()),
		Value: content,
	}); err != nil {
		// if there are any errors in kafka, we will post in FailedPostDirectory
		go func() {
			args := &utils.ArgsFailedPosts{
				Tenant:    el.Tenant,
				Path:      el.Writer.Addr.String(),
				Event:     eventExport,
				FailedDir: el.FldPostDir,
				Module:    utils.Kafka,
				APIOpts:   el.GetMeta(),
			}
			var reply string
			if err = el.connMgr.Call(el.ctx, el.efsConns, utils.EfSv1ProcessEvent,
				args, &reply); err != nil {
				utils.Logger.Warning(fmt.Sprintf(
					"<%s> failed to export log event: %v", utils.EFs, err))
			}
		}()
		// also the content should be printed as a stdout logger type
		return utils.ErrLoggerChanged
	}
	return
}

func (el *ExportLogger) Write(p []byte) (n int, err error) {
	n = len(p)
	err = el.call(string(p), 8)
	return
}

func (sl *ExportLogger) GetSyslog() *syslog.Writer {
	return nil
}

// GetLogLevel() returns the level logger number for the server
func (el *ExportLogger) GetLogLevel() int {
	return el.LogLevel
}

// SetLogLevel changes the log level
func (el *ExportLogger) SetLogLevel(level int) {
	el.LogLevel = level
}

// Alert logs to EEs with alert level
func (el *ExportLogger) Alert(m string) (err error) {
	if el.LogLevel < utils.LOGLEVEL_ALERT {
		return nil
	}
	if err = el.call(m, utils.LOGLEVEL_ALERT); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.NodeID, el.LogLevel).Alert(m)
			err = nil
		}
	}
	return
}

// Crit logs to EEs with critical level
func (el *ExportLogger) Crit(m string) (err error) {
	if el.LogLevel < utils.LOGLEVEL_CRITICAL {
		return nil
	}
	if err = el.call(m, utils.LOGLEVEL_CRITICAL); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.NodeID, el.LogLevel).Crit(m)
			err = nil
		}
	}
	return
}

// Debug logs to EEs with debug level
func (el *ExportLogger) Debug(m string) (err error) {
	if el.LogLevel < utils.LOGLEVEL_DEBUG {
		return nil
	}
	if err = el.call(m, utils.LOGLEVEL_DEBUG); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.NodeID, el.LogLevel).Debug(m)
			err = nil
		}
	}
	return
}

// Emerg logs to EEs with emergency level
func (el *ExportLogger) Emerg(m string) (err error) {
	if el.LogLevel < utils.LOGLEVEL_EMERGENCY {
		return nil
	}
	if err = el.call(m, utils.LOGLEVEL_EMERGENCY); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.NodeID, el.LogLevel).Emerg(m)
			err = nil
		}
	}
	return
}

// Err logs to EEs with error level
func (el *ExportLogger) Err(m string) (err error) {
	if el.LogLevel < utils.LOGLEVEL_ERROR {
		return nil
	}
	if err = el.call(m, utils.LOGLEVEL_ERROR); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.NodeID, el.LogLevel).Err(m)
			err = nil
		}
	}
	return
}

// Info logs to EEs with info level
func (el *ExportLogger) Info(m string) (err error) {
	if el.LogLevel < utils.LOGLEVEL_INFO {
		return nil
	}
	if err = el.call(m, utils.LOGLEVEL_INFO); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.NodeID, el.LogLevel).Info(m)
			err = nil
		}
	}
	return
}

// Notice logs to EEs with notice level
func (el *ExportLogger) Notice(m string) (err error) {
	if el.LogLevel < utils.LOGLEVEL_NOTICE {
		return nil
	}
	if err = el.call(m, utils.LOGLEVEL_NOTICE); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.NodeID, el.LogLevel).Notice(m)
			err = nil
		}
	}
	return
}

// Warning logs to EEs with warning level
func (el *ExportLogger) Warning(m string) (err error) {
	if el.LogLevel < utils.LOGLEVEL_WARNING {
		return nil
	}
	if err = el.call(m, utils.LOGLEVEL_WARNING); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.NodeID, el.LogLevel).Warning(m)
			err = nil
		}
	}
	return
}

func (el *ExportLogger) GetMeta() map[string]any {
	return map[string]any{
		utils.Tenant:         el.Tenant,
		utils.NodeID:         el.NodeID,
		utils.Level:          el.LogLevel,
		utils.Format:         el.Writer.Topic,
		utils.Conn:           el.Writer.Addr.String(),
		utils.FailedPostsDir: el.FldPostDir,
		utils.Attempts:       el.Writer.MaxAttempts,
	}
}
