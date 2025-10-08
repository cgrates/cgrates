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

package utils

import (
	"fmt"
	"io"
	"log"
	"log/syslog"
	"os"
)

var Logger LoggerInterface
var noSysLog bool

func init() {
	var err error
	Logger, err = NewSysLogger(EmptyString, 0)
	if err != nil {
		Logger = NewStdLogger(EmptyString, 0)
	}
}

// log severities following rfc3164
const (
	LOGLEVEL_EMERGENCY = iota
	LOGLEVEL_ALERT
	LOGLEVEL_CRITICAL
	LOGLEVEL_ERROR
	LOGLEVEL_WARNING
	LOGLEVEL_NOTICE
	LOGLEVEL_INFO
	LOGLEVEL_DEBUG
)

type LoggerInterface interface {
	GetSyslog() *syslog.Writer
	SetLogLevel(level int)
	GetLogLevel() int
	Close() error
	Emerg(m string) error
	Alert(m string) error
	Crit(m string) error
	Err(m string) error
	Warning(m string) error
	Notice(m string) error
	Info(m string) error
	Debug(m string) error
	Write(p []byte) (n int, err error)
}

type SysLogger struct {
	logLevel int
	syslog   *syslog.Writer
}

func NewSysLogger(nodeID string, level int) (logger *SysLogger, err error) {
	var l *syslog.Writer
	l, err = syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, fmt.Sprintf("CGRateS <%s>", nodeID))
	logger = &SysLogger{
		logLevel: level,
		syslog:   l,
	}
	return
}

func (sl *SysLogger) GetSyslog() *syslog.Writer {
	return sl.syslog
}

func (sl *SysLogger) Close() (err error) {
	return sl.syslog.Close()
}

func (sl *SysLogger) Write(p []byte) (n int, err error) {
	return sl.syslog.Write(p)
}

// GetLogLevel() returns the level logger number for the server
func (sl *SysLogger) GetLogLevel() int {
	return sl.logLevel
}

// SetLogLevel changes the log level
func (sl *SysLogger) SetLogLevel(level int) {
	sl.logLevel = level
}

// Alert logs to syslog with alert level
func (sl *SysLogger) Alert(m string) (_ error) {
	if sl.logLevel < LOGLEVEL_ALERT {
		return
	}
	return sl.syslog.Alert(m)
}

// Crit logs to syslog with critical level
func (sl *SysLogger) Crit(m string) (_ error) {
	if sl.logLevel < LOGLEVEL_CRITICAL {
		return
	}
	return sl.syslog.Crit(m)
}

// Debug logs to syslog with debug level
func (sl *SysLogger) Debug(m string) (_ error) {
	if sl.logLevel < LOGLEVEL_DEBUG {
		return
	}
	return sl.syslog.Debug(m)
}

// Emerg logs to syslog with emergency level
func (sl *SysLogger) Emerg(m string) (_ error) {
	if sl.logLevel < LOGLEVEL_EMERGENCY {
		return
	}
	return sl.syslog.Emerg(m)
}

// Err logs to syslog with error level
func (sl *SysLogger) Err(m string) (_ error) {
	if sl.logLevel < LOGLEVEL_ERROR {
		return
	}
	return sl.syslog.Err(m)
}

// Info logs to syslog with info level
func (sl *SysLogger) Info(m string) (_ error) {
	if sl.logLevel < LOGLEVEL_INFO {
		return
	}
	return sl.syslog.Info(m)
}

// Notice logs to syslog with notice level
func (sl *SysLogger) Notice(m string) (_ error) {
	if sl.logLevel < LOGLEVEL_NOTICE {
		return
	}
	return sl.syslog.Notice(m)
}

// Warning logs to syslog with warning level
func (sl *SysLogger) Warning(m string) (_ error) {
	if sl.logLevel < LOGLEVEL_WARNING {
		return
	}
	return sl.syslog.Warning(m)
}

