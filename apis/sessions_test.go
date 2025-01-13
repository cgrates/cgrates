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

package apis

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestAuthorizeEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}
	var reply sessions.V1AuthorizeReply
	args := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	if err := ssv1.AuthorizeEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
}

func TestAuthorizeEventWithDigest(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}
	var reply sessions.V1AuthorizeReplyWithDigest
	args := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	if err := ssv1.AuthorizeEventWithDigest(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
}

// func TestInitiateSession(t *testing.T) {
// 	cfg := config.NewDefaultCGRConfig()
// 	connMgr := engine.NewConnManager(cfg)
// 	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
// 	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
// 	ssv1 := &SessionSv1{
// 		ping: struct{}{},
// 		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
// 	}
// 	var reply sessions.V1InitSessionReply
// 	args := &utils.CGREvent{
// 		ID:     "TestMatchingAccountsForEvent",
// 		Tenant: "cgrates.org",
// 		Event: map[string]any{
// 			utils.AccountField: "1001",
// 		},
// 	}
// 	if err := ssv1.InitiateSession(context.Background(), args, &reply); err != nil {
// 		t.Error(err)
// 	}
// }

func TestInitiateSessionWithDigest(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}
	var reply sessions.V1InitReplyWithDigest
	args := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	if err := ssv1.InitiateSessionWithDigest(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
}

func TestUpdateSession(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}
	var reply sessions.V1UpdateSessionReply
	args := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	if err := ssv1.UpdateSession(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
}

func TestSyncSessions(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}
	var reply string
	args := &utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
	}
	if err := ssv1.SyncSessions(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected%v\n but received %v", utils.OK, reply)
	}
}

func TestTerminateSessions(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}
	var reply string
	args := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	if err := ssv1.TerminateSession(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
}

func TestProcessCDR(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}

	var reply string
	args := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
		},
	}
	errExp := "UNSUPPORTED_SERVICE_METHOD"
	if err := ssv1.ProcessCDR(context.Background(), args, &reply); err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}

}

func TestProcessMessage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}

	var reply sessions.V1ProcessMessageReply
	args := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	if err := ssv1.ProcessMessage(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
}

func TestProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}

	var reply sessions.V1ProcessEventReply
	args := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	if err := ssv1.ProcessEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
}

func TestGetActiveSessions(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}

	var reply []*sessions.ExternalSession
	args := &utils.SessionFilter{
		Limit:   utils.IntPointer(2),
		Filters: []string{},
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
	}
	if err := ssv1.GetActiveSessions(context.Background(), args, &reply); err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestGetActiveSessionsCount(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}

	var reply int
	args := &utils.SessionFilter{
		Limit:   utils.IntPointer(2),
		Filters: []string{},
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
	}
	if err := ssv1.GetActiveSessionsCount(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != 0 {
		t.Errorf("Expected 0")
	}
}

func TestForceDisconnect(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}

	var reply string
	args := &utils.SessionFilter{
		Limit:   utils.IntPointer(2),
		Filters: []string{},
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
	}
	if err := ssv1.ForceDisconnect(context.Background(), args, &reply); err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestGetPassiveSessions(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}

	var reply []*sessions.ExternalSession
	args := &utils.SessionFilter{
		Limit:   utils.IntPointer(2),
		Filters: []string{},
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
	}
	if err := ssv1.GetPassiveSessions(context.Background(), args, &reply); err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestGetPassiveSessionsCount(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}

	var reply int
	args := &utils.SessionFilter{
		Limit:   utils.IntPointer(2),
		Filters: []string{},
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
	}
	if err := ssv1.GetPassiveSessionsCount(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != 0 {
		t.Errorf("Expected 0")
	}
}

// func TestSetPassiveSession(t *testing.T) {
// 	cfg := config.NewDefaultCGRConfig()
// 	connMgr := engine.NewConnManager(cfg)
// 	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
// 	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
// 	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
// 	ssv1 := &SessionSv1{
// 		ping: struct{}{},
// 		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
// 	}

// 	var reply string
// 	args := &sessions.Session{
// 		ID: "1001",
// 	}
// 	if err := ssv1.SetPassiveSession(context.Background(), args, &reply); err != utils.ErrNotFound {
// 		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
// 	}
// }

func TestActivateSessions(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}

	var reply string
	args := &utils.SessionIDsWithAPIOpts{
		IDs: []string{"TestMatchingAccountsForEvent"},
	}
	if err := ssv1.ActivateSessions(context.Background(), args, &reply); err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %v\n but received %v", utils.ErrPartiallyExecuted, err)
	}
}

func TestDeactivateSessions(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}

	var reply string
	args := &utils.SessionIDsWithAPIOpts{
		IDs: []string{"TestMatchingAccountsForEvent"},
	}
	if err := ssv1.DeactivateSessions(context.Background(), args, &reply); err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %v\n but received %v", utils.ErrPartiallyExecuted, err)
	}
}

func TestReAuthorize(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}

	var reply string
	args := &utils.SessionFilter{
		Limit:   utils.IntPointer(2),
		Filters: []string{},
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
	}
	if err := ssv1.ReAuthorize(context.Background(), args, &reply); err != utils.ErrNotFound {
		t.Errorf("Expected %v\n but received %v", utils.ErrNotFound, err)
	}
}

func TestDisconnectPeer(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}

	var reply string

	args := &utils.DPRArgs{
		OriginHost:      "origin_host",
		OriginRealm:     "origin_realm",
		DisconnectCause: 2,
	}

	if err := ssv1.DisconnectPeer(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
}

func TestSTIRAuthenticate(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}

	var reply string

	args := &sessions.V1STIRAuthenticateArgs{
		DestinationTn: "dest_tn",
		Identity:      "identity",
	}
	errExpect := "*stir_authenticate: missing parts of the message header"
	if err := ssv1.STIRAuthenticate(context.Background(), args, &reply); err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}

func TestSTIRIdentity(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}

	var reply string

	args := &sessions.V1STIRIdentityArgs{
		PublicKeyPath:  "PKP",
		PrivateKeyPath: "PKP_PRIVATE",
		Payload: &utils.PASSporTPayload{
			ATTest: "at_Test",
		},
	}
	errExpect := "*stir_authenticate: open PKP_PRIVATE: no such file or directory"
	if err := ssv1.STIRIdentity(context.Background(), args, &reply); err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}

func TestRegisterInternalBiJSONConn(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.SessionSCfg().CDRsConns = []string{"*internal"}
	ssv1 := &SessionSv1{
		ping: struct{}{},
		sS:   sessions.NewSessionS(cfg, dm, engine.NewFilterS(cfg, connMgr, dm), connMgr),
	}

	var reply string

	args := "*internal"

	if err := ssv1.RegisterInternalBiJSONConn(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
}
