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

/*
func TestLoggerNewLoggerExport(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	exp := &ExportLogger{
		logLevel: 6,
		nodeID:   "123",
		tenant:   "cgrates.org",
		loggOpts: cfg.LoggerCfg().Opts,
		writer: &kafka.Writer{
			Addr:        kafka.TCP(cfg.LoggerCfg().Opts.KafkaConn),
			Topic:       cfg.LoggerCfg().Opts.KafkaTopic,
			MaxAttempts: cfg.LoggerCfg().Opts.Attempts,
		},
	}
	if rcv, err := NewLogger(utils.MetaKafka, "cgrates.org", "123", cfg.LoggerCfg()); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv.(*ExportLogger), exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestLoggerNewLoggerDefault(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	experr := `unsupported logger: <invalid>`
	if _, err := NewLogger("invalid", "cgrates.org", "123", cfg.LoggerCfg()); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%s>, \nreceived: <%s>", experr, err)
	}
}

func TestLoggerNewExportLogger(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	exp := &ExportLogger{
		logLevel: 7,
		nodeID:   "123",
		tenant:   "cgrates.org",
		loggOpts: cfg.LoggerCfg().Opts,
		writer: &kafka.Writer{
			Addr:        kafka.TCP(cfg.LoggerCfg().Opts.KafkaConn),
			Topic:       cfg.LoggerCfg().Opts.KafkaTopic,
			MaxAttempts: cfg.LoggerCfg().Opts.Attempts,
		},
	}
	if rcv := NewExportLogger("123", "cgrates.org", 7, cfg.LoggerCfg().Opts); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestLoggerExportEmerg(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	eesConn := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
	cfg := config.NewDefaultCGRConfig()
	cfg.CoreSCfg().EEsConns = []string{eesConn}
	Cache = NewCacheS(cfg, nil, nil)
	cM := NewConnManager(cfg)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
				exp := &utils.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						Event: map[string]interface{}{
							utils.NodeID: "123",
							"Message":    "Emergency message",
							"Severity":   utils.LOGLEVEL_EMERGENCY,
						},
					},
				}
				if !reflect.DeepEqual(exp, args) {
					return fmt.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(exp), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM.AddInternalConn(eesConn, utils.EeSv1, rpcInternal)

	el := NewExportLogger("123", "cgrates.org", -1, cfg.LoggerCfg().Opts)

	if err := el.Emerg("Emergency message"); err != nil {
		t.Error(err)
	}
	el.SetLogLevel(0)
	if err := el.Emerg("Emergency message"); err != nil {
		t.Error(err)
	}
}

func TestLoggerExportAlert(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	eesConn := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
	cfg := config.NewDefaultCGRConfig()
	cfg.CoreSCfg().EEsConns = []string{eesConn}
	Cache = NewCacheS(cfg, nil, nil)
	cM := NewConnManager(cfg)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
				exp := &utils.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						Event: map[string]interface{}{
							utils.NodeID: "123",
							"Message":    "Alert message",
							"Severity":   utils.LOGLEVEL_ALERT,
						},
					},
				}
				if !reflect.DeepEqual(exp, args) {
					return fmt.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(exp), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM.AddInternalConn(eesConn, utils.EeSv1, rpcInternal)

	el := NewExportLogger("123", "cgrates.org", 0, cfg.LoggerCfg().Opts)

	if err := el.Alert("Alert message"); err != nil {
		t.Error(err)
	}
	el.SetLogLevel(1)
	if err := el.Alert("Alert message"); err != nil {
		t.Error(err)
	}
}

func TestLoggerExportCrit(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	eesConn := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
	cfg := config.NewDefaultCGRConfig()
	cfg.CoreSCfg().EEsConns = []string{eesConn}
	Cache = NewCacheS(cfg, nil, nil)
	cM := NewConnManager(cfg)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
				exp := &utils.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						Event: map[string]interface{}{
							utils.NodeID: "123",
							"Message":    "Critical message",
							"Severity":   utils.LOGLEVEL_CRITICAL,
						},
					},
				}
				if !reflect.DeepEqual(exp, args) {
					return fmt.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(exp), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM.AddInternalConn(eesConn, utils.EeSv1, rpcInternal)

	el := NewExportLogger("123", "cgrates.org", 1, cfg.LoggerCfg().Opts)

	if err := el.Crit("Critical message"); err != nil {
		t.Error(err)
	}
	el.SetLogLevel(2)
	if err := el.Crit("Critical message"); err != nil {
		t.Error(err)
	}
}

func TestLoggerExportErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	eesConn := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
	cfg := config.NewDefaultCGRConfig()
	cfg.CoreSCfg().EEsConns = []string{eesConn}
	Cache = NewCacheS(cfg, nil, nil)
	cM := NewConnManager(cfg)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
				exp := &utils.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						Event: map[string]interface{}{
							utils.NodeID: "123",
							"Message":    "Error message",
							"Severity":   utils.LOGLEVEL_ERROR,
						},
					},
				}
				if !reflect.DeepEqual(exp, args) {
					return fmt.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(exp), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM.AddInternalConn(eesConn, utils.EeSv1, rpcInternal)

	el := NewExportLogger("123", "cgrates.org", 2, cfg.LoggerCfg().Opts)

	if err := el.Err("Error message"); err != nil {
		t.Error(err)
	}
	el.SetLogLevel(3)
	if err := el.Err("Error message"); err != nil {
		t.Error(err)
	}
}

func TestLoggerExportWarning(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	eesConn := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
	cfg := config.NewDefaultCGRConfig()
	cfg.CoreSCfg().EEsConns = []string{eesConn}
	Cache = NewCacheS(cfg, nil, nil)
	cM := NewConnManager(cfg)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
				exp := &utils.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						Event: map[string]interface{}{
							utils.NodeID: "123",
							"Message":    "Warning message",
							"Severity":   utils.LOGLEVEL_WARNING,
						},
					},
				}
				if !reflect.DeepEqual(exp, args) {
					return fmt.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(exp), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM.AddInternalConn(eesConn, utils.EeSv1, rpcInternal)

	el := NewExportLogger("123", "cgrates.org", 3, cfg.LoggerCfg().Opts)

	if err := el.Warning("Warning message"); err != nil {
		t.Error(err)
	}
	el.SetLogLevel(4)
	if err := el.Warning("Warning message"); err != nil {
		t.Error(err)
	}
}

func TestLoggerExportNotice(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	eesConn := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
	cfg := config.NewDefaultCGRConfig()
	cfg.CoreSCfg().EEsConns = []string{eesConn}
	Cache = NewCacheS(cfg, nil, nil)
	cM := NewConnManager(cfg)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
				exp := &utils.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						Event: map[string]interface{}{
							utils.NodeID: "123",
							"Message":    "Notice message",
							"Severity":   utils.LOGLEVEL_NOTICE,
						},
					},
				}
				if !reflect.DeepEqual(exp, args) {
					return fmt.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(exp), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM.AddInternalConn(eesConn, utils.EeSv1, rpcInternal)

	el := NewExportLogger("123", "cgrates.org", 4, cfg.LoggerCfg().Opts)

	if err := el.Notice("Notice message"); err != nil {
		t.Error(err)
	}
	el.SetLogLevel(5)
	if err := el.Notice("Notice message"); err != nil {
		t.Error(err)
	}
}

func TestLoggerExportInfo(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	eesConn := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
	cfg := config.NewDefaultCGRConfig()
	cfg.CoreSCfg().EEsConns = []string{eesConn}
	Cache = NewCacheS(cfg, nil, nil)
	cM := NewConnManager(cfg)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
				exp := &utils.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						Event: map[string]interface{}{
							utils.NodeID: "123",
							"Message":    "Info message",
							"Severity":   utils.LOGLEVEL_INFO,
						},
					},
				}
				if !reflect.DeepEqual(exp, args) {
					return fmt.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(exp), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM.AddInternalConn(eesConn, utils.EeSv1, rpcInternal)

	el := NewExportLogger("123", "cgrates.org", 5, cfg.LoggerCfg().Opts)

	if err := el.Info("Info message"); err != nil {
		t.Error(err)
	}
	el.SetLogLevel(6)
	if err := el.Info("Info message"); err != nil {
		t.Error(err)
	}
}

func TestLoggerExportDebug(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	eesConn := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
	cfg := config.NewDefaultCGRConfig()
	cfg.CoreSCfg().EEsConns = []string{eesConn}
	Cache = NewCacheS(cfg, nil, nil)
	cM := NewConnManager(cfg)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
				exp := &utils.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						Event: map[string]interface{}{
							utils.NodeID: "123",
							"Message":    "Debug message",
							"Severity":   utils.LOGLEVEL_DEBUG,
						},
					},
				}
				if !reflect.DeepEqual(exp, args) {
					return fmt.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(exp), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM.AddInternalConn(eesConn, utils.EeSv1, rpcInternal)

	el := NewExportLogger("123", "cgrates.org", 6, cfg.LoggerCfg().Opts)

	if err := el.Debug("Debug message"); err != nil {
		t.Error(err)
	}
	el.SetLogLevel(7)
	if err := el.Debug("Debug message"); err != nil {
		t.Error(err)
	}
}

func TestLoggerSetGetLogLevel(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	el := NewExportLogger("123", "cgrates.org", 6, cfg.LoggerCfg().Opts)
	if rcv := el.GetLogLevel(); rcv != 6 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 6, rcv)
	}
	el.SetLogLevel(3)
	if rcv := el.GetLogLevel(); rcv != 3 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 3, rcv)
	}
}

func TestLoggerGetSyslog(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	el := NewExportLogger("123", "cgrates.org", 6, cfg.LoggerCfg().Opts)
	if el.GetSyslog() != nil {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", nil, el.GetSyslog())
	}
}

func TestLoggerExportWrite(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()
	eesConn := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
	cfg := config.NewDefaultCGRConfig()
	cfg.CoreSCfg().EEsConns = []string{eesConn}
	Cache = NewCacheS(cfg, nil, nil)
	cM := NewConnManager(cfg)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
				exp := &utils.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						Event: map[string]interface{}{
							utils.NodeID: "123",
							"Message":    "message",
							"Severity":   8,
						},
					},
				}
				if !reflect.DeepEqual(exp, args) {
					return fmt.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(exp), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM.AddInternalConn(eesConn, utils.EeSv1, rpcInternal)

	el := NewExportLogger("123", "cgrates.org", 8, cfg.LoggerCfg().Opts)

	if _, err := el.Write([]byte("message")); err != nil {
		t.Error(err)
	}
	el.Close()
} */
