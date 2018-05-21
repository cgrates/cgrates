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
	"reflect"
	"strings"
	"sync"
	"time"

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
)

func NewSessionReplicationConns(conns []*config.HaPoolConfig, reconnects int,
	connTimeout, replyTimeout time.Duration) (smgConns []*SMGReplicationConn, err error) {
	smgConns = make([]*SMGReplicationConn, len(conns))
	for i, replConnCfg := range conns {
		if replCon, err := rpcclient.NewRpcClient("tcp", replConnCfg.Address, 0, reconnects,
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
	statS, splS, attrS, cdrsrv rpcclient.RpcClientConnection,
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
	return &SMGeneric{cgrCfg: cgrCfg,
		rals:               rals,
		resS:               resS,
		thdS:               thdS,
		statS:              statS,
		splS:               splS,
		attrS:              attrS,
		cdrsrv:             cdrsrv,
		smgReplConns:       smgReplConns,
		Timezone:           timezone,
		activeSessions:     make(map[string][]*SMGSession),
		ssIdxCfg:           ssIdxCfg,
		aSessionsIndex:     make(map[string]map[string]map[string]utils.StringMap),
		aSessionsRIndex:    make(map[string][]*riFieldNameVal),
		passiveSessions:    make(map[string][]*SMGSession),
		pSessionsIndex:     make(map[string]map[string]map[string]utils.StringMap),
		pSessionsRIndex:    make(map[string][]*riFieldNameVal),
		sessionTerminators: make(map[string]*smgSessionTerminator),
		responseCache:      utils.NewResponseCache(cgrCfg.ResponseCacheTTL)}
}

type SMGeneric struct {
	cgrCfg             *config.CGRConfig             // Separate from smCfg since there can be multiple
	rals               rpcclient.RpcClientConnection // RALs connections
	resS               rpcclient.RpcClientConnection // ResourceS connections
	thdS               rpcclient.RpcClientConnection // ThresholdS connections
	statS              rpcclient.RpcClientConnection // StatS connections
	splS               rpcclient.RpcClientConnection // SupplierS connections
	attrS              rpcclient.RpcClientConnection // AttributeS connections
	cdrsrv             rpcclient.RpcClientConnection // CDR server connections
	smgReplConns       []*SMGReplicationConn         // list of connections where we will replicate our session data
	Timezone           string
	activeSessions     map[string][]*SMGSession // group sessions per sessionId, multiple runs based on derived charging
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
	ttl := s.EventStart.GetSessionTTL(smg.cgrCfg.SessionSCfg().SessionTTL,
		smg.cgrCfg.SessionSCfg().SessionTTLMaxDelay)
	if ttl == 0 {
		return
	}
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
		ttlLastUsed: s.EventStart.GetSessionTTLLastUsed(),
		ttlUsage:    s.EventStart.GetSessionTTLUsage(),
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
	cdr := s.EventStart.AsCDR(smg.cgrCfg, smg.Timezone)
	cdr.Usage = s.TotalUsage
	var reply string
	smg.cdrsrv.Call("CdrsV1.ProcessCDR", cdr, &reply)
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
	s.mux.RLock()
	defer s.mux.RUnlock()
	for fieldName := range smg.ssIdxCfg {
		fieldVal, err := utils.ReflectFieldAsString(s.EventStart, fieldName, "")
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
func (smg *SMGeneric) getSessionIDsMatchingIndexes(fltrs map[string]string, passiveSessions bool) (utils.StringMap, map[string]string) {
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
	runID := fltrs[utils.MEDI_RUNID]
	checkNr := 0
	for fltrName, fltrVal := range fltrs {
		checkNr += 1
		if _, hasFldName := ssIndx[fltrName]; !hasFldName {
			continue
		}
		if _, hasFldVal := ssIndx[fltrName][fltrVal]; !hasFldVal {
			matchedIndexes[fltrName] = utils.META_NONE
			continue
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
		for cgrID := range ssIndx[fltrName][fltrVal] {
			if _, hasCGRID := matchingSessions[cgrID]; !hasCGRID {
				delete(matchingSessions, cgrID)
			}
		}
	}
	return matchingSessions.Clone(), matchedIndexes
}

// getSessionIDsForPrefix works with session relocation returning list of sessions with ID matching prefix for OriginID field
func (smg *SMGeneric) getSessionIDsForPrefix(prefix string, passiveSessions bool) (cgrIDs []string) {
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
				cgrIDs = append(cgrIDs, ssIndx[utils.OriginID][originID][utils.META_DEFAULT].Slice()...)
			}
		}
	}
	idxMux.RUnlock()
	return
}

// sessionStart will handle a new session, pass the connectionId so we can communicate on disconnect request
func (smg *SMGeneric) sessionStart(evStart SMGenericEvent,
	clntConn rpcclient.RpcClientConnection) (err error) {
	cgrID := evStart.GetCGRID(utils.META_DEFAULT)
	_, err = guardian.Guardian.Guard(func() (interface{}, error) { // Lock it on CGRID level
		if pSS := smg.passiveToActive(cgrID); len(pSS) != 0 {
			return nil, nil // ToDo: handle here also debits
		}
		var sessionRuns []*engine.SessionRun
		if err := smg.rals.Call("Responder.GetSessionRuns",
			evStart.AsCDR(smg.cgrCfg, smg.Timezone), &sessionRuns); err != nil {
			return nil, err
		} else if len(sessionRuns) == 0 {
			s := &SMGSession{CGRID: cgrID, EventStart: evStart,
				RunID: utils.META_NONE, Timezone: smg.Timezone,
				rals: smg.rals, cdrsrv: smg.cdrsrv,
				clntConn: clntConn}
			smg.recordASession(s)
			return nil, nil
		}
		stopDebitChan := make(chan struct{})
		for _, sessionRun := range sessionRuns {
			s := &SMGSession{CGRID: cgrID, EventStart: evStart,
				RunID: sessionRun.DerivedCharger.RunID, Timezone: smg.Timezone,
				rals: smg.rals, cdrsrv: smg.cdrsrv,
				CD: sessionRun.CallDescriptor, clntConn: clntConn,
				clientProto: smg.cgrCfg.SessionSCfg().ClientProtocol}
			smg.recordASession(s)
			//utils.Logger.Info(fmt.Sprintf("<%s> Starting session: %s, runId: %s",utils.SessionS, sessionId, s.runId))
			if smg.cgrCfg.SessionSCfg().DebitInterval != 0 {
				s.stopDebit = stopDebitChan
				go s.debitLoop(smg.cgrCfg.SessionSCfg().DebitInterval)
			}
		}
		return nil, nil
	}, smg.cgrCfg.LockingTimeout, cgrID)
	return
}

// sessionEnd will end a session from outside
func (smg *SMGeneric) sessionEnd(cgrID string, usage time.Duration) error {
	_, err := guardian.Guardian.Guard(func() (interface{}, error) { // Lock it on UUID level
		ss := smg.getSessions(cgrID, false)
		if len(ss) == 0 {
			if ss = smg.passiveToActive(cgrID); len(ss) == 0 {
				return nil, nil // ToDo: handle here also debits
			}
		}
		if !smg.unrecordASession(cgrID) { // Unreference it early so we avoid concurrency
			return nil, nil // Did not find the session so no need to close it anymore
		}
		for idx, s := range ss[cgrID] {
			if s.RunID == utils.META_NONE {
				continue
			}
			s.TotalUsage = usage // save final usage as totalUsage
			if idx == 0 && s.stopDebit != nil {
				close(s.stopDebit) // Stop automatic debits
			}
			aTime, err := s.EventStart.GetAnswerTime(utils.META_DEFAULT, smg.Timezone)
			if err != nil || aTime.IsZero() {
				utils.Logger.Err(fmt.Sprintf("<%s> Could not retrieve answer time for session: %s, runId: %s, aTime: %+v, error: %v",
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
		return nil, nil
	}, smg.cgrCfg.LockingTimeout, cgrID)
	return err
}

// sessionRelocate is used when an update will relocate an initial session (eg multiple data streams)
func (smg *SMGeneric) sessionRelocate(initialID, cgrID, newOriginID string) error {
	_, err := guardian.Guardian.Guard(func() (interface{}, error) { // Lock it on initialID level
		if utils.IsSliceMember([]string{initialID, cgrID, newOriginID}, "") { // Not allowed empty params here
			return nil, utils.ErrMandatoryIeMissing
		}
		ssNew := smg.getSessions(cgrID, false)
		if len(ssNew) != 0 { // Already relocated
			return nil, nil
		}
		if pSSNew := smg.getSessions(cgrID, true); len(pSSNew) != 0 { // passive sessions recorded, will be recovered so no need of relocation
			return nil, nil
		}
		ss := smg.getSessions(initialID, false)
		if len(ss) == 0 { // No need of relocation
			if ss = smg.passiveToActive(initialID); len(ss) == 0 {
				return nil, utils.ErrNotFound
			}
		}
		for i, s := range ss[initialID] {
			s.mux.Lock()
			s.CGRID = cgrID                            // Overwrite initial CGRID with new one
			s.EventStart[utils.OriginID] = newOriginID // Overwrite OriginID for session indexing
			s.mux.Unlock()
			smg.recordASession(s)
			if i == 0 {
				smg.unrecordASession(initialID)
			}
		}
		return nil, nil
	}, smg.cgrCfg.LockingTimeout, initialID)
	return err
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
		ss[0].mux.RLock() // lock session so we can clone it after releasing the map lock
	}
	ssMux.RUnlock()
	var ssCln []*SMGSession
	err = utils.Clone(ss, &ssCln)
	if len(ss) != 0 {
		ss[0].mux.RUnlock()
	}
	if err != nil {
		return
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
		if _, alreadyChecked := checkedFilters[fltrFldName]; alreadyChecked && fltrFldName != utils.MEDI_RUNID { // Optimize further checks, RunID should stay since it can create bugs
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
			sMp, err := remainingSessions[i].EventStart.AsMapStringString()
			if err != nil {
				return nil, 0, err
			}
			if _, hasRunID := sMp[utils.MEDI_RUNID]; !hasRunID {
				sMp[utils.MEDI_RUNID] = utils.META_DEFAULT
			}
			matchingAll := true
			for fltrFldName, fltrFldVal := range fltrs {
				if fldVal, hasIt := sMp[fltrFldName]; !hasIt || fltrFldVal != fldVal { // No Match
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

// MaxUsage calculates maximum usage allowed for given gevent
func (smg *SMGeneric) GetMaxUsage(gev SMGenericEvent) (maxUsage time.Duration, err error) {
	cacheKey := "MaxUsage" + gev.GetCGRID(utils.META_DEFAULT)
	if item, err := smg.responseCache.Get(cacheKey); err == nil && item != nil {
		return (item.Value.(time.Duration)), item.Err
	}
	defer smg.responseCache.Cache(cacheKey, &utils.ResponseCacheItem{Value: maxUsage, Err: err})
	storedCdr := gev.AsCDR(config.CgrConfig(), smg.Timezone)
	if _, has := gev[utils.Usage]; !has { // make sure we have a minimum duration configured
		storedCdr.Usage = smg.cgrCfg.SessionSCfg().MaxCallDuration
	}
	var maxDur float64
	if err = smg.rals.Call("Responder.GetDerivedMaxSessionTime", storedCdr, &maxDur); err != nil {
		return
	}
	maxUsage = time.Duration(maxDur)
	if maxUsage != time.Duration(-1) &&
		maxUsage < smg.cgrCfg.SessionSCfg().MinCallDuration {
		return 0, errors.New("UNAUTHORIZED_MIN_DURATION")
	}
	return
}

// Called on session start
func (smg *SMGeneric) InitiateSession(gev SMGenericEvent, clnt rpcclient.RpcClientConnection) (maxUsage time.Duration, err error) {
	cgrID := gev.GetCGRID(utils.META_DEFAULT)
	cacheKey := "InitiateSession" + cgrID
	if item, err := smg.responseCache.Get(cacheKey); err == nil && item != nil {
		return item.Value.(time.Duration), item.Err
	}
	defer smg.responseCache.Cache(cacheKey, &utils.ResponseCacheItem{Value: maxUsage, Err: err}) // schedule response caching
	smg.deletePassiveSessions(cgrID)
	if err = smg.sessionStart(gev, clnt); err != nil {
		smg.sessionEnd(cgrID, 0)
		return
	}
	if smg.cgrCfg.SessionSCfg().DebitInterval != 0 { // Session handled by debit loop
		maxUsage = time.Duration(-1)
		return
	}
	maxUsage, err = smg.UpdateSession(gev, clnt)
	if err != nil || maxUsage == 0 {
		smg.sessionEnd(cgrID, 0)
	}
	return
}

// Execute debits for usage/maxUsage
func (smg *SMGeneric) UpdateSession(gev SMGenericEvent, clnt rpcclient.RpcClientConnection) (maxUsage time.Duration, err error) {
	cgrID := gev.GetCGRID(utils.META_DEFAULT)
	cacheKey := "UpdateSession" + cgrID
	if item, err := smg.responseCache.Get(cacheKey); err == nil && item != nil {
		return item.Value.(time.Duration), item.Err
	}
	defer smg.responseCache.Cache(cacheKey, &utils.ResponseCacheItem{Value: maxUsage, Err: err})
	if smg.cgrCfg.SessionSCfg().DebitInterval != 0 { // Not possible to update a session with debit loop active
		err = errors.New("ACTIVE_DEBIT_LOOP")
		return
	}
	if gev.HasField(utils.InitialOriginID) {
		initialCGRID := gev.GetCGRID(utils.InitialOriginID)
		err = smg.sessionRelocate(initialCGRID, cgrID, gev.GetOriginID(utils.META_DEFAULT))
		if err == utils.ErrNotFound { // Session was already relocated, create a new  session with this update
			err = smg.sessionStart(gev, clnt)
		}
		if err != nil {
			return
		}
		smg.replicateSessionsWithID(initialCGRID, false, smg.smgReplConns)
	}
	smg.resetTerminatorTimer(cgrID,
		gev.GetSessionTTL(
			smg.cgrCfg.SessionSCfg().SessionTTL,
			smg.cgrCfg.SessionSCfg().SessionTTLMaxDelay),
		gev.GetSessionTTLLastUsed(), gev.GetSessionTTLUsage())
	var lastUsed *time.Duration
	var evLastUsed time.Duration
	if evLastUsed, err = gev.GetLastUsed(utils.META_DEFAULT); err == nil {
		lastUsed = &evLastUsed
	} else if err != utils.ErrNotFound {
		return
	}
	if maxUsage, err = gev.GetMaxUsage(utils.META_DEFAULT,
		smg.cgrCfg.SessionSCfg().MaxCallDuration); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrMandatoryIeMissing
		}
		return
	}
	aSessions := smg.getSessions(cgrID, false)
	if len(aSessions) == 0 {
		if aSessions = smg.passiveToActive(cgrID); len(aSessions) == 0 {
			utils.Logger.Err(fmt.Sprintf("<%s> SessionUpdate with no active sessions for event: <%s>", utils.SessionS, cgrID))
			err = rpcclient.ErrSessionNotFound
			return
		}
	}
	defer smg.replicateSessionsWithID(gev.GetCGRID(utils.META_DEFAULT), false, smg.smgReplConns)
	for _, s := range aSessions[cgrID] {
		if s.RunID == utils.META_NONE {
			maxUsage = time.Duration(-1)
			continue
		}
		var maxDur time.Duration
		if maxDur, err = s.debit(maxUsage, lastUsed); err != nil {
			return
		} else if maxDur < maxUsage {
			maxUsage = maxDur
		}
	}
	return
}

// Called on session end, should stop debit loop
func (smg *SMGeneric) TerminateSession(gev SMGenericEvent, clnt rpcclient.RpcClientConnection) (err error) {
	cgrID := gev.GetCGRID(utils.META_DEFAULT)
	cacheKey := "TerminateSession" + cgrID
	if item, err := smg.responseCache.Get(cacheKey); err == nil && item != nil {
		return item.Err
	}
	defer smg.responseCache.Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
	if gev.HasField(utils.InitialOriginID) {
		initialCGRID := gev.GetCGRID(utils.InitialOriginID)
		err = smg.sessionRelocate(initialCGRID, cgrID, gev.GetOriginID(utils.META_DEFAULT))
		if err == utils.ErrNotFound { // Session was already relocated, create a new  session with this update
			err = smg.sessionStart(gev, clnt)
		}
		if err != nil && err != utils.ErrMandatoryIeMissing {
			return
		}
		smg.replicateSessionsWithID(initialCGRID, false, smg.smgReplConns)
	}
	sessionIDs := []string{cgrID}
	if gev.HasField(utils.OriginIDPrefix) { // OriginIDPrefix is present, OriginID will not be anymore considered
		if sessionIDPrefix, errPrefix := gev.GetFieldAsString(utils.OriginIDPrefix); errPrefix == nil {
			if sessionIDs = smg.getSessionIDsForPrefix(sessionIDPrefix, false); len(sessionIDs) == 0 {
				sessionIDs = smg.getSessionIDsForPrefix(sessionIDPrefix, true)
				for _, sessionID := range sessionIDs { // activate sessions for prefix
					smg.passiveToActive(sessionID)
				}
			}
		}
	}
	usage, errUsage := gev.GetUsage(utils.META_DEFAULT)
	var lastUsed time.Duration
	if errUsage != nil {
		if errUsage != utils.ErrNotFound {
			err = errUsage
			return
		}
		lastUsed, err = gev.GetLastUsed(utils.META_DEFAULT)
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
				utils.Logger.Err(fmt.Sprintf("<%s> SessionTerminate with no active sessions for cgrID: <%s>", utils.SessionS, cgrID))
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
}

// Processes one time events (eg: SMS)
func (smg *SMGeneric) ChargeEvent(gev SMGenericEvent) (maxUsage time.Duration, err error) {
	cgrID := gev.GetCGRID(utils.META_DEFAULT)
	cacheKey := "ChargeEvent" + cgrID
	if item, err := smg.responseCache.Get(cacheKey); err == nil && item != nil {
		return item.Value.(time.Duration), item.Err
	}
	defer smg.responseCache.Cache(cacheKey, &utils.ResponseCacheItem{Value: maxUsage, Err: err})
	var sessionRuns []*engine.SessionRun
	if err = smg.rals.Call("Responder.GetSessionRuns", gev.AsCDR(smg.cgrCfg, smg.Timezone), &sessionRuns); err != nil {
		return
	} else if len(sessionRuns) == 0 {
		return
	}
	var maxDurInit bool // Avoid differences between default 0 and received 0
	for _, sR := range sessionRuns {
		cc := new(engine.CallCost)
		if err = smg.rals.Call("Responder.MaxDebit", sR.CallDescriptor, cc); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Could not Debit CD: %+v, RunID: %s, error: %s", utils.SessionS, sR.CallDescriptor, sR.DerivedCharger.RunID, err.Error()))
			break
		}
		sR.CallCosts = append(sR.CallCosts, cc) // Save it so we can revert on issues
		if ccDur := cc.GetDuration(); ccDur == 0 {
			err = utils.ErrInsufficientCredit
			break
		} else if !maxDurInit || ccDur < maxUsage {
			maxUsage = ccDur
		}
	}
	if err != nil { // Refund the ones already taken since we have error on one of the debits
		for _, sR := range sessionRuns {
			if len(sR.CallCosts) == 0 {
				continue
			}
			cc := sR.CallCosts[0]
			if len(sR.CallCosts) > 1 {
				for _, ccSR := range sR.CallCosts {
					cc.Merge(ccSR)
				}
			}
			// collect increments
			var refundIncrements engine.Increments
			cc.Timespans.Decompress()
			for _, ts := range cc.Timespans {
				refundIncrements = append(refundIncrements, ts.Increments...)
			}
			// refund cc
			if len(refundIncrements) > 0 {
				cd := cc.CreateCallDescriptor()
				cd.Increments = refundIncrements
				cd.CgrID = cgrID
				cd.RunID = sR.CallDescriptor.RunID
				cd.Increments.Compress()
				var response float64
				errRefund := smg.rals.Call("Responder.RefundIncrements", cd, &response)
				if errRefund != nil {
					return 0, errRefund
				}
			}
		}
		return
	}
	var withErrors bool
	for _, sR := range sessionRuns {
		if len(sR.CallCosts) == 0 {
			continue
		}
		cc := sR.CallCosts[0]
		if len(sR.CallCosts) > 1 {
			for _, ccSR := range sR.CallCosts[1:] {
				cc.Merge(ccSR)
			}
		}
		cc.Round()
		roundIncrements := cc.GetRoundIncrements()
		if len(roundIncrements) != 0 {
			cd := cc.CreateCallDescriptor()
			cd.Increments = roundIncrements
			var response float64
			if errRefund := smg.rals.Call("Responder.RefundRounding", cd, &response); errRefund != nil {
				utils.Logger.Err(fmt.Sprintf("<SM> ERROR failed to refund rounding: %v", errRefund))
			}
		}
		var reply string
		smCost := &engine.SMCost{
			CGRID:       cgrID,
			CostSource:  utils.MetaSessionS,
			RunID:       sR.DerivedCharger.RunID,
			OriginHost:  gev.GetOriginatorIP(utils.META_DEFAULT),
			OriginID:    gev.GetOriginID(utils.META_DEFAULT),
			CostDetails: engine.NewEventCostFromCallCost(cc, cgrID, sR.DerivedCharger.RunID),
		}
		if errStore := smg.cdrsrv.Call("CdrsV1.StoreSMCost", engine.AttrCDRSStoreSMCost{Cost: smCost,
			CheckDuplicate: true}, &reply); errStore != nil && !strings.HasSuffix(errStore.Error(), utils.ErrExists.Error()) {
			withErrors = true
			utils.Logger.Err(fmt.Sprintf("<%s> Could not save CC: %+v, RunID: %s error: %s", utils.SessionS, cc, sR.DerivedCharger.RunID, errStore.Error()))
		}
	}
	if withErrors {
		err = ErrPartiallyExecuted
		return
	}
	return
}

func (smg *SMGeneric) ProcessCDR(gev SMGenericEvent) (err error) {
	cgrID := gev.GetCGRID(utils.META_DEFAULT)
	cacheKey := "ProcessCDR" + cgrID
	if item, err := smg.responseCache.Get(cacheKey); err == nil && item != nil {
		return item.Err
	}
	defer smg.responseCache.Cache(cacheKey, &utils.ResponseCacheItem{Err: err})
	var reply string
	if err = smg.cdrsrv.Call("CdrsV1.ProcessCDR",
		gev.AsCDR(smg.cgrCfg, smg.Timezone), &reply); err != nil {
		return
	}
	return
}

func (smg *SMGeneric) Connect() error {
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
	ev SMGenericEvent, maxUsage *float64) error {
	maxUsageDur, err := smg.GetMaxUsage(ev)
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
	ev SMGenericEvent, maxUsage *time.Duration) error {
	maxUsageDur, err := smg.GetMaxUsage(ev)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*maxUsage = maxUsageDur
	return nil
}

// Called on session start, returns the maximum number of seconds the session can last
func (smg *SMGeneric) BiRPCV1InitiateSession(clnt rpcclient.RpcClientConnection,
	ev SMGenericEvent, maxUsage *float64) (err error) {
	var minMaxUsage time.Duration
	if minMaxUsage, err = smg.InitiateSession(ev, clnt); err != nil {
		if err != rpcclient.ErrSessionNotFound {
			err = utils.NewErrServerError(err)
		}
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return
}

// BiRPCV2InitiateSession initiates a new session, returns the maximum duration the session can last
func (smg *SMGeneric) BiRPCV2InitiateSession(clnt rpcclient.RpcClientConnection,
	ev SMGenericEvent, maxUsage *time.Duration) (err error) {
	var minMaxUsage time.Duration
	if minMaxUsage, err = smg.InitiateSession(ev, clnt); err != nil {
		if err != rpcclient.ErrSessionNotFound {
			err = utils.NewErrServerError(err)
		}
	} else {
		*maxUsage = minMaxUsage
	}
	return
}

// Interim updates, returns remaining duration from the RALs
func (smg *SMGeneric) BiRPCV1UpdateSession(clnt rpcclient.RpcClientConnection,
	ev SMGenericEvent, maxUsage *float64) (err error) {
	var minMaxUsage time.Duration
	if minMaxUsage, err = smg.UpdateSession(ev, clnt); err != nil {
		if err != rpcclient.ErrSessionNotFound {
			err = utils.NewErrServerError(err)
		}
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return
}

// BiRPCV1UpdateSession updates an existing session, returning the duration which the session can still last
func (smg *SMGeneric) BiRPCV2UpdateSession(clnt rpcclient.RpcClientConnection,
	ev SMGenericEvent, maxUsage *time.Duration) (err error) {
	var minMaxUsage time.Duration
	if minMaxUsage, err = smg.UpdateSession(ev, clnt); err != nil {
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
	ev SMGenericEvent, reply *string) (err error) {
	if err = smg.TerminateSession(ev, clnt); err != nil {
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
	ev SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := smg.ChargeEvent(ev); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Called on individual Events (eg SMS)
func (smg *SMGeneric) BiRPCV2ChargeEvent(clnt rpcclient.RpcClientConnection,
	ev SMGenericEvent, maxUsage *time.Duration) error {
	if minMaxUsage, err := smg.ChargeEvent(ev); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage
	}
	return nil
}

// Called on session end, should send the CDR to CDRS
func (smg *SMGeneric) BiRPCV1ProcessCDR(clnt rpcclient.RpcClientConnection,
	ev SMGenericEvent, reply *string) error {
	if err := smg.ProcessCDR(ev); err != nil {
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
		if smgConns, err = NewSessionReplicationConns(args.Connections, smg.cgrCfg.Reconnects,
			smg.cgrCfg.ConnectTimeout, smg.cgrCfg.ReplyTimeout); err != nil {
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
		if smgConns, err = NewSessionReplicationConns(args.Connections, smg.cgrCfg.Reconnects,
			smg.cgrCfg.ConnectTimeout, smg.cgrCfg.ReplyTimeout); err != nil {
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

type V1AuthorizeArgs struct {
	GetAttributes         bool
	AuthorizeResources    bool
	GetMaxUsage           bool
	GetSuppliers          bool
	SuppliersMaxCost      string
	SuppliersIgnoreErrors bool
	ProcessThresholds     *bool
	ProcessStatQueues     *bool
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

// AsCGRReply is part of utils.CGRReplier interface
func (v1AuthReply *V1AuthorizeReply) AsCGRReply() (cgrReply utils.CGRReply, err error) {
	cgrReply = make(map[string]interface{})
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
	cgrReply[utils.Error] = "" // so we can compare in filters
	return
}

// BiRPCV1Authorize performs authorization for CGREvent based on specific components
func (smg *SMGeneric) BiRPCv1AuthorizeEvent(clnt rpcclient.RpcClientConnection,
	args *V1AuthorizeArgs, authReply *V1AuthorizeReply) (err error) {
	if !args.GetAttributes && !args.AuthorizeResources &&
		!args.GetMaxUsage && !args.GetSuppliers {
		return utils.NewErrMandatoryIeMissing("subsystems")
	}
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = smg.cgrCfg.DefaultTenant
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
		var rplyEv engine.AttrSProcessEventReply
		if err := smg.attrS.Call(utils.AttributeSv1ProcessEvent,
			&args.CGREvent, &rplyEv); err == nil {
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
		maxUsage, err := smg.GetMaxUsage(args.CGREvent.Event)
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
	checkThresholds := smg.thdS != nil
	if args.ProcessThresholds != nil {
		checkThresholds = *args.ProcessThresholds
	}
	if checkThresholds {
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
	checkStatQueues := smg.statS != nil
	if args.ProcessStatQueues != nil {
		checkStatQueues = *args.ProcessStatQueues
	}
	if checkStatQueues {
		if smg.statS == nil {
			return utils.NewErrNotConnected(utils.StatService)
		}
		var statReply []string
		if err := smg.statS.Call(utils.StatSv1ProcessEvent, &args.CGREvent, &statReply); err != nil &&
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
	if args.ProcessThresholds != nil && *args.ProcessThresholds {
		authReply.Thresholds = utils.StringPointer(
			strings.Join(*initAuthRply.ThresholdIDs, utils.FIELDS_SEP))
	}
	if args.ProcessStatQueues != nil && *args.ProcessStatQueues {
		authReply.StatQueues = utils.StringPointer(
			strings.Join(*initAuthRply.StatQueueIDs, utils.FIELDS_SEP))
	}
	return nil
}

type V1InitSessionArgs struct {
	GetAttributes     bool
	AllocateResources bool
	InitSession       bool
	ProcessThresholds *bool
	ProcessStatQueues *bool
	utils.CGREvent
}

type V1InitSessionReply struct {
	Attributes         *engine.AttrSProcessEventReply
	ResourceAllocation *string
	MaxUsage           *time.Duration
	ThresholdIDs       *[]string
	StatQueueIDs       *[]string
}

// AsCGRReply is part of utils.CGRReplier interface
func (v1Rply *V1InitSessionReply) AsCGRReply() (cgrReply utils.CGRReply, err error) {
	cgrReply = make(map[string]interface{})
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
	cgrReply[utils.Error] = ""
	return
}

// BiRPCV2InitiateSession initiates a new session, returns the maximum duration the session can last
func (smg *SMGeneric) BiRPCv1InitiateSession(clnt rpcclient.RpcClientConnection,
	args *V1InitSessionArgs, rply *V1InitSessionReply) (err error) {
	if !args.GetAttributes && !args.AllocateResources && !args.InitSession {
		return utils.NewErrMandatoryIeMissing("subsystems")
	}
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = smg.cgrCfg.DefaultTenant
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
		var rplyEv engine.AttrSProcessEventReply
		if err := smg.attrS.Call(utils.AttributeSv1ProcessEvent,
			&args.CGREvent, &rplyEv); err == nil {
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
		if maxUsage, err := smg.InitiateSession(args.CGREvent.Event, clnt); err != nil {
			return utils.NewErrRALs(err)
		} else {
			rply.MaxUsage = &maxUsage
		}
	}
	checkThresholds := smg.thdS != nil
	if args.ProcessThresholds != nil {
		checkThresholds = *args.ProcessThresholds
	}
	if checkThresholds {
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
	checkStatQueues := smg.statS != nil
	if args.ProcessStatQueues != nil {
		checkStatQueues = *args.ProcessStatQueues
	}
	if checkStatQueues {
		if smg.statS == nil {
			return utils.NewErrNotConnected(utils.StatService)
		}
		var statReply []string
		if err := smg.statS.Call(utils.StatSv1ProcessEvent, &args.CGREvent, &statReply); err != nil &&
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

	if args.ProcessThresholds != nil && *args.ProcessThresholds {
		initReply.Thresholds = utils.StringPointer(
			strings.Join(*initSessionRply.ThresholdIDs, utils.FIELDS_SEP))
	}
	if args.ProcessStatQueues != nil && *args.ProcessStatQueues {
		initReply.StatQueues = utils.StringPointer(
			strings.Join(*initSessionRply.StatQueueIDs, utils.FIELDS_SEP))
	}
	return nil
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

// AsCGRReply is part of utils.CGRReplier interface
func (v1Rply *V1UpdateSessionReply) AsCGRReply() (cgrReply utils.CGRReply, err error) {
	cgrReply = make(map[string]interface{})
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
	cgrReply[utils.Error] = ""
	return
}

// BiRPCV1UpdateSession updates an existing session, returning the duration which the session can still last
func (smg *SMGeneric) BiRPCv1UpdateSession(clnt rpcclient.RpcClientConnection,
	args *V1UpdateSessionArgs, rply *V1UpdateSessionReply) (err error) {
	if !args.GetAttributes && !args.UpdateSession {
		return utils.NewErrMandatoryIeMissing("subsystems")
	}
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = smg.cgrCfg.DefaultTenant
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
		var rplyEv engine.AttrSProcessEventReply
		if err := smg.attrS.Call(utils.AttributeSv1ProcessEvent,
			&args.CGREvent, &rplyEv); err == nil {
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
		if maxUsage, err := smg.UpdateSession(args.CGREvent.Event, clnt); err != nil {
			return utils.NewErrRALs(err)
		} else {
			rply.MaxUsage = &maxUsage
		}
	}
	return
}

type V1TerminateSessionArgs struct {
	TerminateSession  bool
	ReleaseResources  bool
	ProcessThresholds *bool
	ProcessStatQueues *bool
	utils.CGREvent
}

// BiRPCV1TerminateSession will stop debit loops as well as release any used resources
func (smg *SMGeneric) BiRPCv1TerminateSession(clnt rpcclient.RpcClientConnection,
	args *V1TerminateSessionArgs, rply *string) (err error) {
	if !args.TerminateSession && !args.ReleaseResources {
		return utils.NewErrMandatoryIeMissing("subsystems")
	}
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = smg.cgrCfg.DefaultTenant
	}
	if args.CGREvent.ID == "" {
		args.CGREvent.ID = utils.GenUUID()
	}
	if args.TerminateSession {
		if smg.rals == nil {
			return utils.NewErrNotConnected(utils.RALService)
		}
		if err = smg.TerminateSession(args.CGREvent.Event, clnt); err != nil {
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
	checkThresholds := smg.thdS != nil
	if args.ProcessThresholds != nil {
		checkThresholds = *args.ProcessThresholds
	}
	if checkThresholds {
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
	checkStatQueues := smg.statS != nil
	if args.ProcessStatQueues != nil {
		checkStatQueues = *args.ProcessStatQueues
	}
	if checkStatQueues {
		if smg.statS == nil {
			return utils.NewErrNotConnected(utils.StatService)
		}
		var statReply []string
		if err := smg.statS.Call(utils.StatSv1ProcessEvent, &args.CGREvent, &statReply); err != nil &&
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
	cgrEv utils.CGREvent, reply *string) error {
	if err := smg.ProcessCDR(cgrEv.Event); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type V1ProcessEventArgs struct {
	AllocateResources bool
	Debit             bool
	GetAttributes     bool
	utils.CGREvent
}

type V1ProcessEventReply struct {
	MaxUsage           *time.Duration
	ResourceAllocation *string
	Attributes         *engine.AttrSProcessEventReply
}

// AsCGRReply is part of utils.CGRReplier interface
func (v1Rply *V1ProcessEventReply) AsCGRReply() (cgrReply utils.CGRReply, err error) {
	cgrReply = make(map[string]interface{})
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
	cgrReply[utils.Error] = ""
	return
}

// Called on session end, should send the CDR to CDRS
func (smg *SMGeneric) BiRPCv1ProcessEvent(clnt rpcclient.RpcClientConnection,
	args *V1ProcessEventArgs, rply *V1ProcessEventReply) (err error) {
	if args.CGREvent.Tenant == "" {
		args.CGREvent.Tenant = smg.cgrCfg.DefaultTenant
	}
	if args.CGREvent.ID == "" {
		args.CGREvent.ID = utils.GenUUID()
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
		if maxUsage, err := smg.ChargeEvent(args.CGREvent.Event); err != nil {
			return utils.NewErrRALs(err)
		} else {
			rply.MaxUsage = &maxUsage
		}
	}
	if args.GetAttributes {
		if smg.attrS == nil {
			return utils.NewErrNotConnected(utils.AttributeS)
		}
		if args.CGREvent.Context == nil {
			args.CGREvent.Context = utils.StringPointer(utils.MetaSessionS)
		}
		var rplyEv engine.AttrSProcessEventReply
		if err = smg.attrS.Call(utils.AttributeSv1ProcessEvent,
			args.CGREvent, &rplyEv); err != nil {
			return utils.NewErrAttributeS(err)
		}
		rply.Attributes = &rplyEv
	}
	return nil
}
