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

package sessions

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

const (
	MaxTimespansInCost = 50
	MaxSessionTTL      = 10000 // maximum session TTL in miliseconds
)

var (
	ErrPartiallyExecuted = errors.New("Partially executed")
	ErrActiveSession     = errors.New("ACTIVE_SESSION")
	debug                bool
)

func NewSessionReplicationConns(conns []*config.HaPoolConfig, reconnects int,
	connTimeout, replyTimeout time.Duration) (smgConns []*SMGReplicationConn, err error) {
	smgConns = make([]*SMGReplicationConn, len(conns))
	for i, replConnCfg := range conns {
		if replCon, err := rpcclient.NewRpcClient("tcp", replConnCfg.Address, replConnCfg.Tls, "", "", "", 0, reconnects,
			connTimeout, replyTimeout, replConnCfg.Transport[1:], nil, true); err != nil {
			return nil, err
		} else {
			smgConns[i] = &SMGReplicationConn{Connection: replCon, Synchronous: replConnCfg.Synchronous}
		}
	}
	return
}

// ReplicationConnection represents one connection to a passive node where we will replicate session data
type SMGReplicationConn struct {
	Connection  rpcclient.RpcClientConnection
	Synchronous bool
}

func NewSMGeneric(cgrCfg *config.CGRConfig, rals, resS, thdS,
	statS, splS, attrS, cdrsrv, chargerS rpcclient.RpcClientConnection,
	smgReplConns []*SMGReplicationConn, timezone string) *SMGeneric {
	ssIdxCfg := cgrCfg.SessionSCfg().SessionIndexes
	ssIdxCfg[utils.OriginID] = true // Make sure we have indexing for OriginID since it is a requirement on prefix searching
	if rals != nil && reflect.ValueOf(rals).IsNil() {
		rals = nil
	}
	if resS != nil && reflect.ValueOf(resS).IsNil() {
		resS = nil
	}
	if thdS != nil && reflect.ValueOf(thdS).IsNil() {
		thdS = nil
	}
	if statS != nil && reflect.ValueOf(statS).IsNil() {
		statS = nil
	}
	if splS != nil && reflect.ValueOf(splS).IsNil() {
		splS = nil
	}
	if attrS != nil && reflect.ValueOf(attrS).IsNil() {
		attrS = nil
	}
	if cdrsrv != nil && reflect.ValueOf(cdrsrv).IsNil() {
		cdrsrv = nil
	}
	if chargerS != nil && reflect.ValueOf(chargerS).IsNil() {
		chargerS = nil
	}
	return &SMGeneric{
		cgrCfg:             cgrCfg,
		chargerS:           chargerS,
		rals:               rals,
		resS:               resS,
		thdS:               thdS,
		statS:              statS,
		splS:               splS,
		attrS:              attrS,
		cdrsrv:             cdrsrv,
		smgReplConns:       smgReplConns,
		Timezone:           timezone,
		biJsonConns:        make(map[*rpc2.Client]struct{}),
		activeSessions:     make(map[string][]*SMGSession),
		ssIdxCfg:           ssIdxCfg,
		aSessionsIndex:     make(map[string]map[string]map[string]utils.StringMap),
		aSessionsRIndex:    make(map[string][]*riFieldNameVal),
		passiveSessions:    make(map[string][]*SMGSession),
		pSessionsIndex:     make(map[string]map[string]map[string]utils.StringMap),
		pSessionsRIndex:    make(map[string][]*riFieldNameVal),
		sessionTerminators: make(map[string]*smgSessionTerminator),
		responseCache:      utils.NewResponseCache(cgrCfg.GeneralCfg().ResponseCacheTTL)}
}

type SMGeneric struct {
	cgrCfg             *config.CGRConfig // Separate from smCfg since there can be multiple
	chargerS           rpcclient.RpcClientConnection
	rals               rpcclient.RpcClientConnection // RALs connections
	resS               rpcclient.RpcClientConnection // ResourceS connections
	thdS               rpcclient.RpcClientConnection // ThresholdS connections
	statS              rpcclient.RpcClientConnection // StatS connections
	splS               rpcclient.RpcClientConnection // SupplierS connections
	attrS              rpcclient.RpcClientConnection // AttributeS connections
	cdrsrv             rpcclient.RpcClientConnection // CDR server connections
	smgReplConns       []*SMGReplicationConn         // list of connections where we will replicate our session data
	Timezone           string
	intBiJSONConns     []rpcclient.RpcClientConnection
	biJsonConns        map[*rpc2.Client]struct{} // index BiJSONConnection so we can sync them later
	activeSessions     map[string][]*SMGSession  // group sessions per sessionId, multiple runs based on derived charging
	aSessionsMux       sync.RWMutex
	ssIdxCfg           utils.StringMap                                  // index configuration
	aSessionsIndex     map[string]map[string]map[string]utils.StringMap // map[fieldName]map[fieldValue][runID]utils.StringMap[cgrID]
	aSessionsRIndex    map[string][]*riFieldNameVal                     // reverse indexes for active sessions, used on remove
	aSIMux             sync.RWMutex                                     // protects aSessionsIndex
	passiveSessions    map[string][]*SMGSession                         // group passive sessions
	pSessionsMux       sync.RWMutex
	pSessionsIndex     map[string]map[string]map[string]utils.StringMap // map[fieldName]map[fieldValue][runID]utils.StringMap[cgrID]
	pSessionsRIndex    map[string][]*riFieldNameVal                     // reverse indexes for active sessions, used on remove
	pSIMux             sync.RWMutex                                     // protects pSessionsIndex
	sessionTerminators map[string]*smgSessionTerminator                 // terminate and cleanup the session if timer expires
	sTsMux             sync.RWMutex                                     // protects sessionTerminators
	responseCache      *utils.ResponseCache                             // cache replies here
}

// riFieldNameVal is a reverse index entry
type riFieldNameVal struct {
	runID, fieldName, fieldValue string
}

type smgSessionTerminator struct {
	timer       *time.Timer
	endChan     chan bool
	ttl         time.Duration
	ttlLastUsed *time.Duration
	ttlUsage    *time.Duration
}

// setSessionTerminator installs a new terminator for a session
func (smg *SMGeneric) setSessionTerminator(s *SMGSession) {
	ttl, err := s.EventStart.GetDuration(utils.SessionTTL)
	switch err {
	case nil: // all good
	case utils.ErrNotFound:
		ttl = smg.cgrCfg.SessionSCfg().SessionTTL
	default: // not nil
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract %s from event, disabling session timeout for event: <%s>",
				utils.SessionS, utils.SessionTTL, s.EventStart.String()))
		ttl = time.Duration(0)
	}
	if ttl == 0 {
		return
	}
	// random delay computation
	var sessionTTLMaxDelay int64
	maxDelay, err := s.EventStart.GetDuration(utils.SessionTTLMaxDelay)
	switch err {
	case nil: // all good
	case utils.ErrNotFound:
		if smg.cgrCfg.SessionSCfg().SessionTTLMaxDelay != nil {
			maxDelay = *smg.cgrCfg.SessionSCfg().SessionTTLMaxDelay
		}
	default: // not nil
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract %s from event, disabling session timeout for event: <%s>",
				utils.SessionS, utils.SessionTTLMaxDelay, s.EventStart.String()))
		return
	}
	sessionTTLMaxDelay = maxDelay.Nanoseconds() / 1000000 // Milliseconds precision for randomness
	if sessionTTLMaxDelay != 0 {
		rand.Seed(time.Now().Unix())
		ttl += time.Duration(rand.Int63n(sessionTTLMaxDelay) * 1000000)
	}
	ttlLastUsed, err := s.EventStart.GetDurationPtrOrDefault(utils.SessionTTLLastUsed, smg.cgrCfg.SessionSCfg().SessionTTLLastUsed)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract %s from event, disabling session timeout for event: <%s>",
				utils.SessionS, utils.SessionTTLLastUsed, s.EventStart.String()))
		return
	}
	ttlUsage, err := s.EventStart.GetDurationPtrOrDefault(utils.SessionTTLUsage, smg.cgrCfg.SessionSCfg().SessionTTLUsage)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s>, cannot extract %s from event, disabling session timeout for event: <%s>",
				utils.SessionS, utils.SessionTTLUsage, s.EventStart.String()))
		return
	}
	// add to sessionTerimnations
	smg.sTsMux.Lock()
	defer smg.sTsMux.Unlock()
	if _, found := smg.sessionTerminators[s.CGRID]; found { // already there, no need to set up
		return
	}
	timer := time.NewTimer(ttl)
	endChan := make(chan bool, 1)
	terminator := &smgSessionTerminator{
		timer:       timer,
		endChan:     endChan,
		ttl:         ttl,
		ttlLastUsed: ttlLastUsed,
		ttlUsage:    ttlUsage,
	}
	smg.sessionTerminators[s.CGRID] = terminator
	go func(cgrID string) {
		select {
		case <-timer.C:
			smg.ttlTerminate(s, terminator)
		case <-endChan:
			timer.Stop()
		}
		smg.sTsMux.Lock()
		delete(smg.sessionTerminators, cgrID)
		smg.sTsMux.Unlock()
	}(s.CGRID) // Need to pass cgrID since the one from session will change during rename
}

// resetTerminatorTimer updates the timer for the session to a new ttl and terminate info
func (smg *SMGeneric) resetTerminatorTimer(cgrID string, ttl time.Duration, ttlLastUsed, ttlUsage *time.Duration) {
	smg.sTsMux.RLock()
	if st, found := smg.sessionTerminators[cgrID]; found {
		if ttl != 0 {
			st.ttl = ttl
		}
		if ttlLastUsed != nil {
			st.ttlLastUsed = ttlLastUsed
		}
		if ttlUsage != nil {
			st.ttlUsage = ttlUsage
		}
		st.timer.Reset(st.ttl)
	}
	smg.sTsMux.RUnlock()
}

// ttlTerminate is called when a session times-out
func (smg *SMGeneric) ttlTerminate(s *SMGSession, tmtr *smgSessionTerminator) {
	debitUsage := tmtr.ttl
	if tmtr.ttlUsage != nil {
		debitUsage = *tmtr.ttlUsage
	}
	aSessions := smg.getSessions(s.CGRID, false)
	if len(aSessions) == 0 { // will not continue if the session is not longer active
		return
	}
	for _, s := range aSessions[s.CGRID] {
		s.debit(debitUsage, tmtr.ttlLastUsed)
	}
	smg.sessionEnd(s.CGRID, s.TotalUsage)
	cdr, err := s.EventStart.AsCDR(smg.cgrCfg, s.Tenant, smg.Timezone)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> could not create CDR out of event %s, err: %s",
				utils.SessionS, s.EventStart.String(), err.Error()))
	}
	cdr.Usage = s.TotalUsage
	var reply string
	cgrEv := &utils.CGREvent{
		Tenant: s.Tenant,
		ID:     utils.UUIDSha1Prefix(),
		Event:  cdr.AsMapStringIface(),
	}
	if err = smg.cdrsrv.Call(utils.CdrsV2ProcessCDR, cgrEv, &reply); err != nil {
		return
	}
	if smg.resS != nil && s.ResourceID != "" {
		var reply string
		argsRU := utils.ArgRSv1ResourceUsage{
			CGREvent: utils.CGREvent{
				Tenant: s.Tenant,
				Event:  s.EventStart.AsMapInterface(),
			},
			UsageID: s.ResourceID,
			Units:   1,
		}
		if err := smg.resS.Call(utils.ResourceSv1ReleaseResources,
			argsRU, &reply); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s could not release resource with resourceID: %s",
					utils.SessionS, err.Error(), s.ResourceID))
		}
	}
	smg.replicateSessionsWithID(s.CGRID, false, smg.smgReplConns)
}

