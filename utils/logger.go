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
	"runtime"
)

var Logger LoggerInterface

func init() {
	Logger = new(StdLogger)

	// Attempt to connect to syslog. We'll fallback to `log` otherwise.
	var err error
	var l *syslog.Writer
	l, err = syslog.New(syslog.LOG_INFO, "CGRateS")
	if err != nil {
		Logger.Err(fmt.Sprintf("Could not connect to syslog: %v", err))
	} else {
		Logger.SetSyslog(l)
	}
}

type LoggerInterface interface {
	SetSyslog(log *syslog.Writer)
	SetLogLevel(level int)

	GetSyslog() *syslog.Writer

	Alert(m string) error
	Close() error
	Crit(m string) error
	Debug(m string) error
	Emerg(m string) error
	Err(m string) error
	Info(m string) error
	Notice(m string) error
	Warning(m string) error
}

const (
	LOGLEVEL_DEBUG     = 9
	LOGLEVEL_INFO      = 8
	LOGLEVEL_NOTICE    = 7
	LOGLEVEL_WARNING   = 6
	LOGLEVEL_ERROR     = 5
	LOGLEVEL_CRITICAL  = 4
	LOGLEVEL_EMERGENCY = 3
	LOGLEVEL_ALERT     = 2
)

// Logs to standard output
type StdLogger struct {
	logLevel int
	syslog   *syslog.Writer
}

func (sl *StdLogger) SetSyslog(l *syslog.Writer) {
	sl.syslog = l
}
func (sl *StdLogger) GetSyslog() *syslog.Writer {
	return sl.syslog
}
func (sl *StdLogger) SetLogLevel(level int) {
	sl.logLevel = level
}
func (sl *StdLogger) Alert(m string) (err error) {
	if sl.logLevel < LOGLEVEL_ALERT {
		return
	}

	if sl.syslog != nil {
		sl.syslog.Alert(m)
	} else {
		log.Print("[ALERT]" + m)
	}
	return
}
func (sl *StdLogger) Close() (err error) {
	if sl.syslog != nil {
		sl.Close()
	}
	return
}
func (sl *StdLogger) Crit(m string) (err error) {
	if sl.logLevel < LOGLEVEL_CRITICAL {
		return
	}

	if sl.syslog != nil {
		sl.syslog.Crit(m)
	} else {
		log.Print("[CRITICAL]" + m)
	}
	return
}
func (sl *StdLogger) Debug(m string) (err error) {
	if sl.logLevel < LOGLEVEL_DEBUG {
		return
	}

	if sl.syslog != nil {
		sl.syslog.Debug(m)
	} else {
		log.Print("[DEBUG]" + m)
	}
	return
}
func (sl *StdLogger) Emerg(m string) (err error) {
	if sl.logLevel < LOGLEVEL_EMERGENCY {
		return
	}

	if sl.syslog != nil {
		sl.syslog.Emerg(m)
	} else {
		log.Print("[EMERGENCY]" + m)
	}
	return
}
func (sl *StdLogger) Err(m string) (err error) {
	if sl.logLevel < LOGLEVEL_ERROR {
		return
	}

	if sl.syslog != nil {
		sl.syslog.Err(m)
	} else {
		log.Print("[ERROR]" + m)
	}
	return
}
func (sl *StdLogger) Info(m string) (err error) {
	if sl.logLevel < LOGLEVEL_INFO {
		return
	}

	if sl.syslog != nil {
		sl.syslog.Info(m)
	} else {
		log.Print("[INFO]" + m)
	}
	return
}
func (sl *StdLogger) Notice(m string) (err error) {
	if sl.logLevel < LOGLEVEL_NOTICE {
		return
	}

	if sl.syslog != nil {
		sl.syslog.Notice(m)
	} else {
		log.Print("[NOTICE]" + m)
	}
	return
}
func (sl *StdLogger) Warning(m string) (err error) {
	if sl.logLevel < LOGLEVEL_WARNING {
		return
	}

	if sl.syslog != nil {
		sl.syslog.Warning(m)
	} else {
		log.Print("[WARNING]" + m)
	}
	return
}

func LogStack() {
	buf := make([]byte, 300)
	runtime.Stack(buf, false)
	Logger.Debug(string(buf))
}
