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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func NewLogger(loggerType, tenant, nodeID string, level int, connMgr *ConnManager,
	eesConns []string) (utils.LoggerInterface, error) {
	switch loggerType {
	case utils.MetaExportLog:
		return NewExportLogger(nodeID, tenant, level, connMgr, eesConns), nil
	default:
		return utils.NewLogger(loggerType, nodeID, level)
	}
}

// Logs to EEs
type ExportLogger struct {
	connMgr  *ConnManager
	eesConns []string
	logLevel int
	nodeID   string
	tenant   string
}

func NewExportLogger(nodeID, tenant string, level int, connMgr *ConnManager,
	eesConns []string) (el *ExportLogger) {
	el = &ExportLogger{
		connMgr:  connMgr,
		eesConns: eesConns,
		logLevel: level,
		nodeID:   nodeID,
		tenant:   tenant,
	}
	return
}

func (el *ExportLogger) Close() (_ error) {
	return
}

func (el *ExportLogger) call(m string, level int) error {
	var reply map[string]map[string]interface{}
	return el.connMgr.Call(context.Background(), el.eesConns, utils.EeSv1ProcessEvent, &utils.CGREventWithEeIDs{
		CGREvent: &utils.CGREvent{
			Tenant: el.tenant,
			Event: map[string]interface{}{
				utils.NodeID: el.nodeID,
				"Message":    m,
				"Severity":   level,
			}}}, &reply)
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
func (el *ExportLogger) Alert(m string) error {
	if el.logLevel < utils.LOGLEVEL_ALERT {
		return nil
	}
	return el.call(m, utils.LOGLEVEL_ALERT)
}

// Crit logs to EEs with critical level
func (el *ExportLogger) Crit(m string) error {
	if el.logLevel < utils.LOGLEVEL_CRITICAL {
		return nil
	}
	return el.call(m, utils.LOGLEVEL_CRITICAL)
}

// Debug logs to EEs with debug level
func (el *ExportLogger) Debug(m string) error {
	if el.logLevel < utils.LOGLEVEL_DEBUG {
		return nil
	}
	return el.call(m, utils.LOGLEVEL_DEBUG)
}

// Emerg logs to EEs with emergency level
func (el *ExportLogger) Emerg(m string) error {
	if el.logLevel < utils.LOGLEVEL_EMERGENCY {
		return nil
	}
	return el.call(m, utils.LOGLEVEL_EMERGENCY)
}

// Err logs to EEs with error level
func (el *ExportLogger) Err(m string) error {
	if el.logLevel < utils.LOGLEVEL_ERROR {
		return nil
	}
	return el.call(m, utils.LOGLEVEL_ERROR)
}

// Info logs to EEs with info level
func (el *ExportLogger) Info(m string) error {
	if el.logLevel < utils.LOGLEVEL_INFO {
		return nil
	}
	return el.call(m, utils.LOGLEVEL_INFO)
}

// Notice logs to EEs with notice level
func (el *ExportLogger) Notice(m string) error {
	if el.logLevel < utils.LOGLEVEL_NOTICE {
		return nil
	}
	return el.call(m, utils.LOGLEVEL_NOTICE)
}

// Warning logs to EEs with warning level
func (el *ExportLogger) Warning(m string) error {
	if el.logLevel < utils.LOGLEVEL_WARNING {
		return nil
	}
	return el.call(m, utils.LOGLEVEL_WARNING)
}
