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
	"bytes"
	"io"
	"log"
	syslog "log/syslog"
	"os"
	"strings"
	"testing"
)

func TestEmergLogger(t *testing.T) {
	output := new(bytes.Buffer)
	log.SetOutput(output)
	loggertype := MetaSysLog
	id := "id_emerg"
	if _, err := Newlogger(loggertype, id); err != nil {
		t.Error(err)
	}

	newLogger := &StdLogger{nodeID: id}

	newLogger.SetLogLevel(-1)
	if err := newLogger.Emerg(EmptyString); err != nil {
		t.Error(err)
	} else {
		newLogger.SetLogLevel(LOGLEVEL_EMERGENCY)
		if err := newLogger.Emerg("emergency_panic"); err != nil {
			t.Error(err)
		}
		expected := "CGRateS <id_emerg> [EMERGENCY] emergency_panic"
		if rcv := output.String(); !strings.Contains(rcv, expected) {
			t.Errorf("Expected %+v, received %+v", expected, rcv)
		}
	}

	log.SetOutput(os.Stderr)
}

func TestAlertLogger(t *testing.T) {
	output := new(bytes.Buffer)
	log.SetOutput(output)
	loggertype := MetaSysLog
	id := "id_alert"
	if _, err := Newlogger(loggertype, id); err != nil {
		t.Error(err)
	}

	newLogger := &StdLogger{nodeID: id}

	newLogger.SetLogLevel(0)
	if err := newLogger.Alert("Alert"); err != nil {
		t.Error(err)
	} else {
		newLogger.SetLogLevel(LOGLEVEL_ALERT)
		if err := newLogger.Alert("Alert"); err != nil {
			t.Error(err)
		}
		expected := "CGRateS <id_alert> [ALERT] Alert"
		if rcv := output.String(); !strings.Contains(rcv, expected) {
			t.Errorf("Expected %+v, received %+v", expected, rcv)
		}
	}

	log.SetOutput(os.Stderr)
}

func TestCritLogger(t *testing.T) {
	output := new(bytes.Buffer)
	log.SetOutput(output)

	loggertype := MetaSysLog
	id := "id_crit"
	if _, err := Newlogger(loggertype, id); err != nil {
		t.Error(err)
	}

	newLogger := &StdLogger{nodeID: id}

	newLogger.SetLogLevel(1)
	if err := newLogger.Crit("Critical_level"); err != nil {
		t.Error(err)
	} else {
		newLogger.logLevel = LOGLEVEL_CRITICAL
		if err := newLogger.Crit("Critical_level"); err != nil {
			t.Error(err)
		}
		expected := "CGRateS <id_crit> [CRITICAL] Critical_level"
		if rcv := output.String(); !strings.Contains(rcv, expected) {
			t.Errorf("Expected %+v, received %+v", expected, rcv)
		}
	}

	log.SetOutput(os.Stderr)
}

func TestErrorLogger(t *testing.T) {
	output := new(bytes.Buffer)
	log.SetOutput(output)

	loggertype := MetaSysLog
	id := "id_error"
	if _, err := Newlogger(loggertype, id); err != nil {
		t.Error(err)
	}

	newLogger := &StdLogger{nodeID: id}

	newLogger.SetLogLevel(2)
	if err := newLogger.Err("error_panic"); err != nil {
		t.Error(err)
	} else {
		newLogger.SetLogLevel(LOGLEVEL_ERROR)
		if err := newLogger.Err("error_panic"); err != nil {
			t.Error(err)
		}
		expected := "CGRateS <id_error> [ERROR] error_panic"
		if rcv := output.String(); !strings.Contains(rcv, expected) {
			t.Errorf("Expected %+v, received %+v", expected, rcv)
		}
	}

	log.SetOutput(os.Stderr)
}

func TestWarningLogger(t *testing.T) {
	output := new(bytes.Buffer)
	log.SetOutput(output)

	loggertype := MetaSysLog
	id := "id_error"
	if _, err := Newlogger(loggertype, id); err != nil {
		t.Error(err)
	}

	newLogger := &StdLogger{nodeID: id}

	newLogger.SetLogLevel(3)
	if err := newLogger.Warning("warning_panic"); err != nil {
		t.Error(err)
	} else {
		newLogger.SetLogLevel(LOGLEVEL_WARNING)
		if err := newLogger.Warning("warning_panic"); err != nil {
			t.Error(err)
		}
		expected := "CGRateS <id_error> [WARNING] warning_panic"
		if rcv := output.String(); !strings.Contains(rcv, expected) {
			t.Errorf("Expected %+v, received %+v", expected, rcv)
		}
	}

	log.SetOutput(os.Stderr)
}

