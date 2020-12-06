// +build integration

package utils

import (
	"bytes"
	"log/syslog"
	"net"
	"runtime"
	"strings"
	"testing"
)

var (
	testsSyslogLogger = []func(t *testing.T){
		testEmergencyLogger,
		testAlertLogger,
		testCriticalLogger,
		testErrorLogger,
		testWarningLogger,
		testNoticeLogger,
		testInfoLogger,
		testDebugLogger,
	}
)

type unixServer struct {
	l    net.Listener
	buf  *bytes.Buffer
	path string
}

func newUnixServer(path string) (*unixServer, error) {
	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	return &unixServer{
		l:    l,
		buf:  new(bytes.Buffer),
		path: path,
	}, nil
}

func (u *unixServer) Close() error {
	return u.l.Close()
}

func (u *unixServer) serveConn(c net.Conn) error {
	for {
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			return err
		}

		_, err = u.buf.Write(buf[0:nr])
		if err != nil {
			return err
		}
	}
}

func (u *unixServer) Serve() error {
	for {
		fd, err := u.l.Accept()
		if err != nil {
			return err
		}

		go u.serveConn(fd)
	}
}

func (u *unixServer) String() string {
	return u.buf.String()
}

func TestLoggerSyslog(t *testing.T) {
	for _, test := range testsSyslogLogger {
		t.Run("Syslog_logger", test)
	}
}

func testEmergencyLogger(t *testing.T) {
	flPath := "/tmp/testEmergency2"

	l, err := newUnixServer(flPath)
	if err != nil {
		t.Error(err)
	}
	go l.Serve()

	writer, err := syslog.Dial("unix", flPath, syslog.LOG_EMERG, "id_emergency")
	if err != nil {
		t.Error(err)
	}

	newLogger := new(StdLogger)
	newLogger.SetSyslog(writer)

	if err := newLogger.Emerg("emergency_panic"); err != nil {
		t.Error(err)
	}
	runtime.Gosched()
	if err := writer.Close(); err != nil {
		t.Error(err)
	}
	if err := l.Close(); err != nil {
		t.Error(err)
	}
	expected := "emergency_panic"
	if rcv := l.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %q, received %q", expected, rcv)
	}
}

func testAlertLogger(t *testing.T) {
	flPath := "/tmp/testAlert2"

	l, err := newUnixServer(flPath)
	if err != nil {
		t.Error(err)
	}
	go l.Serve()

	writer, err := syslog.Dial("unix", flPath, syslog.LOG_ALERT, "id_alert")
	if err != nil {
		t.Error(err)
	}

	newLogger := new(StdLogger)
	newLogger.SetSyslog(writer)

	newLogger.SetLogLevel(7)

	if err := newLogger.Alert("emergency_alert"); err != nil {
		t.Error(err)
	}
	runtime.Gosched()
	if err := writer.Close(); err != nil {
		t.Error(err)
	}
	if err := l.Close(); err != nil {
		t.Error(err)
	}
	expected := "emergency_alert"
	if rcv := l.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %q, received %q", expected, rcv)
	}
}

func testCriticalLogger(t *testing.T) {
	flPath := "/tmp/testCritical2"

	l, err := newUnixServer(flPath)
	if err != nil {
		t.Error(err)
	}
	go l.Serve()

	writer, err := syslog.Dial("unix", flPath, syslog.LOG_CRIT, "id_critical")
	if err != nil {
		t.Error(err)
	}
	newLogger := new(StdLogger)
	newLogger.SetSyslog(writer)
	newLogger.SetLogLevel(7)

	if err := newLogger.Crit("critical_panic"); err != nil {
		t.Error(err)
	}
	runtime.Gosched()
	if err := writer.Close(); err != nil {
		t.Error(err)
	}
	if err := l.Close(); err != nil {
		t.Error(err)
	}
	expected := "critical_panic"
	if rcv := l.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %q, received %q", expected, rcv)
	}
}