func (smg *SMGeneric) recordASession(s *SMGSession) {
	smg.aSessionsMux.Lock()
	smg.activeSessions[s.CGRID] = append(smg.activeSessions[s.CGRID], s)
	smg.setSessionTerminator(s)
	smg.indexSession(s, false)
	smg.aSessionsMux.Unlock()
}

// Remove session from session list, removes all related in case of multiple runs, true if item was found
func (smg *SMGeneric) unrecordASession(cgrID string) bool {
	smg.aSessionsMux.Lock()
	defer smg.aSessionsMux.Unlock()
	if _, found := smg.activeSessions[cgrID]; !found {
		return false
	}
	delete(smg.activeSessions, cgrID)
	smg.sTsMux.RLock()
	if st, found := smg.sessionTerminators[cgrID]; found {
		st.endChan <- true
	}
	smg.sTsMux.RUnlock()
	smg.unindexSession(cgrID, false)
	return true
}

// indexSession explores settings and builds SessionsIndex
// uses different tables and mutex-es depending on active/passive session
func (smg *SMGeneric) indexSession(s *SMGSession, passiveSessions bool) {
	idxMux := &smg.aSIMux // pointer to original mux since we cannot copy it
	ssIndx := smg.aSessionsIndex
	ssRIdx := smg.aSessionsRIndex
	if passiveSessions {
		idxMux = &smg.pSIMux
		ssIndx = smg.pSessionsIndex
		ssRIdx = smg.pSessionsRIndex
	}
	idxMux.Lock()
	defer idxMux.Unlock()
	s.RLock()
	defer s.RUnlock()
	for fieldName := range smg.ssIdxCfg {
		fieldVal, err := s.EventStart.GetString(fieldName)
		if err != nil {
			if err == utils.ErrNotFound {
				fieldVal = utils.NOT_AVAILABLE
			} else {
				utils.Logger.Err(fmt.Sprintf("<%s> Error retrieving field: %s from event: %+v", utils.SessionS, fieldName, s.EventStart))
				continue
			}
		}
		if fieldVal == "" {
			fieldVal = utils.MetaEmpty
		}
		if _, hasFieldName := ssIndx[fieldName]; !hasFieldName { // Init it here
			ssIndx[fieldName] = make(map[string]map[string]utils.StringMap)
		}
		if _, hasFieldVal := ssIndx[fieldName][fieldVal]; !hasFieldVal {
			ssIndx[fieldName][fieldVal] = make(map[string]utils.StringMap)
		}
		if _, hasFieldVal := ssIndx[fieldName][fieldVal][s.RunID]; !hasFieldVal {
			ssIndx[fieldName][fieldVal][s.RunID] = make(utils.StringMap)
		}
		ssIndx[fieldName][fieldVal][s.RunID][s.CGRID] = true
		if _, hasIt := ssRIdx[s.CGRID]; !hasIt {
			ssRIdx[s.CGRID] = make([]*riFieldNameVal, 0)
		}
		ssRIdx[s.CGRID] = append(ssRIdx[s.CGRID], &riFieldNameVal{runID: s.RunID, fieldName: fieldName, fieldValue: fieldVal})
	}
	return
}

// unindexASession removes a session from indexes
func (smg *SMGeneric) unindexSession(cgrID string, passiveSessions bool) bool {
	idxMux := &smg.aSIMux
	ssIndx := smg.aSessionsIndex
	ssRIdx := smg.aSessionsRIndex
	if passiveSessions {
		idxMux = &smg.pSIMux
		ssIndx = smg.pSessionsIndex
		ssRIdx = smg.pSessionsRIndex
	}
	idxMux.Lock()
	defer idxMux.Unlock()
	if _, hasIt := ssRIdx[cgrID]; !hasIt {
		return false
	}
	for _, riFNV := range ssRIdx[cgrID] {
		delete(ssIndx[riFNV.fieldName][riFNV.fieldValue][riFNV.runID], cgrID)
		if len(ssIndx[riFNV.fieldName][riFNV.fieldValue][riFNV.runID]) == 0 {
			delete(ssIndx[riFNV.fieldName][riFNV.fieldValue], riFNV.runID)
		}
		if len(ssIndx[riFNV.fieldName][riFNV.fieldValue]) == 0 {
			delete(ssIndx[riFNV.fieldName], riFNV.fieldValue)
		}
		if len(ssIndx[riFNV.fieldName]) == 0 {
			delete(ssIndx, riFNV.fieldName)
		}
	}
	delete(ssRIdx, cgrID)
	return true
}

// getSessionIDsMatchingIndexes will check inside indexes if it can find sessionIDs matching all filters
// matchedIndexes returns map[matchedFieldName]possibleMatchedFieldVal so we optimize further to avoid checking them
func (smg *SMGeneric) getSessionIDsMatchingIndexes(fltrs map[string]string,
	passiveSessions bool) (utils.StringMap, map[string]string) {
	idxMux := &smg.aSIMux
	ssIndx := smg.aSessionsIndex
	if passiveSessions {
		idxMux = &smg.pSIMux
		ssIndx = smg.pSessionsIndex
	}
	idxMux.RLock()
	defer idxMux.RUnlock()
	matchedIndexes := make(map[string]string)
	matchingSessions := make(utils.StringMap)
	runID := fltrs[utils.RunID]
	checkNr := 0
	var findFunc func(cgrID, fltrName, fltrVal string) bool
	if runID == "" {
		findFunc = func(cgrID, fltrName, fltrVal string) bool {
			for runID := range ssIndx[fltrName][fltrVal] {
				for cgrmID := range ssIndx[fltrName][fltrVal][runID] {
					if cgrID == cgrmID {
						return true
					}
				}
			}
			return false
		}
	} else { // We know the RunID
		findFunc = func(cgrID, fltrName, fltrVal string) bool {
			for cgrmID := range ssIndx[fltrName][fltrVal][runID] {
				if cgrID == cgrmID {
					return true
				}
			}
			return false
		}
	}
	for fltrName, fltrVal := range fltrs {
		if fltrName == utils.RunID {
			continue
		}
		if _, hasFldName := ssIndx[fltrName]; !hasFldName {
			continue
		}
		checkNr += 1
		if _, hasFldVal := ssIndx[fltrName][fltrVal]; !hasFldVal {
			matchedIndexes[fltrName] = utils.META_NONE
			return make(utils.StringMap), matchedIndexes
		}
		matchedIndexes[fltrName] = fltrVal
		if checkNr == 1 { // First run will init the MatchingSessions
			if runID == "" {
				for runID := range ssIndx[fltrName][fltrVal] {
					matchingSessions.Join(ssIndx[fltrName][fltrVal][runID])
				}
			} else { // We know the RunID
				matchingSessions = ssIndx[fltrName][fltrVal][runID]
			}
			continue
		}
		// Higher run, takes out non matching indexes
		for cgrID := range matchingSessions {
			if !findFunc(cgrID, fltrName, fltrVal) {
				delete(matchingSessions, cgrID)
			}
		}
	}
	return matchingSessions.Clone(), matchedIndexes
}

// getSessionIDsForPrefix works with session relocation returning list of sessions with ID matching prefix for OriginID field
func (smg *SMGeneric) getSessionIDsForPrefix(prefix string,
	passiveSessions bool) (cgrIDs []string) {
	idxMux := &smg.aSIMux
	ssIndx := smg.aSessionsIndex
	if passiveSessions {
		idxMux = &smg.pSIMux
		ssIndx = smg.pSessionsIndex
	}
	idxMux.RLock()
	// map[OriginID:map[12372-1:map[*default:511654dc4da7ce4706276cb458437cdd81d0e2b3]]]
	for originID := range ssIndx[utils.OriginID] {
		if strings.HasPrefix(originID, prefix) {
			if _, hasDefaultRun := ssIndx[utils.OriginID][originID][utils.META_DEFAULT]; hasDefaultRun {
				cgrIDs = append(cgrIDs,
					ssIndx[utils.OriginID][originID][utils.META_DEFAULT].Slice()...)
			}
		}
	}
	idxMux.RUnlock()
	return
}

// v1ForkSessions is using DerivedChargers for session forking
func (smg *SMGeneric) v1ForkSessions(tnt string, evStart *engine.SafEvent,
	clntConn rpcclient.RpcClientConnection, cgrID, resourceID string,
	handlePseudo bool) (ss []*SMGSession, err error) {
	cdr, err := evStart.AsCDR(smg.cgrCfg, tnt, smg.Timezone)
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> could not convert event: %s to CDR, err: %s",
			utils.SessionS, evStart.String(), err.Error()))
		return nil, err
	}
	var sessionRuns []*engine.SessionRun
	if err := smg.rals.Call("Responder.GetSessionRuns",
		cdr, &sessionRuns); err != nil {
		return nil, err
	}
	noneSession := []*SMGSession{
		{Tenant: tnt, CGRID: cgrID,
			ResourceID: resourceID, EventStart: evStart,
			RunID: utils.META_NONE, Timezone: smg.Timezone,
			rals: smg.rals, cdrsrv: smg.cdrsrv,
			clntConn: clntConn}}
	handledSessions := []string{utils.META_PREPAID}
	if handlePseudo {
		handledSessions = append(handledSessions, utils.META_PSEUDOPREPAID)
	}
	for _, sessionRun := range sessionRuns {
		if !utils.IsSliceMember(handledSessions, sessionRun.RequestType) {
			continue // not forking non-prepaid session
		}
		ss = append(ss,
			&SMGSession{CGRID: cgrID, Tenant: tnt,
				ResourceID: resourceID, EventStart: evStart,
				RunID: sessionRun.DerivedCharger.RunID, Timezone: smg.Timezone,
				rals: smg.rals, cdrsrv: smg.cdrsrv,
				CD: sessionRun.CallDescriptor, clntConn: clntConn,
				clientProto: smg.cgrCfg.SessionSCfg().ClientProtocol})
	}
	if len(ss) == 0 { //  we have no *prepaid session to work with
		return noneSession, nil
	}
	return
}

// v2ForkSessions is using ChargerS for session forking
func (smg *SMGeneric) v2ForkSessions(tnt string, evStart *engine.SafEvent,
	clntConn rpcclient.RpcClientConnection,
	cgrID, resourceID string, handlePseudo bool) (ss []*SMGSession, err error) {
	cgrEv := &utils.CGREvent{
		Tenant: tnt,
		ID:     utils.UUIDSha1Prefix(),
		Event:  evStart.AsMapInterface(),
	}
	var chrgrs []*engine.ChrgSProcessEventReply
	if err := smg.chargerS.Call(utils.ChargerSv1ProcessEvent, cgrEv, &chrgrs); err != nil {
		if err.Error() == utils.ErrNotFound.Error() {
			return nil, utils.ErrNoActiveSession
		}
		return nil, err
	}
	noneSession := []*SMGSession{
		{CGRID: cgrID, ResourceID: resourceID, EventStart: evStart,
			RunID: utils.META_NONE, Timezone: smg.Timezone,
			rals: smg.rals, cdrsrv: smg.cdrsrv,
			clntConn: clntConn}}
	handledSessions := []string{utils.META_PREPAID}
	if handlePseudo {
		handledSessions = append(handledSessions, utils.META_PSEUDOPREPAID)
	}
	for _, chrgr := range chrgrs {
		evStart := engine.NewSafEvent(chrgr.CGREvent.Event)
		if !utils.IsSliceMember(handledSessions, evStart.GetStringIgnoreErrors(utils.RequestType)) {
			continue // not forking non-prepaid session
		}
		startTime := evStart.GetTimeIgnoreErrors(utils.AnswerTime, smg.Timezone)
		if startTime.IsZero() { // AnswerTime not parsable, try SetupTime
			startTime = evStart.GetTimeIgnoreErrors(utils.SetupTime, smg.Timezone)
		}
		cd := &engine.CallDescriptor{
			CgrID:       cgrID,
			RunID:       evStart.GetStringIgnoreErrors(utils.RunID),
			TOR:         evStart.GetStringIgnoreErrors(utils.ToR),
			Direction:   utils.OUT,
			Tenant:      tnt,
			Category:    evStart.GetStringIgnoreErrors(utils.Category),
			Subject:     evStart.GetStringIgnoreErrors(utils.Subject),
			Account:     evStart.GetStringIgnoreErrors(utils.Account),
			Destination: evStart.GetStringIgnoreErrors(utils.Destination),
			TimeStart:   startTime,
			TimeEnd:     startTime.Add(evStart.GetDurationIgnoreErrors(utils.Usage)),
			ExtraFields: evStart.AsMapStringIgnoreErrors(utils.NewStringMap(utils.PrimaryCdrFields...)),
		}
		ss = append(ss,
			&SMGSession{CGRID: cgrID, Tenant: tnt,
				ResourceID: resourceID,
				EventStart: evStart,
				RunID:      evStart.GetStringIgnoreErrors(utils.RunID),
				Timezone:   smg.Timezone,
				rals:       smg.rals, cdrsrv: smg.cdrsrv,
				CD: cd, clntConn: clntConn,
				clientProto: smg.cgrCfg.SessionSCfg().ClientProtocol})
	}
	if len(ss) == 0 { //  we have no *prepaid session to work with
		return noneSession, nil
	}
	return
}

