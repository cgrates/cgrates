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

package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/segmentio/kafka-go"
)

func TestLoggerNewExportLogger(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.LoggerCfg().Level = 7
	cfg.GeneralCfg().NodeID = "123"
	cM := NewConnManager(cfg)
	exp := &ExportLogger{
		ctx:        context.Background(),
		efsConns:   []string{"*internal:*efs"},
		connMgr:    cM,
		FldPostDir: "/var/spool/cgrates/failed_posts",
		LogLevel:   7,
		NodeID:     "123",
		Tenant:     "cgrates.org",
		Writer: &kafka.Writer{
			Addr:        kafka.TCP(cfg.LoggerCfg().Opts.KafkaConn),
			Topic:       cfg.LoggerCfg().Opts.KafkaTopic,
			MaxAttempts: cfg.LoggerCfg().Opts.KafkaAttempts,
		},
	}
	if rcv := NewExportLogger(context.Background(), "cgrates.org", cM, cfg); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestCloseExportLogger(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.LoggerCfg().Level = 7
	cfg.GeneralCfg().NodeID = "123"
	cM := NewConnManager(cfg)
	el := NewExportLogger(context.Background(), "cgrates.org", cM, cfg)

	if el == nil {
		t.Error("Export logger shouldn't be empty")
	}

	if err := el.Close(); err != nil {
		t.Error(err)
	}
	exp := &ExportLogger{
		ctx:        context.Background(),
		connMgr:    cM,
		efsConns:   []string{"*internal:*efs"},
		LogLevel:   7,
		FldPostDir: cfg.LoggerCfg().Opts.FailedPostsDir,
		NodeID:     "123",
		Tenant:     "cgrates.org",
		Writer:     nil,
	}
	if !reflect.DeepEqual(exp, el) {
		t.Errorf("\nExpected \n<%+v>, \nReceived \n<%+v>", exp, el)
	}

}

func TestExportLoggerCallErrWriter(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)

	connID := "connID"
	cfg := config.NewDefaultCGRConfig()
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			ID:      connID,
			Address: utils.MetaInternal,
		},
	}
	efsConn := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs)

	cfg.LoggerCfg().EFsConns = []string{efsConn}

	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EfSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				return nil
			},
		},
	}
	cM := NewConnManager(cfg)
	cM.connCache.Set(connID, nil, nil)
	cM.AddInternalConn(efsConn, utils.EfSv1, rpcInternal)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), cM)
	Cache = NewCacheS(cfg, dm, cM, nil)

	el := NewExportLogger(context.Background(), "cgrates.org", cM, cfg)

	if err := el.call("test msg", 7); err != utils.ErrLoggerChanged || err == nil {
		t.Error(err)
	}

}

func TestLoggerExportEmergNil(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cM := NewConnManager(cfg)

	el := NewExportLogger(context.Background(), "cgrates.org", cM, cfg)

	if err := el.Emerg("Emergency message"); err != nil {
		t.Error(err)
	}

}

// func TestLoggerExportDebugOk(t *testing.T) {

// 	tmpC := Cache
// 	defer func() {
// 		Cache = tmpC

// 	}()
// 	Cache.Clear(nil)

// 	connID := "connID"
// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
// 	cfg.RPCConns()[connID].Conns = []*config.RemoteHost{
// 		{
// 			ID:      connID,
// 			Address: utils.MetaInternal,
// 		},
// 	}
// 	efsConn := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs)

// 	cfg.LoggerCfg().EFsConns = []string{efsConn}

// 	rpcInternal := make(chan birpc.ClientConnector, 1)
// 	rpcInternal <- &ccMock{
// 		calls: map[string]func(ctx *context.Context, args any, reply any) error{
// 			utils.EfSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
// 				return nil
// 			},
// 		},
// 	}
// 	cM := NewConnManager(cfg)
// 	cM.connCache.Set(connID, nil, nil)
// 	cM.AddInternalConn(efsConn, utils.EfSv1, rpcInternal)
// 	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
// 	dm := NewDataManager(db, cfg.CacheCfg(), cM)
// 	Cache = NewCacheS(cfg, dm, cM, nil)