func testErrorLogger(t *testing.T) {
	flPath := "/tmp/testError2"

	l, err := newUnixServer(flPath)
	if err != nil {
		t.Error(err)
	}
	go l.Serve()

	writer, err := syslog.Dial("unix", flPath, syslog.LOG_ERR, "id_error")
	if err != nil {
		t.Error(err)
	}
	newLogger := new(StdLogger)
	newLogger.SetSyslog(writer)
	newLogger.SetLogLevel(7)

	if err := newLogger.Err("error_panic"); err != nil {
		t.Error(err)
	}
	runtime.Gosched()
	if err := writer.Close(); err != nil {
		t.Error(err)
	}
	if err := l.Close(); err != nil {
		t.Error(err)
	}
	expected := "error_panic"
	if rcv := l.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %q, received %q", expected, rcv)
	}
}

func testWarningLogger(t *testing.T) {
	flPath := "/tmp/testWarning2"

	l, err := newUnixServer(flPath)
	if err != nil {
		t.Error(err)
	}
	go l.Serve()

	writer, err := syslog.Dial("unix", flPath, syslog.LOG_WARNING, "id_warning")
	if err != nil {
		t.Error(err)
	}
	newLogger := new(StdLogger)
	newLogger.SetSyslog(writer)
	newLogger.SetLogLevel(7)

	if err := newLogger.Warning("warning_panic"); err != nil {
		t.Error(err)
	}
	runtime.Gosched()
	if err := writer.Close(); err != nil {
		t.Error(err)
	}
	if err := l.Close(); err != nil {
		t.Error(err)
	}
	expected := ""
	if rcv := l.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %q, received %q", expected, rcv)
	}
}

func testNoticeLogger(t *testing.T) {
	flPath := "/tmp/testNotice2"

	l, err := newUnixServer(flPath)
	if err != nil {
		t.Error(err)
	}
	go l.Serve()

	writer, err := syslog.Dial("unix", flPath, syslog.LOG_NOTICE, "id_notice")
	if err != nil {
		t.Error(err)
	}
	newLogger := new(StdLogger)
	newLogger.SetSyslog(writer)
	newLogger.SetLogLevel(7)

	if err := newLogger.Notice("notice_panic"); err != nil {
		t.Error(err)
	}
	runtime.Gosched()

	if err := writer.Close(); err != nil {
		t.Error(err)
	}
	if err := l.Close(); err != nil {
		t.Error(err)
	}
	expected := "notice_panic"
	if rcv := l.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %q, received %q", expected, rcv)
	}
}

func testInfoLogger(t *testing.T) {
	flPath := "/tmp/testInfo2"

	l, err := newUnixServer(flPath)
	if err != nil {
		t.Error(err)
	}
	go l.Serve()

	writer, err := syslog.Dial("unix", flPath, syslog.LOG_NOTICE, "id_info")
	if err != nil {
		t.Error(err)
	}
	newLogger := new(StdLogger)
	newLogger.SetSyslog(writer)
	newLogger.SetLogLevel(7)

	if err := newLogger.Info("info_panic"); err != nil {
		t.Error(err)
	}
	runtime.Gosched()

	if err := writer.Close(); err != nil {
		t.Error(err)
	}
	if err := l.Close(); err != nil {
		t.Error(err)
	}
	expected := "info_panic"
	if rcv := l.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %q, received %q", expected, rcv)
	}
}

func testDebugLogger(t *testing.T) {
	flPath := "/tmp/testDebug2"

	l, err := newUnixServer(flPath)
	if err != nil {
		t.Error(err)
	}
	go l.Serve()

	writer, err := syslog.Dial("unix", flPath, syslog.LOG_NOTICE, "id_debug")
	if err != nil {
		t.Error(err)
	}
	newLogger := new(StdLogger)
	newLogger.SetSyslog(writer)
	newLogger.SetLogLevel(7)

	if err := newLogger.Debug("debug_panic"); err != nil {
		t.Error(err)
	}
	runtime.Gosched()

	if err := writer.Close(); err != nil {
		t.Error(err)
	}
	if err := l.Close(); err != nil {
		t.Error(err)
	}
	expected := "debug_panic"
	if rcv := l.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %q, received %q", expected, rcv)
	}
}