// sessionStart will handle a new session, pass the connectionId so we can communicate on disconnect request
func (smg *SMGeneric) sessionStart(tnt, cgrID string, evStart *engine.SafEvent,
	clntConn rpcclient.RpcClientConnection, resourceID string,
	dbtItval time.Duration) (err error) {
	var ss []*SMGSession
	if smg.chargerS == nil { // old way of session forking
		ss, err = smg.v1ForkSessions(tnt, evStart, clntConn, cgrID, resourceID, false)
	} else {
		ss, err = smg.v2ForkSessions(tnt, evStart, clntConn, cgrID, resourceID, false)
	}
	if err != nil {
		return
	}
	stopDebitChan := make(chan struct{})
	for _, s := range ss {
		smg.recordASession(s)
		if s.RunID != utils.META_NONE &&
			dbtItval != 0 {
			s.stopDebit = stopDebitChan
			go s.debitLoop(dbtItval)
		}
	}
	return
}

// sessionUpdate will reset terminator, perform debits and replicate sessions
func (smg *SMGeneric) sessionUpdate(tnt, cgrID string, ev *engine.SafEvent,
	clnt rpcclient.RpcClientConnection, resourceID string,
	dbtItval time.Duration) (maxUsage time.Duration, err error) {
	// make sure the session exists, otherwise create
	aSessions := smg.getSessions(cgrID, false)
	if len(aSessions) == 0 {
		if aSessions = smg.passiveToActive(cgrID); len(aSessions) == 0 {
			if ev.HasField(utils.InitialOriginID) {
				initialCGRID := utils.Sha1(
					ev.GetStringIgnoreErrors(utils.InitialOriginID),
					ev.GetStringIgnoreErrors(utils.OriginHost))
				err = smg.sessionRelocate(initialCGRID,
					cgrID, ev.GetStringIgnoreErrors(utils.OriginID))
				smg.replicateSessionsWithID(initialCGRID, false, smg.smgReplConns) // report changes
			}
			if !ev.HasField(utils.InitialOriginID) || err == utils.ErrNotFound { // create a new  session with this update
				err = smg.sessionStart(tnt, cgrID, ev, clnt, resourceID, dbtItval)
			}
			if err != nil {
				return
			}
			aSessions = smg.getSessions(cgrID, false) // try again to populate after starting above
			if len(aSessions) == 0 {
				utils.Logger.Err(
					fmt.Sprintf("<%s> no active sessions for event: <%s>",
						utils.SessionS, cgrID))
				err = rpcclient.ErrSessionNotFound
				return
			}
		}
	}
	defer smg.replicateSessionsWithID(cgrID, false, smg.smgReplConns)

	var sesTTL, evLastUsed time.Duration
	if sesTTL, err = getSessionTTL(ev, smg.cgrCfg.SessionSCfg().SessionTTL,
		smg.cgrCfg.SessionSCfg().SessionTTLMaxDelay); err != nil {
		return
	}
	var ttlLastUsed, ttlUsage, lastUsed *time.Duration
	if ttlLastUsed, err = ev.GetDurationPtrOrDefault(utils.SessionTTLLastUsed,
		smg.cgrCfg.SessionSCfg().SessionTTLLastUsed); err != nil {
		return
	}
	if ttlUsage, err = ev.GetDurationPtrOrDefault(utils.SessionTTLUsage,
		smg.cgrCfg.SessionSCfg().SessionTTLUsage); err != nil {
		return
	}
	smg.resetTerminatorTimer(cgrID, sesTTL, ttlLastUsed, ttlUsage)
	if evLastUsed, err = ev.GetDuration(utils.LastUsed); err == nil {
		lastUsed = &evLastUsed
	} else if err != utils.ErrNotFound {
		return
	}
	if maxUsage, err = ev.GetDuration(utils.Usage); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		maxUsage = smg.cgrCfg.SessionSCfg().MaxCallDuration
		err = nil
	}
	for _, s := range aSessions[cgrID] {
		var maxDur time.Duration
		var maxUsageSet bool
		if s.RunID == utils.META_NONE {
			maxDur = time.Duration(-1)
		} else if maxDur, err = s.debit(maxUsage, lastUsed); err != nil {
			return
		}
		if maxDur == time.Duration(-1) && !maxUsageSet {
			maxUsage = maxDur
		} else if maxDur < maxUsage {
			maxUsage = maxDur
		}
	}
	return
}

// sessionEnd will end a session from outside
func (smg *SMGeneric) sessionEnd(cgrID string, usage time.Duration) (err error) {
	ss := smg.getSessions(cgrID, false)
	if len(ss) == 0 {
		if ss = smg.passiveToActive(cgrID); len(ss) == 0 {
			return // ToDo: handle here also debits
		}
	}
	if !smg.unrecordASession(cgrID) { // Unreference it early so we avoid concurrency
		return // Did not find the session so no need to close it anymore
	}
	for idx, s := range ss[cgrID] {
		if s.RunID == utils.META_NONE {
			continue
		}
		s.TotalUsage = usage // save final usage as totalUsage
		if idx == 0 && s.stopDebit != nil {
			close(s.stopDebit) // Stop automatic debits
		}
		aTime, err := s.EventStart.GetTime(utils.AnswerTime, smg.Timezone)
		if err != nil || aTime.IsZero() {
			utils.Logger.Warning(fmt.Sprintf("<%s> could not retrieve answer time for session: %s, runId: %s, aTime: %+v, error: %v",
				utils.SessionS, cgrID, s.RunID, aTime, err))
			continue // Unanswered session
		}
		if err := s.close(usage); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Could not close session: %s, runId: %s, error: %s", utils.SessionS, cgrID, s.RunID, err.Error()))
		}
		if err := s.storeSMCost(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Could not save session: %s, runId: %s, error: %s", utils.SessionS, cgrID, s.RunID, err.Error()))
		}
	}
	return
}

// sessionRelocate is used when an update will relocate an initial session (eg multiple data streams)
func (smg *SMGeneric) sessionRelocate(initialID, cgrID, newOriginID string) (err error) {
	if utils.IsSliceMember([]string{initialID, cgrID, newOriginID}, "") { // Not allowed empty params here
		return utils.ErrMandatoryIeMissing
	}
	ssNew := smg.getSessions(cgrID, false)
	if len(ssNew) != 0 { // Already relocated
		return
	}
	if pSSNew := smg.getSessions(cgrID, true); len(pSSNew) != 0 { // passive sessions recorded, will be recovered so no need of relocation
		return
	}
	ss := smg.getSessions(initialID, false)
	if len(ss) == 0 { // No need of relocation
		if ss = smg.passiveToActive(initialID); len(ss) == 0 {
			return utils.ErrNotFound
		}
	}
	for i, s := range ss[initialID] {
		s.Lock()
		s.CGRID = cgrID                               // Overwrite initial CGRID with new one
		s.EventStart.Set(utils.CGRID, cgrID)          // Overwrite CGRID for final CDR
		s.EventStart.Set(utils.OriginID, newOriginID) // Overwrite OriginID for session indexing
		s.Unlock()
		smg.recordASession(s)
		if i == 0 {
			smg.unrecordASession(initialID)
		}
	}
	return
}

// replicateSessions will replicate session based on configuration
func (smg *SMGeneric) replicateSessionsWithID(cgrID string, passiveSessions bool, smgReplConns []*SMGReplicationConn) (err error) {
	if len(smgReplConns) == 0 ||
		(smg.cgrCfg.SessionSCfg().DebitInterval != 0 && !passiveSessions) { // Replicating active not supported
		return
	}
	ssMux := &smg.aSessionsMux
	ssMp := smg.activeSessions // reference it so we don't overwrite the new map without protection
	if passiveSessions {
		ssMux = &smg.pSessionsMux
		ssMp = smg.passiveSessions
	}
	ssMux.RLock()
	ss := ssMp[cgrID]
	if len(ss) != 0 {
		ss[0].RLock() // lock session so we can clone it after releasing the map lock
	}
	ssMux.RUnlock()
	ssCln := make([]*SMGSession, len(ss))
	for i, s := range ss {
		ssCln[i] = s.Clone()
	}
	if len(ss) != 0 {
		ss[0].RUnlock()
	}
	var wg sync.WaitGroup
	for _, rplConn := range smgReplConns {
		if rplConn.Synchronous {
			wg.Add(1)
		}
		go func(conn rpcclient.RpcClientConnection, sync bool, ss []*SMGSession) {
			var reply string
			argSet := ArgsSetPassiveSessions{CGRID: cgrID, Sessions: ss}
			conn.Call("SMGenericV1.SetPassiveSessions", argSet, &reply)
			if sync {
				wg.Done()
			}
		}(rplConn.Connection, rplConn.Synchronous, ssCln)
	}
	wg.Wait() // wait for synchronous replication to finish
	return
}

// getSessions is used to return in a thread-safe manner active or passive sessions
func (smg *SMGeneric) getSessions(cgrID string, passiveSessions bool) (aSS map[string][]*SMGSession) {
	ssMux := &smg.aSessionsMux
	ssMp := smg.activeSessions // reference it so we don't overwrite the new map without protection
	if passiveSessions {
		ssMux = &smg.pSessionsMux
		ssMp = smg.passiveSessions
	}
	ssMux.RLock()
	defer ssMux.RUnlock()
	aSS = make(map[string][]*SMGSession)
	if len(cgrID) == 0 {
		for k, v := range ssMp {
			aSS[k] = v // Copy to avoid concurrency on sessions map
		}
		return
	}
	if ss, hasCGRID := ssMp[cgrID]; hasCGRID {
		aSS[cgrID] = ss
	}
	return
}

// setPassiveSession is called when a session is set via RPC in passive sessions table
func (smg *SMGeneric) setPassiveSessions(cgrID string, ss []*SMGSession) (err error) {
	if len(ss) == 0 {
		return
	}
	for _, cacheKey := range []string{"InitiateSession" + cgrID,
		"UpdateSession" + cgrID, "TerminateSession" + cgrID} {
		if _, err := smg.responseCache.Get(cacheKey); err == nil { // Stop processing passive when there has been an update over active RPC
			smg.deletePassiveSessions(cgrID)
			return ErrActiveSession
		}
	}
	smg.unrecordASession(cgrID)
	smg.pSessionsMux.Lock()
	smg.passiveSessions[cgrID] = ss
	smg.pSessionsMux.Unlock()
	for _, s := range ss {
		smg.indexSession(s, true)
	}
	return
}

