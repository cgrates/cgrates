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
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
)

func TestLoggerNewLoggerStdoutOK(t *testing.T) {
	exp := &StdLogger{
		nodeID:   "1234",
		logLevel: 7,
		w: &logWriter{
			log.New(os.Stderr, EmptyString, log.LstdFlags),
		},
	}
	if rcv, err := NewLogger(MetaStdLog, "1234", 7); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestLoggerNewLoggerSyslogOK(t *testing.T) {
	exp := &SysLogger{
		logLevel: 7,
	}
	if rcv, err := NewLogger(MetaSysLog, EmptyString, 7); err != nil {
		t.Error(err)
	} else {
		exp.syslog = rcv.GetSyslog()
		if !reflect.DeepEqual(rcv, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
		}
	}
}

func TestLoggerNewLoggerUnsupported(t *testing.T) {
	experr := `unsupported logger: <unsupported>`
	if _, err := NewLogger("unsupported", EmptyString, 7); err == nil || err.Error() != experr {
		t.Errorf("expected: <%s>, \nreceived: <%+v>", experr, err)
	}
}

func TestLoggerSysloggerSetGetLogLevel(t *testing.T) {
	sl, err := NewSysLogger("1234", 4)
	if err != nil {
		t.Error(err)
	}
	if rcv := sl.GetLogLevel(); rcv != 4 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 4, rcv)
	}
	sl.SetLogLevel(5)
	if rcv := sl.GetLogLevel(); rcv != 5 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 5, rcv)
	}
}

func TestLoggerStdLoggerEmerg(t *testing.T) {
	logMsg := "log message"
	var buf bytes.Buffer
	sl := NewStdLoggerWithWriter(&buf, "1234", -1)
	if err := sl.Emerg(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != "" {
		t.Error("did not expect the message to be logged")
	}

	sl.SetLogLevel(LOGLEVEL_EMERGENCY)
	expMsg := fmt.Sprintf("CGRateS <%s> [EMERGENCY] %s\n", sl.nodeID, logMsg)
	if err := sl.Emerg(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != expMsg {
		t.Errorf("expected: <%s>, \nreceived: <%s>", expMsg, buf.String())
	}
}

func TestLoggerStdLoggerAlert(t *testing.T) {
	logMsg := "log message"
	var buf bytes.Buffer
	sl := NewStdLoggerWithWriter(&buf, "1234", -1)
	if err := sl.Alert(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != "" {
		t.Error("did not expect the message to be logged")
	}

	sl.SetLogLevel(LOGLEVEL_ALERT)
	expMsg := fmt.Sprintf("CGRateS <%s> [ALERT] %s\n", sl.nodeID, logMsg)
	if err := sl.Alert(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != expMsg {
		t.Errorf("expected: <%s>, \nreceived: <%s>", expMsg, buf.String())
	}
}

func TestLoggerStdLoggerCrit(t *testing.T) {
	logMsg := "log message"
	var buf bytes.Buffer
	sl := NewStdLoggerWithWriter(&buf, "1234", -1)
	if err := sl.Crit(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != "" {
		t.Error("did not expect the message to be logged")
	}

	sl.SetLogLevel(LOGLEVEL_CRITICAL)
	expMsg := fmt.Sprintf("CGRateS <%s> [CRITICAL] %s\n", sl.nodeID, logMsg)
	if err := sl.Crit(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != expMsg {
		t.Errorf("expected: <%s>, \nreceived: <%s>", expMsg, buf.String())
	}
}

func TestLoggerStdLoggerErr(t *testing.T) {
	logMsg := "log message"
	var buf bytes.Buffer
	sl := NewStdLoggerWithWriter(&buf, "1234", -1)
	if err := sl.Err(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != "" {
		t.Error("did not expect the message to be logged")
	}

	sl.SetLogLevel(LOGLEVEL_ERROR)
	expMsg := fmt.Sprintf("CGRateS <%s> [ERROR] %s\n", sl.nodeID, logMsg)
	if err := sl.Err(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != expMsg {
		t.Errorf("expected: <%s>, \nreceived: <%s>", expMsg, buf.String())
	}
}

func TestLoggerStdLoggerWarning(t *testing.T) {
	logMsg := "log message"
	var buf bytes.Buffer
	sl := NewStdLoggerWithWriter(&buf, "1234", -1)
	if err := sl.Warning(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != "" {
		t.Error("did not expect the message to be logged")
	}

	sl.SetLogLevel(LOGLEVEL_WARNING)
	expMsg := fmt.Sprintf("CGRateS <%s> [WARNING] %s\n", sl.nodeID, logMsg)
	if err := sl.Warning(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != expMsg {
		t.Errorf("expected: <%s>, \nreceived: <%s>", expMsg, buf.String())
	}
}

func TestLoggerStdLoggerNotice(t *testing.T) {
	logMsg := "log message"
	var buf bytes.Buffer
	sl := NewStdLoggerWithWriter(&buf, "1234", -1)
	if err := sl.Notice(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != "" {
		t.Error("did not expect the message to be logged")
	}

	sl.SetLogLevel(LOGLEVEL_NOTICE)
	expMsg := fmt.Sprintf("CGRateS <%s> [NOTICE] %s\n", sl.nodeID, logMsg)
	if err := sl.Notice(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != expMsg {
		t.Errorf("expected: <%s>, \nreceived: <%s>", expMsg, buf.String())
	}
}

func TestLoggerStdLoggerInfo(t *testing.T) {
	logMsg := "log message"
	var buf bytes.Buffer
	sl := NewStdLoggerWithWriter(&buf, "1234", -1)
	if err := sl.Info(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != "" {
		t.Error("did not expect the message to be logged")
	}

	sl.SetLogLevel(LOGLEVEL_INFO)
	expMsg := fmt.Sprintf("CGRateS <%s> [INFO] %s\n", sl.nodeID, logMsg)
	if err := sl.Info(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != expMsg {
		t.Errorf("expected: <%s>, \nreceived: <%s>", expMsg, buf.String())
	}
}

func TestLoggerStdLoggerDebug(t *testing.T) {
	logMsg := "log message"
	var buf bytes.Buffer
	sl := NewStdLoggerWithWriter(&buf, "1234", -1)
	if err := sl.Debug(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != "" {
		t.Error("did not expect the message to be logged")
	}

	sl.SetLogLevel(LOGLEVEL_DEBUG)
	expMsg := fmt.Sprintf("CGRateS <%s> [DEBUG] %s\n", sl.nodeID, logMsg)
	if err := sl.Debug(logMsg); err != nil {
		t.Error(err)
	} else if buf.String() != expMsg {
		t.Errorf("expected: <%s>, \nreceived: <%s>", expMsg, buf.String())
	}
}