type StdLogger struct {
	w        io.WriteCloser
	logLevel int
	nodeID   string
}

type NopCloser struct {
	io.Writer
}

func (NopCloser) Close() error { return nil }

type logWriter struct {
	*log.Logger
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	err = l.Output(2, string(p))
	return
}
func (*logWriter) Close() error { return nil }

func NewStdLogger(nodeID string, level int) *StdLogger {
	return &StdLogger{
		nodeID:   nodeID,
		logLevel: level,
		w: &logWriter{
			log.New(os.Stderr, EmptyString, log.LstdFlags),
		},
	}
}

func NewStdLoggerWithWriter(w io.Writer, nodeID string, level int) *StdLogger {
	return &StdLogger{
		nodeID:   nodeID,
		logLevel: level,
		w:        NopCloser{w},
	}
}

func (sl *StdLogger) GetSyslog() *syslog.Writer {
	return nil
}

func (sl *StdLogger) Close() (err error) {
	return sl.w.Close()
}
func (sl *StdLogger) Write(p []byte) (n int, err error) {
	return sl.w.Write(p)
}

// GetLogLevel() returns the level logger number for the server
func (sl *StdLogger) GetLogLevel() int {
	return sl.logLevel
}

// SetLogLevel changes the log level
func (sl *StdLogger) SetLogLevel(level int) {
	sl.logLevel = level
}

// Alert logs to stderr with alert level
func (sl *StdLogger) Alert(m string) (err error) {
	if sl.logLevel < LOGLEVEL_ALERT {
		return
	}
	_, err = fmt.Fprintf(sl.w, "CGRateS <%s> [ALERT] %s\n", sl.nodeID, m)
	return
}

// Crit logs to stderr with critical level
func (sl *StdLogger) Crit(m string) (err error) {
	if sl.logLevel < LOGLEVEL_CRITICAL {
		return
	}
	_, err = fmt.Fprintf(sl.w, "CGRateS <%s> [CRITICAL] %s\n", sl.nodeID, m)
	return
}

// Debug logs to stderr with debug level
func (sl *StdLogger) Debug(m string) (err error) {
	if sl.logLevel < LOGLEVEL_DEBUG {
		return
	}
	_, err = fmt.Fprintf(sl.w, "CGRateS <%s> [DEBUG] %s\n", sl.nodeID, m)
	return
}

// Emerg logs to stderr with emergency level
func (sl *StdLogger) Emerg(m string) (err error) {
	if sl.logLevel < LOGLEVEL_EMERGENCY {
		return
	}
	_, err = fmt.Fprintf(sl.w, "CGRateS <%s> [EMERGENCY] %s\n", sl.nodeID, m)
	return
}

// Err logs to stderr with error level
func (sl *StdLogger) Err(m string) (err error) {
	if sl.logLevel < LOGLEVEL_ERROR {
		return
	}
	_, err = fmt.Fprintf(sl.w, "CGRateS <%s> [ERROR] %s\n", sl.nodeID, m)
	return
}

// Info logs to stderr with info level
func (sl *StdLogger) Info(m string) (err error) {
	if sl.logLevel < LOGLEVEL_INFO {
		return
	}
	_, err = fmt.Fprintf(sl.w, "CGRateS <%s> [INFO] %s\n", sl.nodeID, m)
	return
}

// Notice logs to stderr with notice level
func (sl *StdLogger) Notice(m string) (err error) {
	if sl.logLevel < LOGLEVEL_NOTICE {
		return
	}
	_, err = fmt.Fprintf(sl.w, "CGRateS <%s> [NOTICE] %s\n", sl.nodeID, m)
	return
}

// Warning logs to stderr with warning level
func (sl *StdLogger) Warning(m string) (err error) {
	if sl.logLevel < LOGLEVEL_WARNING {
		return
	}
	_, err = fmt.Fprintf(sl.w, "CGRateS <%s> [WARNING] %s\n", sl.nodeID, m)
	return
}