// remPassiveSession is called when a session is removed via RPC from passive sessions table
// ToDo: test
func (smg *SMGeneric) removePassiveSessions(cgrID string) (err error) {
	for _, cacheKey := range []string{"InitiateSession" + cgrID, "UpdateSession" + cgrID, "TerminateSession" + cgrID} {
		if _, err := smg.responseCache.Get(cacheKey); err == nil { // Stop processing passive when there has been an update over active RPC
			smg.deletePassiveSessions(cgrID)
			return ErrActiveSession
		}
	}
	smg.unrecordASession(cgrID) // just in case there is an active session
	smg.deletePassiveSessions(cgrID)
	return
}

// deletePassiveSessions is used to remove a reference from the passiveSessions table
// ToDo: test it
func (smg *SMGeneric) deletePassiveSessions(cgrID string) {
	smg.unindexSession(cgrID, true)
	smg.pSessionsMux.Lock()
	delete(smg.passiveSessions, cgrID)
	smg.pSessionsMux.Unlock()
}

// passiveToActive will transition the sessions from passive to active table
// ToDo: test
func (smg *SMGeneric) passiveToActive(cgrID string) (pSessions map[string][]*SMGSession) {
	pSessions = smg.getSessions(cgrID, true)
	if len(pSessions) == 0 {
		return
	}
	for _, s := range pSessions[cgrID] {
		smg.recordASession(s)
		s.rals = smg.rals
		s.cdrsrv = smg.cdrsrv
	}
	smg.deletePassiveSessions(cgrID)
	return
}

// asActiveSessions returns sessions from either active or passive table as []*ActiveSession
func (smg *SMGeneric) asActiveSessions(fltrs map[string]string, count, passiveSessions bool) (aSessions []*ActiveSession, counter int, err error) {
	aSessions = make([]*ActiveSession, 0) // Make sure we return at least empty list and not nil
	// Check first based on indexes so we can downsize the list of matching sessions
	matchingSessionIDs, checkedFilters := smg.getSessionIDsMatchingIndexes(fltrs, passiveSessions)
	if len(matchingSessionIDs) == 0 && len(checkedFilters) != 0 {
		return
	}
	for fltrFldName := range fltrs {
		if _, alreadyChecked := checkedFilters[fltrFldName]; alreadyChecked && fltrFldName != utils.RunID { // Optimize further checks, RunID should stay since it can create bugs
			delete(fltrs, fltrFldName)
		}
	}
	var remainingSessions []*SMGSession // Survived index matching
	var ss map[string][]*SMGSession
	if passiveSessions {
		ss = smg.getSessions(fltrs[utils.CGRID], true)
	} else {
		ss = smg.getSessions(fltrs[utils.CGRID], false)
	}
	for cgrID, sGrp := range ss {
		if _, hasCGRID := matchingSessionIDs[cgrID]; !hasCGRID && len(checkedFilters) != 0 {
			continue
		}
		for _, s := range sGrp {
			remainingSessions = append(remainingSessions, s)
		}
	}
	if len(fltrs) != 0 { // Still have some filters to match
		for i := 0; i < len(remainingSessions); {
			if !remainingSessions[i].EventStart.HasField(utils.RunID) {
				remainingSessions[i].EventStart.Set(utils.RunID, utils.META_DEFAULT)
			}
			matchingAll := true
			for fltrFldName, fltrFldVal := range fltrs {
				if remainingSessions[i].EventStart.GetStringIgnoreErrors(fltrFldName) != fltrFldVal { // No Match
					matchingAll = false
					break
				}
			}
			if !matchingAll {
				remainingSessions = append(remainingSessions[:i], remainingSessions[i+1:]...)
				continue // if we have stripped, don't increase index so we can check next element by next run
			}
			i++
		}
	}
	if count {
		return nil, len(remainingSessions), nil
	}
	for _, s := range remainingSessions {
		aSessions = append(aSessions, s.AsActiveSession(smg.Timezone)) // Expensive for large number of sessions
	}
	return
}

// Methods to apply on sessions, mostly exported through RPC/Bi-RPC

// MaxUsage calculates maximum usage allowed for given event
func (smg *SMGeneric) GetMaxUsage(tnt string, ev *engine.SafEvent) (maxUsage time.Duration, err error) {
	cgrID := GetSetCGRID(ev)
	cacheKey := "MaxUsage" + cgrID
	if item, err := smg.responseCache.Get(cacheKey); err == nil && item != nil {
		return (item.Value.(time.Duration)), item.Err
	}
	defer smg.responseCache.Cache(cacheKey, &utils.ResponseCacheItem{Value: maxUsage, Err: err})
	if has := ev.HasField(utils.Usage); !has { // make sure we have a minimum duration configured
		ev.Set(utils.Usage, smg.cgrCfg.SessionSCfg().MaxCallDuration)
	}
	// fork sessions
	var ss []*SMGSession
	if smg.chargerS == nil { // old way of session forking
		ss, err = smg.v1ForkSessions(tnt, ev, nil, cgrID, "", true)
	} else {
		ss, err = smg.v2ForkSessions(tnt, ev, nil, cgrID, "", true)
	}
	if err != nil {
		return
	}
	var minUsage *time.Duration // find out the minimum usage
	for _, s := range ss {
		if s.RunID == utils.META_NONE {
			minUsage = utils.DurationPointer(-1)
			break
		}
		var maxDur time.Duration
		if err = smg.rals.Call("Responder.GetMaxSessionTime", s.CD, &maxDur); err != nil {
			return
		}
		if minUsage == nil || maxDur < *minUsage {
			minUsage = &maxDur
		}
	}
	maxUsage = *minUsage
	if maxUsage != time.Duration(-1) &&
		maxUsage < smg.cgrCfg.SessionSCfg().MinCallDuration {
		return 0, errors.New("UNAUTHORIZED_MIN_DURATION")
	}
	return
}

// Called on session start
func (smg *SMGeneric) InitiateSession(tnt string, ev *engine.SafEvent,
	clnt rpcclient.RpcClientConnection, resourceID string,
	dbtItval time.Duration) (maxUsage time.Duration, err error) {
	cgrID := GetSetCGRID(ev)
	_, err = guardian.Guardian.Guard(func() (iface interface{}, err error) { // Lock it on CGRID level
		cacheKey := "InitiateSession" + cgrID
		if item, err := smg.responseCache.Get(cacheKey); err == nil && item != nil {
			return item.Value.(time.Duration), item.Err
		}
		defer smg.responseCache.Cache(cacheKey,
			&utils.ResponseCacheItem{Value: maxUsage, Err: err}) // schedule response caching
		smg.deletePassiveSessions(cgrID)
		if err = smg.sessionStart(tnt, cgrID, ev, clnt, resourceID, dbtItval); err != nil {
			smg.sessionEnd(cgrID, 0)
			return
		}
		if dbtItval != 0 { // Session handled by debit loop
			maxUsage = time.Duration(-1)
			return
		}
		maxUsage, err = smg.sessionUpdate(tnt, cgrID, ev, clnt, resourceID, dbtItval)
		if err != nil || maxUsage == 0 {
			smg.sessionEnd(cgrID, 0)
		}
		return
	}, smg.cgrCfg.GeneralCfg().LockingTimeout, cgrID)
	return
}

// Execute debits for usage/maxUsage
func (smg *SMGeneric) UpdateSession(tnt string, ev *engine.SafEvent,
	clnt rpcclient.RpcClientConnection, resourceID string,
	dbtItval time.Duration) (maxUsage time.Duration, err error) {
	cgrID := GetSetCGRID(ev)
	_, err = guardian.Guardian.Guard(func() (iface interface{}, err error) { // Lock it on CGRID level
		cacheKey := "UpdateSession" + cgrID
		if item, err := smg.responseCache.Get(cacheKey); err == nil && item != nil {
			return item.Value.(time.Duration), item.Err
		}
		defer smg.responseCache.Cache(cacheKey,
			&utils.ResponseCacheItem{Value: maxUsage, Err: err})
		maxUsage, err = smg.sessionUpdate(tnt, cgrID, ev, clnt, resourceID, dbtItval)
		if err != nil {
			smg.sessionEnd(cgrID, 0)
		}
		return
	}, smg.cgrCfg.GeneralCfg().LockingTimeout, cgrID)
	return
}

// Called on session end, should stop debit loop
func (smg *SMGeneric) TerminateSession(tnt string, ev *engine.SafEvent,
	clnt rpcclient.RpcClientConnection, resourceID string,
	dbtItvl time.Duration) (err error) {
	cgrID := GetSetCGRID(ev)
	_, err = guardian.Guardian.Guard(func() (iface interface{}, err error) { // Lock it on CGRID level
		cacheKey := "TerminateSession" + cgrID
		if item, err := smg.responseCache.Get(cacheKey); err == nil && item != nil {
			return nil, item.Err
		}
		defer smg.responseCache.Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
		if ev.HasField(utils.InitialOriginID) {
			initialCGRID := utils.Sha1(
				ev.GetStringIgnoreErrors(utils.InitialOriginID),
				ev.GetStringIgnoreErrors(utils.OriginHost))
			err = smg.sessionRelocate(initialCGRID, cgrID,
				ev.GetStringIgnoreErrors(utils.OriginID))
			if err == utils.ErrNotFound { // Session was already relocated, create a new  session with this update
				err = smg.sessionStart(tnt, cgrID, ev, clnt, resourceID, dbtItvl)
			}
			if err != nil && err != utils.ErrMandatoryIeMissing {
				return
			}
			smg.replicateSessionsWithID(initialCGRID, false, smg.smgReplConns)
		}
		sessionIDs := []string{cgrID}
		if ev.HasField(utils.OriginIDPrefix) { // OriginIDPrefix is present, OriginID will not be anymore considered
			sessionIDPrefix := ev.GetStringIgnoreErrors(utils.OriginIDPrefix)
			if sessionIDs = smg.getSessionIDsForPrefix(sessionIDPrefix, false); len(sessionIDs) == 0 {
				sessionIDs = smg.getSessionIDsForPrefix(sessionIDPrefix, true)
				for _, sessionID := range sessionIDs { // activate sessions for prefix
					smg.passiveToActive(sessionID)
				}
			}
		}
		usage, errUsage := ev.GetDuration(utils.Usage)
		var lastUsed time.Duration
		if errUsage != nil {
			if errUsage != utils.ErrNotFound {
				err = errUsage
				return
			}
			lastUsed, err = ev.GetDuration(utils.LastUsed)
			if err != nil {
				if err == utils.ErrNotFound {
					err = utils.ErrMandatoryIeMissing
				}
				return
			}
		}
		var hasActiveSession bool
		for _, sessionID := range sessionIDs {
			aSessions := smg.getSessions(sessionID, false)
			if len(aSessions) == 0 {
				if aSessions = smg.passiveToActive(cgrID); len(aSessions) == 0 {
					utils.Logger.Err(fmt.Sprintf("<%s> terminate with no active sessions for cgrID: <%s>", utils.SessionS, cgrID))
					continue
				}
			}
			hasActiveSession = true
			defer smg.replicateSessionsWithID(sessionID, false, smg.smgReplConns)
			s := aSessions[sessionID][0]
			if errUsage != nil {
				usage = s.TotalUsage - s.LastUsage + lastUsed
			}
			if errSEnd := smg.sessionEnd(sessionID, usage); errSEnd != nil {
				err = errSEnd // Last error will be the one returned as API result
			}
		}
		if !hasActiveSession {
			err = rpcclient.ErrSessionNotFound
			return
		}
		return
	}, smg.cgrCfg.GeneralCfg().LockingTimeout, cgrID)
	return
}