func TestNoticeLogger(t *testing.T) {
	output := new(bytes.Buffer)
	log.SetOutput(output)

	loggertype := MetaSysLog
	id := "id_notice"
	if _, err := Newlogger(loggertype, id); err != nil {
		t.Error(err)
	}

	newLogger := &StdLogger{nodeID: id}

	newLogger.SetLogLevel(4)
	if err := newLogger.Notice("notice_panic"); err != nil {
		t.Error(err)
	} else {
		newLogger.SetLogLevel(LOGLEVEL_NOTICE)
		if err := newLogger.Notice("notice_panic"); err != nil {
			t.Error(err)
		}
		expected := "CGRateS <id_notice> [NOTICE] notice_panic"
		if rcv := output.String(); !strings.Contains(rcv, expected) {
			t.Errorf("Expected %+v, received %+v", expected, rcv)
		}
	}

	log.SetOutput(os.Stderr)
}

func TestInfoLogger(t *testing.T) {
	output := new(bytes.Buffer)
	log.SetOutput(output)

	loggertype := MetaSysLog
	id := "id_info"
	if _, err := Newlogger(loggertype, id); err != nil {
		t.Error(err)
	}

	newLogger := &StdLogger{nodeID: id}

	newLogger.SetLogLevel(5)
	if err := newLogger.Info("info_panic"); err != nil {
		t.Error(err)
	} else {
		newLogger.SetLogLevel(LOGLEVEL_INFO)
		if err := newLogger.Info("info_panic"); err != nil {
			t.Error(err)
		}
		expected := "CGRateS <id_info> [INFO] info_panic"
		if rcv := output.String(); !strings.Contains(rcv, expected) {
			t.Errorf("Expected %+v, received %+v", expected, rcv)
		}
	}

	log.SetOutput(os.Stderr)
}

func TestDebugLogger(t *testing.T) {
	output := new(bytes.Buffer)
	log.SetOutput(output)

	loggertype := MetaSysLog
	id := "id_debug"
	if _, err := Newlogger(loggertype, id); err != nil {
		t.Error(err)
	}

	newLogger, err := Newlogger(MetaStdLog, id)
	if err != nil {
		t.Error(err)
	}
	newLogger.SetLogLevel(6)
	if err := newLogger.Debug("debug_panic"); err != nil {
		t.Error(err)
	} else {
		newLogger.SetLogLevel(LOGLEVEL_DEBUG)
		if err := newLogger.Debug("debug_panic"); err != nil {
			t.Error(err)
		}
		newLogger.GetSyslog()
		expected := "GRateS <id_debug> [DEBUG] debug_panic"
		if rcv := output.String(); !strings.Contains(rcv, expected) {
			t.Errorf("Expected %+v, received %+v", expected, rcv)
		}
	}

	log.SetOutput(os.Stderr)
}

func TestWriteLogger(t *testing.T) {
	log.SetOutput(os.Stderr)

	loggertype := MetaSysLog
	id := "id_write"
	if _, err := Newlogger(loggertype, id); err != nil {
		t.Error(err)
	}

	newWriter, err := syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, "CGRates id_write")
	if err != nil {
		t.Error(err)
	}
	newLogger := &StdLogger{nodeID: id}
	newLogger.SetSyslog(newWriter)

	if n, err := newLogger.Write([]byte(EmptyString)); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Errorf("Expected 1, received %+v", n)
	}

	log.SetOutput(os.Stderr)
}

func TestCloseLogger(t *testing.T) {
	log.SetOutput(io.Discard)

	loggertype := MetaStdLog
	if _, err := Newlogger(loggertype, EmptyString); err != nil {
		t.Error(err)
	}

	newLogger := &StdLogger{nodeID: EmptyString}
	if err := newLogger.Close(); err != nil {
		t.Error(err)
	}
	newWriter, err := syslog.New(0, "CGRates")
	if err != nil {
		t.Error(err)
	}
	x := newLogger.GetSyslog()
	x = newWriter

	newLogger.SetSyslog(x)
	if err := newLogger.Close(); err != nil {
		t.Error(err)
	}
}

func TestLogStackLogger(t *testing.T) {
	LogStack()
}

func TestNewLoggerInvalidLoggerType(t *testing.T) {
	log.SetOutput(io.Discard)

	loggertype := "Invalid_TYPE"
	expected := "unsuported logger: <Invalid_TYPE>"
	if _, err := Newlogger(loggertype, EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}