// 	el := NewExportLogger(context.Background(), "123", "cgrates.org", 7, cM, cfg)

// 	logMsg := "log message"

// 	if err := el.Debug(logMsg); err != nil {
// 		t.Error(err)
// 	}

// 	el.SetLogLevel(utils.LOGLEVEL_DEBUG)

// }

/*
	func TestLoggerExportAlert(t *testing.T) {
		testCache := Cache
		tmpC := config.CgrConfig()
		tmpCM := connMgr
		defer func() {
			Cache = testCache
			config.SetCgrConfig(tmpC)
			connMgr = tmpCM
		}()

		cfg := config.NewDefaultCGRConfig()
		cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
			utils.MetaEEs)}

		data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
		cM := NewConnManager(cfg)
		dm := NewDataManager(data, cfg.CacheCfg(), nil)
		fltrs := NewFilterS(cfg, nil, dm)
		Cache = NewCacheS(cfg, dm, nil, nil)
		newCDRSrv := NewCDRServer(cfg, dm, fltrs, cM)
		ccM := &ccMock{
			calls: map[string]func(ctx *context.Context, args any, reply any) error{
				utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
					*reply.(*map[string]map[string]any) = map[string]map[string]any{}
					return utils.ErrNotFound
				},
			},
		}
		rpcInternal := make(chan birpc.ClientConnector, 1)
		rpcInternal <- ccM
		newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
			utils.MetaEEs), utils.EeSv1, rpcInternal)

		el := NewExportLogger(context.Background(), "123", "cgrates.org", 0, cM, cfg)

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
			calls: map[string]func(ctx *context.Context, args any, reply any) error{
				utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
					delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
					exp := &utils.CGREventWithEeIDs{
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							Event: map[string]any{
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
*/
// func TestLoggerExportErr(t *testing.T) {
// 	tmp := Cache
// 	defer func() {
// 		Cache = tmp
// 	}()
// 	efsConn := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs)
// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.LoggerCfg().EFsConns = []string{efsConn}
// 	Cache = NewCacheS(cfg, nil, nil, nil)
// 	cM := NewConnManager(cfg)
// 	ccM := &ccMock{
// 		calls: map[string]func(ctx *context.Context, args any, reply any) error{
// 			utils.EfSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
// 				delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
// 				exp := &utils.CGREventWithEeIDs{
// 					CGREvent: &utils.CGREvent{
// 						Tenant: "cgrates.org",
// 						Event: map[string]any{
// 							utils.NodeID: "123",
// 							"Message":    "Error message",
// 							"Severity":   utils.LOGLEVEL_ERROR,
// 						},
// 					},
// 				}
// 				if !reflect.DeepEqual(exp, args) {
// 					return fmt.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
// 						utils.ToJSON(exp), utils.ToJSON(args))
// 				}
// 				return nil
// 			},
// 		},
// 	}
// 	rpcInternal := make(chan birpc.ClientConnector, 1)
// 	rpcInternal <- ccM
// 	cM.AddInternalConn(efsConn, utils.EfSv1, rpcInternal)

// 	el := NewExportLogger(context.Background(), "123", "cgrates.org", 2, cM, cfg)

// 	if err := el.Err("Error message"); err != nil {
// 		t.Error(err)
// 	}
// 	el.SetLogLevel(3)
// 	if err := el.Err("Error message"); err != nil {
// 		t.Error(err)
// 	}
// }

/*
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
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
				exp := &utils.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						Event: map[string]any{
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
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
				exp := &utils.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						Event: map[string]any{
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
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
				exp := &utils.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						Event: map[string]any{
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
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
				exp := &utils.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						Event: map[string]any{
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
*/

func TestLoggerSetGetLogLevel(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := NewConnManager(cfg)
	el := NewExportLogger(context.Background(), "cgrates.org", cM, cfg)
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
	cM := NewConnManager(cfg)
	el := NewExportLogger(context.Background(), "cgrates.org", cM, cfg)
	if el.GetSyslog() != nil {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", nil, el.GetSyslog())
	}
}

/*
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
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				delete(args.(*utils.CGREventWithEeIDs).Event, "Timestamp")
				exp := &utils.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						Event: map[string]any{
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
}*/
