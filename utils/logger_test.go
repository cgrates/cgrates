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
	"log/syslog"
	"reflect"
	"testing"
)

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

func TestLoggerAlert(t *testing.T) {
	sl := &SysLogger{
		logLevel: 0,
		syslog:   &syslog.Writer{},
	}
	err := sl.Alert("test")

	if err != nil {
		t.Error(err)
	}

	sl2 := &SysLogger{
		logLevel: 100,
		syslog:   &syslog.Writer{},
	}

	err = sl2.Alert("test")

	if err != nil {
		t.Error(err)
	}
}

func TestLoggerCrit(t *testing.T) {
	sl := &SysLogger{
		logLevel: 0,
		syslog:   &syslog.Writer{},
	}
	err := sl.Crit("test")

	if err != nil {
		t.Error(err)
	}

	sl2 := &SysLogger{
		logLevel: 100,
		syslog:   &syslog.Writer{},
	}

	err = sl2.Crit("test")

	if err != nil {
		t.Error(err)
	}
}

func TestLoggerDebug(t *testing.T) {
	sl := &SysLogger{
		logLevel: 0,
		syslog:   &syslog.Writer{},
	}
	err := sl.Debug("test")

	if err != nil {
		t.Error(err)
	}

	sl2 := &SysLogger{
		logLevel: 100,
		syslog:   &syslog.Writer{},
	}

	err = sl2.Debug("test")

	if err != nil {
		t.Error(err)
	}
}

//Loggs: Broadcast message from systemd-journald@debian (Tue 2023-09-12 12:28:00 CEST):
/*func TestLoggerEmerg(t *testing.T) {
	sl := &SysLogger{
		logLevel: 0,
		syslog:   &syslog.Writer{},
	}
	err := sl.Emerg("test")

	if err != nil {
		t.Error(err)
	}

	sl2 := &SysLogger{
		logLevel: 100,
		syslog:   &syslog.Writer{},
	}

	err = sl2.Emerg("test")

	if err != nil {
		t.Error(err)
	}
}*/

func TestLoggerErr(t *testing.T) {
	sl := &SysLogger{
		logLevel: 0,
		syslog:   &syslog.Writer{},
	}
	err := sl.Err("test")

	if err != nil {
		t.Error(err)
	}

	sl2 := &SysLogger{
		logLevel: 100,
		syslog:   &syslog.Writer{},
	}

	err = sl2.Err("test")

	if err != nil {
		t.Error(err)
	}
}

func TestLoggerInfo(t *testing.T) {
	sl := &SysLogger{
		logLevel: 0,
		syslog:   &syslog.Writer{},
	}
	err := sl.Info("test")

	if err != nil {
		t.Error(err)
	}

	sl2 := &SysLogger{
		logLevel: 100,
		syslog:   &syslog.Writer{},
	}

	err = sl2.Info("test")

	if err != nil {
		t.Error(err)
	}
}

func TestLoggerNotice(t *testing.T) {
	sl := &SysLogger{
		logLevel: 0,
		syslog:   &syslog.Writer{},
	}
	err := sl.Notice("test")

	if err != nil {
		t.Error(err)
	}

	sl2 := &SysLogger{
		logLevel: 100,
		syslog:   &syslog.Writer{},
	}

	err = sl2.Notice("test")

	if err != nil {
		t.Error(err)
	}
}

func TestLoggerWarning(t *testing.T) {
	sl := &SysLogger{
		logLevel: -1,
		syslog:   &syslog.Writer{},
	}
	err := sl.Warning("test")

	if err != nil {
		t.Error(err)
	}

	sl2 := &SysLogger{
		logLevel: 100,
		syslog:   &syslog.Writer{},
	}

	err = sl2.Warning("test")

	if err != nil {
		t.Error(err)
	}
}
