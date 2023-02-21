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
	"io"
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
	if noSysLog {
		t.SkipNow()
	}
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
	if noSysLog {
		t.SkipNow()
	}
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

func TestCloseSysLogger(t *testing.T) {
	if noSysLog {
		t.SkipNow()
	}
	sl, _ := NewSysLogger("test", 2)

	if err := sl.Close(); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}

}
func TestWriteSysLogger(t *testing.T) {
	if noSysLog {
		t.SkipNow()
	}
	sl, _ := NewSysLogger("test", 2)
	exp := 6
	testbyte := []byte{97, 98, 99, 100, 101, 102}
	if rcv, err := sl.Write(testbyte); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%v>, received: <%v>", exp, rcv)
	}

}

func TestAlertSysLogger(t *testing.T) {
	if noSysLog {
		t.SkipNow()
	}
	sl, _ := NewSysLogger("test", 0)

	if err := sl.Alert("Alert Message"); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}
	sl, _ = NewSysLogger("test2", 2)
	if err := sl.Alert("Alert Message 2"); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}
}
func TestCritSysLogger(t *testing.T) {
	if noSysLog {
		t.SkipNow()
	}
	sl, _ := NewSysLogger("test", 1)

	if err := sl.Crit("Critical Message"); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}
	sl, _ = NewSysLogger("test2", 4)
	if err := sl.Crit("Critical Message 2"); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}
}

func TestDebugSysLogger(t *testing.T) {
	if noSysLog {
		t.SkipNow()
	}
	sl, _ := NewSysLogger("test", 6)

	if err := sl.Debug("Debug Message"); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}
	sl, _ = NewSysLogger("test2", 8)
	if err := sl.Debug("Debug Message 2"); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}
}

func TestEmergSysLogger(t *testing.T) {
	if noSysLog {
		t.SkipNow()
	}
	sl, _ := NewSysLogger("test", -1)

	if err := sl.Emerg("Emergency Message"); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}
	// always broadcasts message from journal
	// sl, _ = NewSysLogger("test2", 1)
	// if err := sl.Emerg("Emergency Message 2"); err != nil {
	// 	t.Errorf("Expected <nil>, received %v", err)
	// }

}

func TestErrSysLogger(t *testing.T) {
	if noSysLog {
		t.SkipNow()
	}
	sl, _ := NewSysLogger("test", 2)

	if err := sl.Err("Error Message"); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}
	sl, _ = NewSysLogger("test2", 4)
	if err := sl.Err("Error Message 2"); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}
}
func TestInfoSysLogger(t *testing.T) {
	if noSysLog {
		t.SkipNow()
	}
	sl, _ := NewSysLogger("test", 5)

	if err := sl.Info("Info Message"); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}
	sl, _ = NewSysLogger("test2", 7)
	if err := sl.Info("Info Message 2"); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}
}
func TestNoticeSysLogger(t *testing.T) {
	if noSysLog {
		t.SkipNow()
	}
	sl, _ := NewSysLogger("test", 4)

	if err := sl.Notice("Notice Message"); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}
	sl, _ = NewSysLogger("test2", 6)
	if err := sl.Notice("Notice Message 2"); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}
}
func TestWarningSysLogger(t *testing.T) {
	if noSysLog {
		t.SkipNow()
	}
	sl, _ := NewSysLogger("test", 3)

	if err := sl.Warning("Warning Message"); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}
	sl, _ = NewSysLogger("test2", 5)
	if err := sl.Warning("Warning Message 2"); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}
}

func TestCloseNopCloser(t *testing.T) {
	var nC NopCloser

	if err := nC.Close(); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}

}

func TestWriteLogWriter(t *testing.T) {
	l := &logWriter{
		log.New(io.Discard, EmptyString, log.LstdFlags),
	}
	exp := 1
	if rcv, err := l.Write([]byte{51}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%+v> <%T>, received <%+v> <%T>", exp, exp, rcv, rcv)
	}
}

func TestCloseLogWriter(t *testing.T) {
	var lW logWriter
	if err := lW.Close(); err != nil {
		t.Errorf("Expected <nil>, received %v", err)
	}

}

func TestGetSyslogStdLogger(t *testing.T) {
	sl := &StdLogger{}
	if rcv := sl.GetSyslog(); rcv != nil {
		t.Errorf("Expected <nil>, received %v", rcv)
	}
}
func TestCloseStdLogger(t *testing.T) {
	sl := &StdLogger{w: Logger}
	if rcv := sl.Close(); rcv != nil {
		t.Errorf("Expected <nil>, received %v", rcv)
	}
}

func TestWriteStdLogger(t *testing.T) {
	sl := &StdLogger{
		w: &logWriter{
			log.New(io.Discard, EmptyString, log.LstdFlags),
		},
	}
	exp := 1
	if rcv, err := sl.Write([]byte{222}); err != nil {
		t.Error(err)
	} else if rcv != exp {
		t.Errorf("Expected <%v>, received <%v>", exp, rcv)
	}
}

func TestGetLogLevelStdLogger(t *testing.T) {
	sl := &StdLogger{}
	exp := 0
	if rcv := sl.GetLogLevel(); rcv != exp {
		t.Errorf("Expected <%v>, received %v %T", exp, rcv, rcv)
	}

}
