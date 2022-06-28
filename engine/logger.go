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

package engine

import (
	"log/syslog"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/segmentio/kafka-go"
)

func NewLogger(loggerType, tenant, nodeID string, loggCfg *config.LoggerCfg) (utils.LoggerInterface, error) {
	switch loggerType {
	case utils.MetaKafka:
		return NewExportLogger(nodeID, tenant, loggCfg.Level, loggCfg.Opts), nil
	default:
		return utils.NewLogger(loggerType, nodeID, loggCfg.Level)
	}
}

// Logs to kafka
type ExportLogger struct {
	logLevel int
	fPost    *utils.FailoverPoster
	loggOpts *config.LoggerOptsCfg
	writer   *kafka.Writer
	nodeID   string
	tenant   string
}

func NewExportLogger(nodeID, tenant string, level int, opts *config.LoggerOptsCfg) (el *ExportLogger) {
	el = &ExportLogger{
		logLevel: level,
		loggOpts: opts,
		nodeID:   nodeID,
		tenant:   tenant,
		writer: &kafka.Writer{
			Addr:        kafka.TCP(opts.KafkaConn),
			Topic:       opts.KafkaTopic,
			MaxAttempts: opts.Attempts,
		},
	}
	return
}

func (el *ExportLogger) Close() (err error) {
	if el.writer != nil {
		err = el.writer.Close()
		el.writer = nil
	}
	return
}

func (el *ExportLogger) call(m string, level int) (err error) {
	eventExport := &utils.CGREvent{
		Tenant: el.tenant,
		Event: map[string]interface{}{
			utils.NodeID: el.nodeID,
			"Message":    m,
			"Severity":   level,
			"Timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	// event will be exported through kafka as json format
	var content []byte
	if content, err = utils.ToUnescapedJSON(eventExport); err != nil {
		return
	}
	if err = el.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(utils.GenUUID()),
		Value: content,
	}); err != nil {
		// if there are any errors in kafka, we will post in FailedPostDirectory
		el.fPost = utils.NewFailoverPoster()
		if err = el.fPost.AddMessage(el.loggOpts.FailedPostsDir,
			el.loggOpts.KafkaConn, el.loggOpts.KafkaTopic, eventExport); err != nil {
			return
		}
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
	return el.logLevel
}

// SetLogLevel changes the log level
func (el *ExportLogger) SetLogLevel(level int) {
	el.logLevel = level
}

// Alert logs to EEs with alert level
func (el *ExportLogger) Alert(m string) (err error) {
	if el.logLevel < utils.LOGLEVEL_ALERT {
		return nil
	}
	if err = el.call(m, utils.LOGLEVEL_ALERT); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.nodeID, el.logLevel).Alert(m)
			err = nil
		}
	}
	return //el.call(m, utils.LOGLEVEL_ALERT)
}

// Crit logs to EEs with critical level
func (el *ExportLogger) Crit(m string) (err error) {
	if el.logLevel < utils.LOGLEVEL_CRITICAL {
		return nil
	}
	if el.call(m, utils.LOGLEVEL_CRITICAL); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.nodeID, el.logLevel).Crit(m)
			err = nil
		}
	}
	return // el.call(m, utils.LOGLEVEL_CRITICAL)
}

// Debug logs to EEs with debug level
func (el *ExportLogger) Debug(m string) (err error) {
	if el.logLevel < utils.LOGLEVEL_DEBUG {
		return nil
	}
	if err = el.call(m, utils.LOGLEVEL_DEBUG); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.nodeID, el.logLevel).Debug(m)
			err = nil
		}
	}
	return // el.call(m, utils.LOGLEVEL_DEBUG)
}

// Emerg logs to EEs with emergency level
func (el *ExportLogger) Emerg(m string) (err error) {
	if el.logLevel < utils.LOGLEVEL_EMERGENCY {
		return nil
	}
	if err = el.call(m, utils.LOGLEVEL_EMERGENCY); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.nodeID, el.logLevel).Emerg(m)
			err = nil
		}
	}
	return // el.call(m, utils.LOGLEVEL_EMERGENCY)
}

// Err logs to EEs with error level
func (el *ExportLogger) Err(m string) (err error) {
	if el.logLevel < utils.LOGLEVEL_ERROR {
		return nil
	}
	if err = el.call(m, utils.LOGLEVEL_ERROR); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.nodeID, el.logLevel).Err(m)
			err = nil
		}
	}
	return // el.call(m, utils.LOGLEVEL_ERROR)
}

// Info logs to EEs with info level
func (el *ExportLogger) Info(m string) (err error) {
	if el.logLevel < utils.LOGLEVEL_INFO {
		return nil
	}
	if err = el.call(m, utils.LOGLEVEL_INFO); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.nodeID, el.logLevel).Info(m)
			err = nil
		}
	}
	return // el.call(m, utils.LOGLEVEL_INFO)
}

// Notice logs to EEs with notice level
func (el *ExportLogger) Notice(m string) (err error) {
	if el.logLevel < utils.LOGLEVEL_NOTICE {
		return nil
	}
	if err = el.call(m, utils.LOGLEVEL_NOTICE); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.nodeID, el.logLevel).Notice(m)
			err = nil
		}
	}
	return // el.call(m, utils.LOGLEVEL_NOTICE)
}

// Warning logs to EEs with warning level
func (el *ExportLogger) Warning(m string) (err error) {
	if el.logLevel < utils.LOGLEVEL_WARNING {
		return nil
	}
	if err = el.call(m, utils.LOGLEVEL_WARNING); err != nil {
		if err == utils.ErrLoggerChanged {
			utils.NewStdLogger(el.nodeID, el.logLevel).Warning(m)
			err = nil
		}
	}
	return // el.call(m, utils.LOGLEVEL_WARNING)
}
