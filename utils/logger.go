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

package utils

import (
	"fmt"
	"log"
	"log/syslog"
	"reflect"
	"runtime"
)

var Logger LoggerInterface
var nodeID string

func init() {
	if Logger == nil || reflect.ValueOf(Logger).IsNil() {
		err := Newlogger(MetaSysLog, nodeID)
		if err != nil {
			Logger.Err(fmt.Sprintf("Could not connect to syslog: %v", err))
		}
	}
}

//functie Newlogger (logger type)
func Newlogger(loggertype, id string) (err error) {
	Logger = new(StdLogger)
	nodeID = id
	var l *syslog.Writer
	if loggertype == MetaSysLog {
		if l, err = syslog.New(syslog.LOG_INFO, fmt.Sprintf("CGRateS <%s> ", nodeID)); err != nil {
			return err
		} else {
			Logger.SetSyslog(l)
		}
		return nil
	} else if loggertype != MetaStdLog {
		return fmt.Errorf("unsuported logger: <%s>", loggertype)
	}
	return nil
}

type LoggerInterface interface {
	SetSyslog(log *syslog.Writer)
	SetLogLevel(level int)
	GetSyslog() *syslog.Writer
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

// Logs to standard output
type StdLogger struct {
	logLevel int
	syslog   *syslog.Writer
}

func (sl *StdLogger) Close() (err error) {
	if sl.syslog != nil {
		sl.Close()
	}
	return
}
func (sl *StdLogger) Write(p []byte) (n int, err error) {
	s := string(p[:])
	fmt.Print(s)
	return 1, nil
}

//SetSyslog sets the logger for the server
func (sl *StdLogger) SetSyslog(l *syslog.Writer) {
	sl.syslog = l
}

// GetSyslog returns the logger for the server
func (sl *StdLogger) GetSyslog() *syslog.Writer {
	return sl.syslog
}

// SetLogLevel changes the log level
func (sl *StdLogger) SetLogLevel(level int) {
	sl.logLevel = level
}

// Alert logs to syslog with alert level
func (sl *StdLogger) Alert(m string) (err error) {
	if sl.logLevel < LOGLEVEL_ALERT {
		return
	}
	if sl.syslog != nil {
		sl.syslog.Alert(m)
	} else {
		log.Print("CGRateS <" + nodeID + "> [ALERT] " + m)
	}
	return
}

// Crit logs to syslog with critical level
func (sl *StdLogger) Crit(m string) (err error) {
	if sl.logLevel < LOGLEVEL_CRITICAL {
		return
	}
	if sl.syslog != nil {
		sl.syslog.Crit(m)
	} else {
		log.Print("CGRateS <" + nodeID + "> [CRITICAL] " + m)
	}
	return
}

// Debug logs to syslog with debug level
func (sl *StdLogger) Debug(m string) (err error) {
	if sl.logLevel < LOGLEVEL_DEBUG {
		return
	}
	if sl.syslog != nil {
		sl.syslog.Debug(m)
	} else {
		log.Print("CGRateS <" + nodeID + "> [DEBUG] " + m)
	}
	return
}

// Emerg logs to syslog with emergency level
func (sl *StdLogger) Emerg(m string) (err error) {
	if sl.logLevel < LOGLEVEL_EMERGENCY {
		return
	}
	if sl.syslog != nil {
		sl.syslog.Emerg(m)
	} else {
		log.Print("CGRateS <" + nodeID + "> [EMERGENCY] " + m)
	}
	return
}

// Err logs to syslog with error level
func (sl *StdLogger) Err(m string) (err error) {
	if sl.logLevel < LOGLEVEL_ERROR {
		return
	}
	if sl.syslog != nil {
		sl.syslog.Err(m)
	} else {
		log.Print("CGRateS <" + nodeID + "> [ERROR] " + m)
	}
	return
}

// Info logs to syslog with info level
func (sl *StdLogger) Info(m string) (err error) {
	if sl.logLevel < LOGLEVEL_INFO {
		return
	}
	if sl.syslog != nil {
		sl.syslog.Info(m)
	} else {
		log.Print("CGRateS <" + nodeID + "> [INFO] " + m)
	}
	return
}

// Notice logs to syslog with notice level
func (sl *StdLogger) Notice(m string) (err error) {
	if sl.logLevel < LOGLEVEL_NOTICE {
		return
	}
	if sl.syslog != nil {
		sl.syslog.Notice(m)
	} else {
		log.Print("CGRateS <" + nodeID + "> [NOTICE] " + m)
	}
	return
}

// Warning logs to syslog with warning level
func (sl *StdLogger) Warning(m string) (err error) {
	if sl.logLevel < LOGLEVEL_WARNING {
		return
	}

	if sl.syslog != nil {
		sl.syslog.Warning(m)
	} else {
		log.Print("CGRateS <" + nodeID + "> [WARNING] " + m)
	}
	return
}

// LogStack logs to syslog the stack trace using debug level
func LogStack() {
	buf := make([]byte, 300)
	runtime.Stack(buf, false)
	Logger.Debug(string(buf))
}
