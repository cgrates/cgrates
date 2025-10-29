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
	"bytes"
	"io"
	"log"
	"log/syslog"
	"os"
	"strings"
	"testing"
)

func TestEmergLogger(t *testing.T) {
	output := new(bytes.Buffer)
	log.SetOutput(output)
	id := "id_emerg"

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
	id := "id_alert"

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

	id := "id_crit"

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

	id := "id_error"

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

	id := "id_error"

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

	id := "id_notice"

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

	id := "id_info"

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

	id := "id_debug"

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
		expected := "GRateS <id_debug> [DEBUG] debug_panic"
		if rcv := output.String(); !strings.Contains(rcv, expected) {
			t.Errorf("Expected %+v, received %+v", expected, rcv)
		}
	}

	log.SetOutput(os.Stderr)
}

func TestWriteLogger(t *testing.T) {
	log.SetOutput(os.Stderr)

	id := "id_write"

	newLogger := &StdLogger{nodeID: id}

	if n, err := newLogger.Write([]byte(EmptyString)); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Errorf("Expected 1, received %+v", n)
	}

	log.SetOutput(os.Stderr)
}

func TestCloseLoggerStdLog(t *testing.T) {
	log.SetOutput(io.Discard)

	loggertype := MetaStdLog
	if _, err := Newlogger(loggertype, EmptyString); err != nil {
		t.Error(err)
	}

	newLogger := &StdLogger{nodeID: EmptyString}
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
	expected := "unsupported logger: <Invalid_TYPE>"
	if _, err := Newlogger(loggertype, EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoggerClose(t *testing.T) {
	nm := 1
	str := "test"
	sl := &StdLogger{
		logLevel: nm,
		nodeID:   str,
		syslog:   &syslog.Writer{},
	}

	err := sl.Close()
	if err != nil {
		t.Error(err)
	}
}

func TestLoggerAlert(t *testing.T) {
	nm := 100
	str := "test"
	sl := &StdLogger{
		logLevel: nm,
		nodeID:   str,
		syslog:   &syslog.Writer{},
	}

	err := sl.Alert(str)
	if err != nil {
		t.Error(err)
	}
}

func TestLoggerCrit(t *testing.T) {
	nm := 100
	str := "test"
	sl := &StdLogger{
		logLevel: nm,
		nodeID:   str,
		syslog:   &syslog.Writer{},
	}

	err := sl.Crit(str)
	if err != nil {
		t.Error(err)
	}
}

func TestLoggerDebug(t *testing.T) {
	nm := 100
	str := "test"
	sl := &StdLogger{
		logLevel: nm,
		nodeID:   str,
		syslog:   &syslog.Writer{},
	}

	err := sl.Debug(str)
	if err != nil {
		t.Error(err)
	}
}

//Loggs: Broadcast message from systemd-journald@debian (Tue 2023-09-12 10:09:31 CEST):
/*func TestLoggerEmerg(t *testing.T) {
	nm := 100
	str := "test"
	sl := &StdLogger{
		logLevel: nm,
		nodeID:   str,
		syslog:   &syslog.Writer{},
	}

	err := sl.Emerg(str)
	if err != nil {
		t.Error(err)
	}
}*/

func TestLoggerErr(t *testing.T) {
	nm := 100
	str := "test"
	sl := &StdLogger{
		logLevel: nm,
		nodeID:   str,
		syslog:   &syslog.Writer{},
	}

	err := sl.Err(str)
	if err != nil {
		t.Error(err)
	}
}

func TestLoggerInfo(t *testing.T) {
	nm := 100
	str := "test"
	sl := &StdLogger{
		logLevel: nm,
		nodeID:   str,
		syslog:   &syslog.Writer{},
	}

	err := sl.Info(str)
	if err != nil {
		t.Error(err)
	}
}

func TestLoggerNotice(t *testing.T) {
	nm := 100
	str := "test"
	sl := &StdLogger{
		logLevel: nm,
		nodeID:   str,
		syslog:   &syslog.Writer{},
	}

	err := sl.Notice(str)
	if err != nil {
		t.Error(err)
	}
}

func TestLoggerWarning(t *testing.T) {
	nm := 100
	str := "test"
	sl := &StdLogger{
		logLevel: nm,
		nodeID:   str,
		syslog:   &syslog.Writer{},
	}

	err := sl.Warning(str)
	if err != nil {
		t.Error(err)
	}
}