// Processes one time events (eg: SMS)
func (smg *SMGeneric) ChargeEvent(tnt string, ev *engine.SafEvent) (maxUsage time.Duration, err error) {
	cgrID := GetSetCGRID(ev)
	cacheKey := "ChargeEvent" + cgrID
	if item, err := smg.responseCache.Get(cacheKey); err == nil && item != nil {
		return item.Value.(time.Duration), item.Err
	}
	defer smg.responseCache.Cache(cacheKey, &utils.ResponseCacheItem{Value: maxUsage, Err: err})
	// fork sessions
	var ss []*SMGSession
	if smg.chargerS == nil { // old way of session forking
		ss, err = smg.v1ForkSessions(tnt, ev, nil, cgrID, "", false)
	} else {
		ss, err = smg.v2ForkSessions(tnt, ev, nil, cgrID, "", false)
	}
	if err != nil {
		return
	}
	// debit each forked session
	var maxDur *time.Duration // Avoid differences between default 0 and received 0
	for _, s := range ss {
		var durDebit time.Duration
		if durDebit, err = s.debit(s.CD.GetDuration(), nil); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Could not Debit CD: %+v, RunID: %s, error: %s",
				utils.SessionS, s.CD, s.RunID, err.Error()))
			break
		}
		if durDebit == 0 {
			err = utils.ErrInsufficientCredit
			break
		}
		if maxDur == nil || *maxDur > durDebit {
			maxDur = utils.DurationPointer(durDebit)
		}
	}
	if err != nil { // Refund the ones already taken since we have error on one of the debits
		for _, s := range ss {
			if err := s.close(time.Duration(0)); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> error: %s closing session with runID: %s",
					utils.SessionS, err.Error(), s.RunID))
			}
		}
		return
	}
	// store session log
	for _, s := range ss {
		if errStore := s.storeSMCost(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s storing session with runID: %s",
				utils.SessionS, errStore.Error(), s.RunID))
			err = ErrPartiallyExecuted
		}
	}
	if err != nil {
		return
	}
	if maxDur != nil {
		maxUsage = *maxDur
	}
	return
}

func (smg *SMGeneric) ProcessCDR(tnt string, ev *engine.SafEvent) (err error) {
	cgrID := GetSetCGRID(ev)
	cacheKey := "ProcessCDR" + cgrID
	if item, err := smg.responseCache.Get(cacheKey); err == nil && item != nil {
		return item.Err
	}
	defer smg.responseCache.Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
	cgrEv := &utils.CGREvent{
		Tenant: tnt,
		ID:     utils.UUIDSha1Prefix(),
		Event:  ev.AsMapInterface(),
	}
	var reply string
	if err = smg.cdrsrv.Call(utils.CdrsV2ProcessCDR, cgrEv, &reply); err != nil {
		return
	}
	return
}

func (smg *SMGeneric) Connect() error {
	if smg.cgrCfg.SessionSCfg().ChannelSyncInterval != 0 {
		go func() {
			for { // Schedule sync channels to run repetately
				time.Sleep(smg.cgrCfg.SessionSCfg().ChannelSyncInterval)
				smg.syncSessions()
			}

		}()
	}
	return nil
}

// System shutdown
func (smg *SMGeneric) Shutdown() error {
	for ssId := range smg.getSessions("", false) { // Force sessions shutdown
		smg.sessionEnd(ssId, time.Duration(smg.cgrCfg.MaxCallDuration))
	}
	return nil
}

// RpcClientConnection interface
func (smg *SMGeneric) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return smg.CallBiRPC(nil, serviceMethod, args, reply) // Capture the version part out of original call
}

// Part of utils.BiRPCServer to help internal connections do calls over rpcclient.RpcClientConnection interface
func (smg *SMGeneric) CallBiRPC(clnt rpcclient.RpcClientConnection,
	serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// get method BiRPCV1.Method
	method := reflect.ValueOf(smg).MethodByName(
		"BiRPC" + parts[0][len(parts[0])-2:] + parts[1]) // Inherit the version V1 in the method name and add prefix
	if !method.IsValid() {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// construct the params
	var clntVal reflect.Value
	if clnt == nil {
		clntVal = reflect.New(
			reflect.TypeOf(new(utils.BiRPCInternalClient))).Elem() // Kinda cheat since we make up a type here
	} else {
		clntVal = reflect.ValueOf(clnt)
	}
	params := []reflect.Value{clntVal, reflect.ValueOf(args),
		reflect.ValueOf(reply)}
	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}

func (smg *SMGeneric) BiRPCV1GetMaxUsage(clnt rpcclient.RpcClientConnection,
	ev engine.MapEvent, maxUsage *float64) error {
	maxUsageDur, err := smg.GetMaxUsage(
		utils.FirstNonEmpty(ev.GetStringIgnoreErrors(utils.Tenant),
			smg.cgrCfg.GeneralCfg().DefaultTenant),
		engine.NewSafEvent(ev))
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if maxUsageDur == time.Duration(-1) {
		*maxUsage = -1.0
	} else {
		*maxUsage = maxUsageDur.Seconds()
	}
	return nil
}

// BiRPCV2GetMaxUsage returns the maximum usage as duration/int64
func (smg *SMGeneric) BiRPCV2GetMaxUsage(clnt rpcclient.RpcClientConnection,
	ev engine.MapEvent, maxUsage *time.Duration) error {
	maxUsageDur, err := smg.GetMaxUsage(
		utils.FirstNonEmpty(ev.GetStringIgnoreErrors(utils.Tenant),
			smg.cgrCfg.GeneralCfg().DefaultTenant),
		engine.NewSafEvent(ev))
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*maxUsage = maxUsageDur
	return nil
}

// Called on session start, returns the maximum number of seconds the session can last
func (smg *SMGeneric) BiRPCV1InitiateSession(clnt rpcclient.RpcClientConnection,
	ev engine.MapEvent, maxUsage *float64) (err error) {
	var minMaxUsage time.Duration
	tnt := utils.FirstNonEmpty(ev.GetStringIgnoreErrors(utils.Tenant),
		smg.cgrCfg.GeneralCfg().DefaultTenant)
	if minMaxUsage, err = smg.InitiateSession(tnt,
		engine.NewSafEvent(ev), clnt, "",
		smg.cgrCfg.SessionSCfg().DebitInterval); err != nil {
		if err != rpcclient.ErrSessionNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	if minMaxUsage == time.Duration(-1) {
		// handle auth for OpenSIPS 2.1
		var authUsage time.Duration
		if authUsage, err = smg.GetMaxUsage(tnt, engine.NewSafEvent(ev)); err != nil {
			return
		}
		if authUsage != time.Duration(-1) {
			*maxUsage = authUsage.Seconds()
		}
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return
}

// BiRPCV2InitiateSession initiates a new session, returns the maximum duration the session can last
func (smg *SMGeneric) BiRPCV2InitiateSession(clnt rpcclient.RpcClientConnection,
	ev engine.MapEvent, maxUsage *time.Duration) (err error) {
	var minMaxUsage time.Duration
	if minMaxUsage, err = smg.InitiateSession(
		utils.FirstNonEmpty(ev.GetStringIgnoreErrors(utils.Tenant),
			smg.cgrCfg.GeneralCfg().DefaultTenant),
		engine.NewSafEvent(ev), clnt, "",
		smg.cgrCfg.SessionSCfg().DebitInterval); err != nil {
		if err != rpcclient.ErrSessionNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	} else {
		*maxUsage = minMaxUsage
	}
	return
}

// Interim updates, returns remaining duration from the RALs
func (smg *SMGeneric) BiRPCV1UpdateSession(clnt rpcclient.RpcClientConnection,
	ev engine.MapEvent, maxUsage *float64) (err error) {
	var minMaxUsage time.Duration
	if minMaxUsage, err = smg.UpdateSession(
		utils.FirstNonEmpty(ev.GetStringIgnoreErrors(utils.Tenant),
			smg.cgrCfg.GeneralCfg().DefaultTenant),
		engine.NewSafEvent(ev), clnt, "",
		smg.cgrCfg.SessionSCfg().DebitInterval); err != nil {
		if err != rpcclient.ErrSessionNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	if minMaxUsage == time.Duration(-1) {
		*maxUsage = -1.0
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return
}

// BiRPCV1UpdateSession updates an existing session, returning the duration which the session can still last
func (smg *SMGeneric) BiRPCV2UpdateSession(clnt rpcclient.RpcClientConnection,
	ev engine.MapEvent, maxUsage *time.Duration) (err error) {
	var minMaxUsage time.Duration
	if minMaxUsage, err = smg.UpdateSession(
		utils.FirstNonEmpty(ev.GetStringIgnoreErrors(utils.Tenant),
			smg.cgrCfg.GeneralCfg().DefaultTenant),
		engine.NewSafEvent(ev), clnt, "",
		smg.cgrCfg.SessionSCfg().DebitInterval); err != nil {
		if err != rpcclient.ErrSessionNotFound {
			err = utils.NewErrServerError(err)
		}
	} else {
		*maxUsage = minMaxUsage
	}
	return
}

// Called on session end, should stop debit loop
func (smg *SMGeneric) BiRPCV1TerminateSession(clnt rpcclient.RpcClientConnection,
	ev engine.MapEvent, reply *string) (err error) {
	if err = smg.TerminateSession(
		utils.FirstNonEmpty(ev.GetStringIgnoreErrors(utils.Tenant),
			smg.cgrCfg.GeneralCfg().DefaultTenant),
		engine.NewSafEvent(ev), clnt, "",
		smg.cgrCfg.SessionSCfg().DebitInterval); err != nil {
		if err != rpcclient.ErrSessionNotFound {
			err = utils.NewErrServerError(err)
		}
	} else {
		*reply = utils.OK
	}
	return
}

// Called on individual Events (eg SMS)
func (smg *SMGeneric) BiRPCV1ChargeEvent(clnt rpcclient.RpcClientConnection,
	ev engine.MapEvent, maxUsage *float64) error {
	if minMaxUsage, err := smg.ChargeEvent(
		utils.FirstNonEmpty(ev.GetStringIgnoreErrors(utils.Tenant),
			smg.cgrCfg.GeneralCfg().DefaultTenant),
		engine.NewSafEvent(ev)); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Called on individual Events (eg SMS)
func (smg *SMGeneric) BiRPCV2ChargeEvent(clnt rpcclient.RpcClientConnection,
	ev engine.MapEvent, maxUsage *time.Duration) error {
	if minMaxUsage, err := smg.ChargeEvent(
		utils.FirstNonEmpty(ev.GetStringIgnoreErrors(utils.Tenant),
			smg.cgrCfg.GeneralCfg().DefaultTenant),
		engine.NewSafEvent(ev)); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage
	}
	return nil
}

// Called on session end, should send the CDR to CDRS
func (smg *SMGeneric) BiRPCV1ProcessCDR(clnt rpcclient.RpcClientConnection,
	ev engine.MapEvent, reply *string) error {
	if err := smg.ProcessCDR(
		utils.FirstNonEmpty(ev.GetStringIgnoreErrors(utils.Tenant),
			smg.cgrCfg.GeneralCfg().DefaultTenant),
		engine.NewSafEvent(ev)); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

func (smg *SMGeneric) BiRPCV1GetActiveSessions(clnt rpcclient.RpcClientConnection,
	fltr map[string]string, reply *[]*ActiveSession) error {
	for fldName, fldVal := range fltr {
		if fldVal == "" {
			fltr[fldName] = utils.META_NONE
		}
	}
	aSessions, _, err := smg.asActiveSessions(fltr, false, false)
	if err != nil {
		return utils.NewErrServerError(err)
	} else if len(aSessions) == 0 {
		return utils.ErrNotFound
	}
	*reply = aSessions
	return nil
}

func (smg *SMGeneric) BiRPCV1GetActiveSessionsCount(clnt rpcclient.RpcClientConnection,
	fltr map[string]string, reply *int) error {
	for fldName, fldVal := range fltr {
		if fldVal == "" {
			fltr[fldName] = utils.META_NONE
		}
	}
	if _, count, err := smg.asActiveSessions(fltr, true, false); err != nil {
		return err
	} else {
		*reply = count
	}
	return nil
}

func (smg *SMGeneric) BiRPCV1GetPassiveSessions(clnt rpcclient.RpcClientConnection,
	fltr map[string]string, reply *[]*ActiveSession) error {
	for fldName, fldVal := range fltr {
		if fldVal == "" {
			fltr[fldName] = utils.META_NONE
		}
	}
	aSessions, _, err := smg.asActiveSessions(fltr, false, true)
	if err != nil {
		return utils.NewErrServerError(err)
	} else if len(aSessions) == 0 {
		return utils.ErrNotFound
	}
	*reply = aSessions
	return nil
}

func (smg *SMGeneric) BiRPCV1GetPassiveSessionsCount(clnt rpcclient.RpcClientConnection,
	fltr map[string]string, reply *int) error {
	for fldName, fldVal := range fltr {
		if fldVal == "" {
			fltr[fldName] = utils.META_NONE
		}
	}
	if _, count, err := smg.asActiveSessions(fltr, true, true); err != nil {
		return err
	} else {
		*reply = count
	}
	return nil
}

type ArgsSetPassiveSessions struct {
	CGRID    string
	Sessions []*SMGSession
}

// BiRPCV1SetPassiveSession used for replicating SMGSessions
func (smg *SMGeneric) BiRPCV1SetPassiveSessions(clnt rpcclient.RpcClientConnection,
	args ArgsSetPassiveSessions, reply *string) (err error) {
	if len(args.Sessions) == 0 {
		err = smg.removePassiveSessions(args.CGRID)
	} else {
		err = smg.setPassiveSessions(args.CGRID, args.Sessions)
	}
	if err == nil {
		*reply = utils.OK
	}
	return
}

type ArgsReplicateSessions struct {
	Filter      map[string]string
	Connections []*config.HaPoolConfig
}

// BiRPCV1ReplicateActiveSessions will replicate active sessions to either args.Connections or the internal configured ones
// args.Filter is used to filter the sessions which are replicated, CGRID is the only one possible for now
func (smg *SMGeneric) BiRPCV1ReplicateActiveSessions(clnt rpcclient.RpcClientConnection,
	args ArgsReplicateSessions, reply *string) (err error) {
	smgConns := smg.smgReplConns
	if len(args.Connections) != 0 {
		if smgConns, err = NewSessionReplicationConns(args.Connections,
			smg.cgrCfg.GeneralCfg().Reconnects, smg.cgrCfg.GeneralCfg().ConnectTimeout,
			smg.cgrCfg.GeneralCfg().ReplyTimeout); err != nil {
			return
		}
	}
	aSs := smg.getSessions(args.Filter[utils.CGRID], false)
	for cgrID := range aSs {
		smg.replicateSessionsWithID(cgrID, false, smgConns)
	}
	*reply = utils.OK
	return
}

// BiRPCV1ReplicatePassiveSessions will replicate passive sessions to either args.Connections or the internal configured ones
// args.Filter is used to filter the sessions which are replicated, CGRID is the only one possible for now
func (smg *SMGeneric) BiRPCV1ReplicatePassiveSessions(clnt rpcclient.RpcClientConnection,
	args ArgsReplicateSessions, reply *string) (err error) {
	smgConns := smg.smgReplConns
	if len(args.Connections) != 0 {
		if smgConns, err = NewSessionReplicationConns(args.Connections,
			smg.cgrCfg.GeneralCfg().Reconnects, smg.cgrCfg.GeneralCfg().ConnectTimeout,
			smg.cgrCfg.GeneralCfg().ReplyTimeout); err != nil {
			return
		}
	}
	aSs := smg.getSessions(args.Filter[utils.CGRID], true)
	for cgrID := range aSs {
		smg.replicateSessionsWithID(cgrID, true, smgConns)
	}
	*reply = utils.OK
	return
}

// NewV1AuthorizeArgs is a constructor for V1AuthorizeArgs
func NewV1AuthorizeArgs(attrs, res, maxUsage, thrslds,
	statQueues, suppls, supplsIgnoreErrs, supplsEventCost bool,
	cgrEv utils.CGREvent) (args *V1AuthorizeArgs) {
	args = &V1AuthorizeArgs{
		GetAttributes:         attrs,
		AuthorizeResources:    res,
		GetMaxUsage:           maxUsage,
		ProcessThresholds:     thrslds,
		ProcessStats:          statQueues,
		SuppliersIgnoreErrors: supplsIgnoreErrs,
		GetSuppliers:          suppls,
		CGREvent:              cgrEv,
	}
	if supplsEventCost {
		args.SuppliersMaxCost = utils.MetaSuppliersEventCost
	}
	return
}

type V1AuthorizeArgs struct {
	GetAttributes         bool
	AuthorizeResources    bool
	GetMaxUsage           bool
	ProcessThresholds     bool
	ProcessStats          bool
	GetSuppliers          bool
	SuppliersMaxCost      string
	SuppliersIgnoreErrors bool
	utils.CGREvent
	utils.Paginator
}

type V1AuthorizeReply struct {
	Attributes         *engine.AttrSProcessEventReply
	ResourceAllocation *string
	MaxUsage           *time.Duration
	Suppliers          *engine.SortedSuppliers
	ThresholdIDs       *[]string
	StatQueueIDs       *[]string
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1AuthReply *V1AuthorizeReply) AsNavigableMap(
	ignr []*config.CfgCdrField) (*config.NavigableMap, error) {
	cgrReply := make(map[string]interface{})
	if v1AuthReply != nil {
		if v1AuthReply.Attributes != nil {
			attrs := make(map[string]interface{})
			for _, fldName := range v1AuthReply.Attributes.AlteredFields {
				if v1AuthReply.Attributes.CGREvent.HasField(fldName) {
					attrs[fldName] = v1AuthReply.Attributes.CGREvent.Event[fldName]
				}
			}
			cgrReply[utils.CapAttributes] = attrs
		}
		if v1AuthReply.ResourceAllocation != nil {
			cgrReply[utils.CapResourceAllocation] = *v1AuthReply.ResourceAllocation
		}
		if v1AuthReply.MaxUsage != nil {
			cgrReply[utils.CapMaxUsage] = *v1AuthReply.MaxUsage
		}
		if v1AuthReply.Suppliers != nil {
			cgrReply[utils.CapSuppliers] = v1AuthReply.Suppliers.Digest()
		}
		if v1AuthReply.ThresholdIDs != nil {
			cgrReply[utils.CapThresholds] = *v1AuthReply.ThresholdIDs
		}
		if v1AuthReply.StatQueueIDs != nil {
			cgrReply[utils.CapStatQueues] = *v1AuthReply.StatQueueIDs
		}
	}
	return config.NewNavigableMap(cgrReply), nil
}

// BiRPCV1Authorize performs authorization for CGREvent based on specific components
func (smg *SMGeneric) BiRPCv1AuthorizeEvent(clnt rpcclient.RpcClientConnection,
	args *V1AuthorizeArgs, authReply *V1AuthorizeReply) (err error) {
	if !args.GetAttributes && !args.AuthorizeResources &&
		!args.GetMaxUsage && !args.GetSuppliers {
		return utils.NewErrMandatoryIeMissing("subsystems")
	}
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = smg.cgrCfg.GeneralCfg().DefaultTenant
	}
	if args.CGREvent.ID == "" {
		args.CGREvent.ID = utils.GenUUID()
	}
	if args.GetAttributes {
		if smg.attrS == nil {
			return utils.NewErrNotConnected(utils.AttributeS)
		}
		if args.CGREvent.Context == nil { // populate if not already in
			args.CGREvent.Context = utils.StringPointer(utils.MetaSessionS)
		}
		attrArgs := &engine.AttrArgsProcessEvent{
			CGREvent: args.CGREvent,
		}
		var rplyEv engine.AttrSProcessEventReply
		if err := smg.attrS.Call(utils.AttributeSv1ProcessEvent,
			attrArgs, &rplyEv); err == nil {
			args.CGREvent = *rplyEv.CGREvent
			authReply.Attributes = &rplyEv
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if args.GetMaxUsage {
		if smg.rals == nil {
			return utils.NewErrNotConnected(utils.RALService)
		}
		maxUsage, err := smg.GetMaxUsage(args.CGREvent.Tenant,
			engine.NewSafEvent(args.CGREvent.Event))
		if err != nil {
			return utils.NewErrRALs(err)
		}
		authReply.MaxUsage = &maxUsage
	}
	if args.AuthorizeResources {
		if smg.resS == nil {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		originID, _ := args.CGREvent.FieldAsString(utils.OriginID)
		if originID == "" {
			originID = utils.UUIDSha1Prefix()
		}
		var allocMsg string
		attrRU := utils.ArgRSv1ResourceUsage{
			CGREvent: args.CGREvent,
			UsageID:  originID,
			Units:    1,
		}
		if err = smg.resS.Call(utils.ResourceSv1AuthorizeResources,
			attrRU, &allocMsg); err != nil {
			return utils.NewErrResourceS(err)
		}
		authReply.ResourceAllocation = &allocMsg
	}
	if args.GetSuppliers {
		if smg.splS == nil {
			return utils.NewErrNotConnected(utils.SupplierS)
		}
		cgrEv := args.CGREvent.Clone()
		if acd, has := cgrEv.Event[utils.ACD]; has {
			cgrEv.Event[utils.Usage] = acd
		}
		var splsReply engine.SortedSuppliers
		sArgs := &engine.ArgsGetSuppliers{
			IgnoreErrors: args.SuppliersIgnoreErrors,
			MaxCost:      args.SuppliersMaxCost,
			CGREvent:     *cgrEv,
			Paginator:    args.Paginator,
		}
		if err = smg.splS.Call(utils.SupplierSv1GetSuppliers,
			sArgs, &splsReply); err != nil {
			return utils.NewErrSupplierS(err)
		}
		if splsReply.SortedSuppliers != nil {
			authReply.Suppliers = &splsReply
		}
	}
	if smg.thdS != nil && args.ProcessThresholds {
		if smg.thdS == nil {
			return utils.NewErrNotConnected(utils.ThresholdS)
		}
		var tIDs []string
		thEv := &engine.ArgsProcessEvent{
			CGREvent: args.CGREvent,
		}
		if err := smg.thdS.Call(utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<SessionS> error: %s processing event %+v with ThresholdS.", err.Error(), thEv))
		}
		authReply.ThresholdIDs = &tIDs
	}
	if smg.statS != nil && args.ProcessStats {
		if smg.statS == nil {
			return utils.NewErrNotConnected(utils.StatService)
		}
		var statReply []string
		if err := smg.statS.Call(utils.StatSv1ProcessEvent, &engine.StatsArgsProcessEvent{CGREvent: args.CGREvent}, &statReply); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<SessionS> error: %s processing event %+v with StatS.", err.Error(), args.CGREvent))
		}
		authReply.StatQueueIDs = &statReply
	}
	return nil
}

type V1AuthorizeReplyWithDigest struct {
	AttributesDigest   *string
	ResourceAllocation *string
	MaxUsage           *float64 // special treat returning time.Duration.Seconds()
	SuppliersDigest    *string
	Thresholds         *string
	StatQueues         *string
}

// BiRPCv1AuthorizeEventWithDigest performs authorization for CGREvent based on specific components
// returning one level fields instead of multiple ones returned by BiRPCv1AuthorizeEvent
func (smg *SMGeneric) BiRPCv1AuthorizeEventWithDigest(clnt rpcclient.RpcClientConnection,
	args *V1AuthorizeArgs, authReply *V1AuthorizeReplyWithDigest) (err error) {
	if !args.GetAttributes && !args.AuthorizeResources &&
		!args.GetMaxUsage && !args.GetSuppliers {
		return utils.NewErrMandatoryIeMissing("subsystems")
	}
	var initAuthRply V1AuthorizeReply
	if err = smg.BiRPCv1AuthorizeEvent(clnt, args, &initAuthRply); err != nil {
		return
	}
	if args.GetAttributes && initAuthRply.Attributes != nil {
		authReply.AttributesDigest = utils.StringPointer(initAuthRply.Attributes.Digest())
	}
	if args.AuthorizeResources {
		authReply.ResourceAllocation = initAuthRply.ResourceAllocation
	}
	if args.GetMaxUsage {
		authReply.MaxUsage = utils.Float64Pointer(-1.0)
		if *initAuthRply.MaxUsage != time.Duration(-1) {
			authReply.MaxUsage = utils.Float64Pointer(initAuthRply.MaxUsage.Seconds())
		}
	}
	if args.GetSuppliers {
		authReply.SuppliersDigest = utils.StringPointer(initAuthRply.Suppliers.Digest())
	}
	if args.ProcessThresholds {
		authReply.Thresholds = utils.StringPointer(
			strings.Join(*initAuthRply.ThresholdIDs, utils.FIELDS_SEP))
	}
	if args.ProcessStats {
		authReply.StatQueues = utils.StringPointer(
			strings.Join(*initAuthRply.StatQueueIDs, utils.FIELDS_SEP))
	}
	return nil
}

func NewV1InitSessionArgs(attrs, resrc, acnt, thrslds, stats bool,
	cgrEv utils.CGREvent) *V1InitSessionArgs {
	return &V1InitSessionArgs{
		GetAttributes:     attrs,
		AllocateResources: resrc,
		InitSession:       acnt,
		ProcessThresholds: thrslds,
		ProcessStats:      stats,
		CGREvent:          cgrEv,
	}
}

type V1InitSessionArgs struct {
	GetAttributes     bool
	AllocateResources bool
	InitSession       bool
	ProcessThresholds bool
	ProcessStats      bool
	utils.CGREvent
}

type V1InitSessionReply struct {
	Attributes         *engine.AttrSProcessEventReply
	ResourceAllocation *string
	MaxUsage           *time.Duration
	ThresholdIDs       *[]string
	StatQueueIDs       *[]string
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1InitSessionReply) AsNavigableMap(
	ignr []*config.CfgCdrField) (*config.NavigableMap, error) {
	cgrReply := make(map[string]interface{})
	if v1Rply != nil {
		if v1Rply.Attributes != nil {
			attrs := make(map[string]interface{})
			for _, fldName := range v1Rply.Attributes.AlteredFields {
				if v1Rply.Attributes.CGREvent.HasField(fldName) {
					attrs[fldName] = v1Rply.Attributes.CGREvent.Event[fldName]
				}
			}
			cgrReply[utils.CapAttributes] = attrs
		}
		if v1Rply.ResourceAllocation != nil {
			cgrReply[utils.CapResourceAllocation] = *v1Rply.ResourceAllocation
		}
		if v1Rply.MaxUsage != nil {
			cgrReply[utils.CapMaxUsage] = *v1Rply.MaxUsage
		}
		if v1Rply.ThresholdIDs != nil {
			cgrReply[utils.CapThresholds] = *v1Rply.ThresholdIDs
		}
		if v1Rply.StatQueueIDs != nil {
			cgrReply[utils.CapStatQueues] = *v1Rply.StatQueueIDs
		}
	}
	return config.NewNavigableMap(cgrReply), nil
}

// BiRPCV2InitiateSession initiates a new session, returns the maximum duration the session can last
func (smg *SMGeneric) BiRPCv1InitiateSession(clnt rpcclient.RpcClientConnection,
	args *V1InitSessionArgs, rply *V1InitSessionReply) (err error) {
	if !args.GetAttributes && !args.AllocateResources && !args.InitSession {
		return utils.NewErrMandatoryIeMissing("subsystems")
	}
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = smg.cgrCfg.GeneralCfg().DefaultTenant
	}
	if args.CGREvent.ID == "" {
		args.CGREvent.ID = utils.GenUUID()
	}
	if args.GetAttributes {
		if smg.attrS == nil {
			return utils.NewErrNotConnected(utils.AttributeS)
		}
		if args.CGREvent.Context == nil { // populate if not already in
			args.CGREvent.Context = utils.StringPointer(utils.MetaSessionS)
		}
		attrArgs := &engine.AttrArgsProcessEvent{
			CGREvent: args.CGREvent,
		}
		var rplyEv engine.AttrSProcessEventReply
		if err := smg.attrS.Call(utils.AttributeSv1ProcessEvent,
			attrArgs, &rplyEv); err == nil {
			args.CGREvent = *rplyEv.CGREvent
			rply.Attributes = &rplyEv
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if args.AllocateResources {
		if smg.resS == nil {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		originID, err := args.CGREvent.FieldAsString(utils.OriginID)
		if err != nil {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		attrRU := utils.ArgRSv1ResourceUsage{
			CGREvent: args.CGREvent,
			UsageID:  originID,
			Units:    1,
		}
		var allocMessage string
		if err = smg.resS.Call(utils.ResourceSv1AllocateResources,
			attrRU, &allocMessage); err != nil {
			return utils.NewErrResourceS(err)
		}
		rply.ResourceAllocation = &allocMessage
	}
	if args.InitSession {
		if smg.rals == nil {
			return utils.NewErrNotConnected(utils.RALService)
		}
		var err error
		originID := ""
		if args.AllocateResources {
			originID, err = args.CGREvent.FieldAsString(utils.OriginID)
			if err != nil {
				return utils.NewErrMandatoryIeMissing(utils.OriginID)
			}
		}
		ev := engine.NewSafEvent(args.CGREvent.Event)
		dbtItvl := smg.cgrCfg.SessionSCfg().DebitInterval
		if ev.HasField(utils.CGRDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
			if dbtItvl, err = ev.GetDuration(utils.CGRDebitInterval); err != nil {
				return utils.NewErrRALs(err)
			}
		}
		if maxUsage, err := smg.InitiateSession(
			args.CGREvent.Tenant,
			ev, clnt, originID, dbtItvl); err != nil {
			return utils.NewErrRALs(err)
		} else {
			rply.MaxUsage = &maxUsage
		}
	}
	if args.ProcessThresholds {
		if smg.thdS == nil {
			return utils.NewErrNotConnected(utils.ThresholdS)
		}
		var tIDs []string
		thEv := &engine.ArgsProcessEvent{
			CGREvent: args.CGREvent,
		}
		if err := smg.thdS.Call(utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<SessionS> error: %s processing event %+v with ThresholdS.", err.Error(), thEv))
		}
		rply.ThresholdIDs = &tIDs
	}
	if args.ProcessStats {
		if smg.statS == nil {
			return utils.NewErrNotConnected(utils.StatService)
		}
		var statReply []string
		if err := smg.statS.Call(utils.StatSv1ProcessEvent, &engine.StatsArgsProcessEvent{CGREvent: args.CGREvent}, &statReply); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<SessionS> error: %s processing event %+v with StatS.", err.Error(), args.CGREvent))
		}
		rply.StatQueueIDs = &statReply
	}
	return
}

type V1InitReplyWithDigest struct {
	AttributesDigest   *string
	ResourceAllocation *string
	MaxUsage           *float64
	Thresholds         *string
	StatQueues         *string
}

func (smg *SMGeneric) BiRPCv1InitiateSessionWithDigest(clnt rpcclient.RpcClientConnection,
	args *V1InitSessionArgs, initReply *V1InitReplyWithDigest) (err error) {
	if !args.GetAttributes && !args.AllocateResources &&
		!args.InitSession {
		return utils.NewErrMandatoryIeMissing("subsystems")
	}
	var initSessionRply V1InitSessionReply
	if err = smg.BiRPCv1InitiateSession(clnt, args, &initSessionRply); err != nil {
		return
	}

	if args.GetAttributes && initSessionRply.Attributes != nil {
		initReply.AttributesDigest = utils.StringPointer(initSessionRply.Attributes.Digest())
	}

	if args.AllocateResources {
		initReply.ResourceAllocation = initSessionRply.ResourceAllocation
	}

	if args.InitSession {
		initReply.MaxUsage = utils.Float64Pointer(-1.0)
		if *initSessionRply.MaxUsage != time.Duration(-1) {
			initReply.MaxUsage = utils.Float64Pointer(initSessionRply.MaxUsage.Seconds())
		}
	}

	if args.ProcessThresholds {
		initReply.Thresholds = utils.StringPointer(
			strings.Join(*initSessionRply.ThresholdIDs, utils.FIELDS_SEP))
	}
	if args.ProcessStats {
		initReply.StatQueues = utils.StringPointer(
			strings.Join(*initSessionRply.StatQueueIDs, utils.FIELDS_SEP))
	}
	return nil
}

func NewV1UpdateSessionArgs(attrs, acnts bool,
	cgrEv utils.CGREvent) *V1UpdateSessionArgs {
	return &V1UpdateSessionArgs{GetAttributes: attrs,
		UpdateSession: acnts, CGREvent: cgrEv}
}

type V1UpdateSessionArgs struct {
	GetAttributes bool
	UpdateSession bool
	utils.CGREvent
}

type V1UpdateSessionReply struct {
	Attributes *engine.AttrSProcessEventReply
	MaxUsage   *time.Duration
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1UpdateSessionReply) AsNavigableMap(
	ignr []*config.CfgCdrField) (*config.NavigableMap, error) {
	cgrReply := make(map[string]interface{})
	if v1Rply != nil {
		if v1Rply.Attributes != nil {
			attrs := make(map[string]interface{})
			for _, fldName := range v1Rply.Attributes.AlteredFields {
				if v1Rply.Attributes.CGREvent.HasField(fldName) {
					attrs[fldName] = v1Rply.Attributes.CGREvent.Event[fldName]
				}
			}
			cgrReply[utils.CapAttributes] = attrs
		}
		if v1Rply.MaxUsage != nil {
			cgrReply[utils.CapMaxUsage] = *v1Rply.MaxUsage
		}
	}
	return config.NewNavigableMap(cgrReply), nil
}

// BiRPCV1UpdateSession updates an existing session, returning the duration which the session can still last
func (smg *SMGeneric) BiRPCv1UpdateSession(clnt rpcclient.RpcClientConnection,
	args *V1UpdateSessionArgs, rply *V1UpdateSessionReply) (err error) {
	if !args.GetAttributes && !args.UpdateSession {
		return utils.NewErrMandatoryIeMissing("subsystems")
	}
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = smg.cgrCfg.GeneralCfg().DefaultTenant
	}
	if args.CGREvent.ID == "" {
		args.CGREvent.ID = utils.GenUUID()
	}
	if args.GetAttributes {
		if smg.attrS == nil {
			return utils.NewErrNotConnected(utils.AttributeS)
		}
		if args.CGREvent.Context == nil { // populate if not already in
			args.CGREvent.Context = utils.StringPointer(utils.MetaSessionS)
		}
		attrArgs := &engine.AttrArgsProcessEvent{
			CGREvent: args.CGREvent,
		}
		var rplyEv engine.AttrSProcessEventReply
		if err := smg.attrS.Call(utils.AttributeSv1ProcessEvent,
			attrArgs, &rplyEv); err == nil {
			args.CGREvent = *rplyEv.CGREvent
			rply.Attributes = &rplyEv
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if args.UpdateSession {
		if smg.rals == nil {
			return utils.NewErrNotConnected(utils.RALService)
		}
		originID, err := args.CGREvent.FieldAsString(utils.OriginID)
		if err != nil {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		ev := engine.NewSafEvent(args.CGREvent.Event)
		dbtItvl := smg.cgrCfg.SessionSCfg().DebitInterval
		if ev.HasField(utils.CGRDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
			if dbtItvl, err = ev.GetDuration(utils.CGRDebitInterval); err != nil {
				return utils.NewErrRALs(err)
			}
		}
		if maxUsage, err := smg.UpdateSession(args.CGREvent.Tenant,
			ev, clnt, originID, dbtItvl); err != nil {
			return utils.NewErrRALs(err)
		} else {
			rply.MaxUsage = &maxUsage
		}
	}
	return
}

func NewV1TerminateSessionArgs(acnts, resrc, thrds, stats bool,
	cgrEv utils.CGREvent) *V1TerminateSessionArgs {
	return &V1TerminateSessionArgs{
		TerminateSession:  acnts,
		ReleaseResources:  resrc,
		ProcessThresholds: thrds,
		ProcessStats:      stats,
		CGREvent:          cgrEv}
}

type V1TerminateSessionArgs struct {
	TerminateSession  bool
	ReleaseResources  bool
	ProcessThresholds bool
	ProcessStats      bool
	utils.CGREvent
}

// BiRPCV1TerminateSession will stop debit loops as well as release any used resources
func (smg *SMGeneric) BiRPCv1TerminateSession(clnt rpcclient.RpcClientConnection,
	args *V1TerminateSessionArgs, rply *string) (err error) {
	if !args.TerminateSession && !args.ReleaseResources {
		return utils.NewErrMandatoryIeMissing("subsystems")
	}
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = smg.cgrCfg.GeneralCfg().DefaultTenant
	}
	if args.CGREvent.ID == "" {
		args.CGREvent.ID = utils.GenUUID()
	}
	if args.TerminateSession {
		if smg.rals == nil {
			return utils.NewErrNotConnected(utils.RALService)
		}
		originID, err := args.CGREvent.FieldAsString(utils.OriginID)
		if err != nil {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		ev := engine.NewSafEvent(args.CGREvent.Event)
		dbtItvl := smg.cgrCfg.SessionSCfg().DebitInterval
		if ev.HasField(utils.CGRDebitInterval) { // dynamic DebitInterval via CGRDebitInterval
			if dbtItvl, err = ev.GetDuration(utils.CGRDebitInterval); err != nil {
				return utils.NewErrRALs(err)
			}
		}
		if err = smg.TerminateSession(args.CGREvent.Tenant,
			ev, clnt, originID, dbtItvl); err != nil {
			return utils.NewErrRALs(err)
		}
	}
	if args.ReleaseResources {
		if smg.resS == nil {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		originID, err := args.CGREvent.FieldAsString(utils.OriginID)
		if err != nil {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		var reply string
		argsRU := utils.ArgRSv1ResourceUsage{
			CGREvent: args.CGREvent,
			UsageID:  originID, // same ID should be accepted by first group since the previous resource should be expired
			Units:    1,
		}
		if err = smg.resS.Call(utils.ResourceSv1ReleaseResources,
			argsRU, &reply); err != nil {
			return utils.NewErrResourceS(err)
		}
	}
	if args.ProcessThresholds {
		if smg.thdS == nil {
			return utils.NewErrNotConnected(utils.ThresholdS)
		}
		var tIDs []string
		thEv := &engine.ArgsProcessEvent{
			CGREvent: args.CGREvent,
		}
		if err := smg.thdS.Call(utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<SessionS> error: %s processing event %+v with ThresholdS.", err.Error(), thEv))
		}
	}
	if args.ProcessStats {
		if smg.statS == nil {
			return utils.NewErrNotConnected(utils.StatS)
		}
		var statReply []string
		if err := smg.statS.Call(utils.StatSv1ProcessEvent, &engine.StatsArgsProcessEvent{CGREvent: args.CGREvent}, &statReply); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<SessionS> error: %s processing event %+v with StatS.", err.Error(), args.CGREvent))
		}
	}
	*rply = utils.OK
	return
}

// Called on session end, should send the CDR to CDRS
func (smg *SMGeneric) BiRPCv1ProcessCDR(clnt rpcclient.RpcClientConnection,
	cgrEv *utils.CGREvent, reply *string) error {
	cgrEv.Context = utils.StringPointer(utils.MetaSessionS)
	return smg.cdrsrv.Call(utils.CdrsV2ProcessCDR, cgrEv, reply)
}

func NewV1ProcessEventArgs(resrc, acnts, attrs, thds, stats bool,
	cgrEv utils.CGREvent) *V1ProcessEventArgs {
	return &V1ProcessEventArgs{
		AllocateResources: resrc,
		Debit:             acnts,
		GetAttributes:     attrs,
		ProcessThresholds: thds,
		ProcessStats:      stats,
		CGREvent:          cgrEv,
	}
}

type V1ProcessEventArgs struct {
	GetAttributes     bool
	AllocateResources bool
	Debit             bool
	ProcessThresholds bool
	ProcessStats      bool

	utils.CGREvent
}

type V1ProcessEventReply struct {
	MaxUsage           *time.Duration
	ResourceAllocation *string
	Attributes         *engine.AttrSProcessEventReply
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1ProcessEventReply) AsNavigableMap(
	ignr []*config.CfgCdrField) (*config.NavigableMap, error) {
	cgrReply := make(map[string]interface{})
	if v1Rply != nil {
		if v1Rply.MaxUsage != nil {
			cgrReply[utils.CapMaxUsage] = *v1Rply.MaxUsage
		}
		if v1Rply.ResourceAllocation != nil {
			cgrReply[utils.CapResourceAllocation] = *v1Rply.ResourceAllocation
		}
		if v1Rply.Attributes != nil {
			attrs := make(map[string]interface{})
			for _, fldName := range v1Rply.Attributes.AlteredFields {
				if v1Rply.Attributes.CGREvent.HasField(fldName) {
					attrs[fldName] = v1Rply.Attributes.CGREvent.Event[fldName]
				}
			}
			cgrReply[utils.CapAttributes] = attrs
		}
	}
	return config.NewNavigableMap(cgrReply), nil
}

// Called on session end, should send the CDR to CDRS
func (smg *SMGeneric) BiRPCv1ProcessEvent(clnt rpcclient.RpcClientConnection,
	args *V1ProcessEventArgs, rply *V1ProcessEventReply) (err error) {
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = smg.cgrCfg.GeneralCfg().DefaultTenant
	}
	if args.CGREvent.ID == "" {
		args.CGREvent.ID = utils.GenUUID()
	}
	if args.GetAttributes {
		if smg.attrS == nil {
			return utils.NewErrNotConnected(utils.AttributeS)
		}
		if args.CGREvent.Context == nil { // populate if not already in
			args.CGREvent.Context = utils.StringPointer(utils.MetaSessionS)
		}
		attrArgs := &engine.AttrArgsProcessEvent{
			CGREvent: args.CGREvent,
		}
		var rplyEv engine.AttrSProcessEventReply
		if err := smg.attrS.Call(utils.AttributeSv1ProcessEvent,
			attrArgs, &rplyEv); err == nil {
			args.CGREvent = *rplyEv.CGREvent
			rply.Attributes = &rplyEv
		} else if err.Error() != utils.ErrNotFound.Error() {
			return utils.NewErrAttributeS(err)
		}
	}
	if args.AllocateResources {
		if smg.resS == nil {
			return utils.NewErrNotConnected(utils.ResourceS)
		}
		originID, err := args.CGREvent.FieldAsString(utils.OriginID)
		if err != nil {
			return utils.NewErrMandatoryIeMissing(utils.OriginID)
		}
		attrRU := utils.ArgRSv1ResourceUsage{
			CGREvent: args.CGREvent,
			UsageID:  originID,
			Units:    1,
		}
		var allocMessage string
		if err = smg.resS.Call(utils.ResourceSv1AllocateResources,
			attrRU, &allocMessage); err != nil {
			return utils.NewErrResourceS(err)
		}
		rply.ResourceAllocation = &allocMessage
	}
	if args.Debit {
		if smg.rals == nil {
			return utils.NewErrNotConnected(utils.RALService)
		}
		if maxUsage, err := smg.ChargeEvent(args.CGREvent.Tenant,
			engine.NewSafEvent(args.CGREvent.Event)); err != nil {
			return utils.NewErrRALs(err)
		} else {
			rply.MaxUsage = &maxUsage
		}
	}
	if args.ProcessThresholds {
		if smg.thdS == nil {
			return utils.NewErrNotConnected(utils.ThresholdS)
		}
		var tIDs []string
		thEv := &engine.ArgsProcessEvent{
			CGREvent: args.CGREvent,
		}
		if err := smg.thdS.Call(utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<SessionS> error: %s processing event %+v with ThresholdS.", err.Error(), thEv))
		}
	}
	if args.ProcessStats {
		if smg.statS == nil {
			return utils.NewErrNotConnected(utils.StatS)
		}
		var statReply []string
		if err := smg.statS.Call(utils.StatSv1ProcessEvent, &engine.StatsArgsProcessEvent{CGREvent: args.CGREvent}, &statReply); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<SessionS> error: %s processing event %+v with StatS.", err.Error(), args.CGREvent))
		}
	}
	return nil
}

func (smg *SMGeneric) OnBiJSONConnect(c *rpc2.Client) {
	var s struct{}
	smg.biJsonConns[c] = s
}

func (smg *SMGeneric) OnBiJSONDisconnect(c *rpc2.Client) {
	delete(smg.biJsonConns, c)
}

func (smg *SMGeneric) syncSessions() {
	var rpcClnts []rpcclient.RpcClientConnection
	for _, conn := range smg.intBiJSONConns {
		rpcClnts = append(rpcClnts, conn)
	}
	for conn := range smg.biJsonConns {
		rpcClnts = append(rpcClnts, conn)
	}
	queriedCGRIDs := make(utils.StringMap)
	var err error
	for _, conn := range rpcClnts {
		var queriedSessionIDs []*SessionID
		if conn != nil {
			errChan := make(chan error)
			go func() {
				errChan <- conn.Call(utils.SessionSv1GetActiveSessionIDs,
					"", &queriedSessionIDs)
			}()
			select {
			case err = <-errChan:
				if err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error quering session ids : %+v", utils.SessionS, err))
					continue
				}
			case <-time.After(smg.cgrCfg.GeneralCfg().ReplyTimeout):
				utils.Logger.Warning(
					fmt.Sprintf("<%s> timeout quering session ids ", utils.SessionS))
				continue
			}
			for _, sessionID := range queriedSessionIDs {
				queriedCGRIDs[sessionID.CGRID()] = true
			}
		}
	}
	var toBeRemoved []string
	smg.aSessionsMux.RLock()
	for cgrid := range smg.activeSessions {
		if _, has := queriedCGRIDs[cgrid]; !has {
			toBeRemoved = append(toBeRemoved, cgrid)
		}
	}
	smg.aSessionsMux.RUnlock()
	for _, cgrID := range toBeRemoved {
		aSessions := smg.getSessions(cgrID, false)
		if len(aSessions[cgrID]) == 0 {
			continue
		}
		terminator := &smgSessionTerminator{
			ttl: time.Duration(0),
		}
		smg.ttlTerminate(aSessions[cgrID][0], terminator)
	}
}

func (smg *SMGeneric) BiRPCv1SyncSessions(clnt rpcclient.RpcClientConnection,
	ignParam string, reply *string) error {
	smg.syncSessions()
	*reply = utils.OK
	return nil
}

func (smg *SMGeneric) BiRPCV1ForceDisconnect(clnt rpcclient.RpcClientConnection,
	fltr map[string]string, reply *string) error {
	for fldName, fldVal := range fltr {
		if fldVal == "" {
			fltr[fldName] = utils.META_NONE
		}
	}
	aSessions, _, err := smg.asActiveSessions(fltr, false, false)
	if err != nil {
		return utils.NewErrServerError(err)
	} else if len(aSessions) == 0 {
		return utils.ErrNotFound
	}
	for _, aSession := range aSessions {
		sessions := smg.getSessions(aSession.CGRID, false)
		if len(sessions[aSession.CGRID]) == 0 {
			continue
		}
		terminator := &smgSessionTerminator{
			ttl: time.Duration(0),
		}
		smg.ttlTerminate(sessions[aSession.CGRID][0], terminator)
	}
	*reply = utils.OK
	return nil
}

func (smg *SMGeneric) BiRPCv1RegisterInternalBiJSONConn(clnt rpcclient.RpcClientConnection,
	ignParam string, reply *string) error {
	smg.intBiJSONConns = append(smg.intBiJSONConns, clnt)
	*reply = utils.OK
	return nil
}
