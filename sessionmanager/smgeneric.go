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
package sessionmanager

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

const (
	MaxTimespansInCost = 50
)

var ErrPartiallyExecuted = errors.New("Partially executed")

// ReplicationConnection represents one connection to a passive node where we will replicate session data
type SMGReplicationConnection struct {
	conn rpcclient.RpcClientConnection
	sync bool
}

func NewSMGeneric(cgrCfg *config.CGRConfig, rater rpcclient.RpcClientConnection, cdrsrv rpcclient.RpcClientConnection, timezone string) *SMGeneric {
	return &SMGeneric{cgrCfg: cgrCfg,
		rater:              rater,
		cdrsrv:             cdrsrv,
		timezone:           timezone,
		activeSessions:     make(map[string][]*SMGSession),
		aSessionsIndex:     make(map[string]map[string]utils.StringMap),
		passiveSessions:    make(map[string][]*SMGSession),
		sessionTerminators: make(map[string]*smgSessionTerminator)}
}

type SMGeneric struct {
	cgrCfg             *config.CGRConfig // Separate from smCfg since there can be multiple
	rater              rpcclient.RpcClientConnection
	cdrsrv             rpcclient.RpcClientConnection
	smgReplConns       []*SMGReplicationConnection // list of connections where we will replicate our session data
	timezone           string
	activeSessions     map[string][]*SMGSession // group sessions per sessionId, multiple runs based on derived charging
	aSessionsMux       sync.RWMutex
	aSessionsIndex     map[string]map[string]utils.StringMap // map[fieldName]map[fieldValue]utils.StringMap[sesionID]
	aSIMux             sync.RWMutex                          // protects aSessionsIndex
	passiveSessions    map[string][]*SMGSession              // group passive sessions
	pSessionsMux       sync.RWMutex
	sessionTerminators map[string]*smgSessionTerminator // terminate and cleanup the session if timer expires

}
type smgSessionTerminator struct {
	timer       *time.Timer
	endChan     chan bool
	ttl         time.Duration
	ttlLastUsed *time.Duration
	ttlUsage    *time.Duration
}

// Updates the timer for the session to a new ttl and terminate info
func (smg *SMGeneric) resetTerminatorTimer(uuid string, ttl time.Duration, ttlLastUsed, ttlUsage *time.Duration) {
	smg.aSessionsMux.RLock()
	if st, found := smg.sessionTerminators[uuid]; found {
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
	smg.aSessionsMux.RUnlock()
}

// Called when a session timeouts
func (smg *SMGeneric) ttlTerminate(s *SMGSession, tmtr *smgSessionTerminator) {
	debitUsage := tmtr.ttl
	if tmtr.ttlUsage != nil {
		debitUsage = *tmtr.ttlUsage
	}
	for _, s := range smg.getASession(s.EventStart.GetUUID()) {
		s.debit(debitUsage, tmtr.ttlLastUsed)
	}
	smg.sessionEnd(s.EventStart.GetUUID(), s.TotalUsage)
	cdr := s.EventStart.AsStoredCdr(smg.cgrCfg, smg.timezone)
	cdr.Usage = s.TotalUsage
	var reply string
	smg.cdrsrv.Call("CdrsV1.ProcessCDR", cdr, &reply)
}

func (smg *SMGeneric) recordASession(uuid string, s *SMGSession) {
	smg.aSessionsMux.Lock()
	smg.activeSessions[uuid] = append(smg.activeSessions[uuid], s)
	if smg.cgrCfg.SmGenericConfig.SessionTTL != 0 {
		if _, found := smg.sessionTerminators[uuid]; !found {
			ttl := smg.cgrCfg.SmGenericConfig.SessionTTL
			if ttlEv := s.EventStart.GetSessionTTL(); ttlEv != 0 {
				ttl = ttlEv
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
			smg.sessionTerminators[uuid] = terminator
			go func() {
				select {
				case <-timer.C:
					smg.ttlTerminate(s, terminator)
				case <-endChan:
					timer.Stop()
				}
			}()

		}
	}
	smg.indexASession(uuid, s)
	smg.aSessionsMux.Unlock()
}

// Remove session from session list, removes all related in case of multiple runs, true if item was found
func (smg *SMGeneric) unrecordASession(uuid string) bool {
	smg.aSessionsMux.Lock()
	defer smg.aSessionsMux.Unlock()
	if _, found := smg.activeSessions[uuid]; !found {
		return false
	}
	delete(smg.activeSessions, uuid)
	if st, found := smg.sessionTerminators[uuid]; found {
		st.endChan <- true
		delete(smg.sessionTerminators, uuid)
	}
	smg.unindexASession(uuid)
	return true
}

// indexASession explores settings and builds smg.aSessionsIndex based on that
func (smg *SMGeneric) indexASession(uuid string, s *SMGSession) bool {
	smg.aSIMux.Lock()
	defer smg.aSIMux.Unlock()
	ev := s.EventStart
	for _, fieldName := range smg.cgrCfg.SmGenericConfig.SessionIndexes {
		fieldVal, err := utils.ReflectFieldAsString(ev, fieldName, "")
		if err != nil {
			if err == utils.ErrNotFound {
				fieldVal = utils.NOT_AVAILABLE
			} else {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Error retrieving field: %s from event: %+v", fieldName, ev))
				continue
			}
		}
		if fieldVal == "" {
			fieldVal = utils.MetaEmpty
		}
		if _, hasFieldName := smg.aSessionsIndex[fieldName]; !hasFieldName { // Init it here so we can minimize
			smg.aSessionsIndex[fieldName] = make(map[string]utils.StringMap)
		}
		if _, hasFieldVal := smg.aSessionsIndex[fieldName][fieldVal]; !hasFieldVal {
			smg.aSessionsIndex[fieldName][fieldVal] = make(utils.StringMap)
		}
		smg.aSessionsIndex[fieldName][fieldVal][uuid] = true
	}
	return true
}

// unindexASession removes a session from indexes
func (smg *SMGeneric) unindexASession(uuid string) bool {
	smg.aSIMux.Lock()
	defer smg.aSIMux.Unlock()
	var found bool
	for fldName := range smg.aSessionsIndex {
		for fldVal := range smg.aSessionsIndex[fldName] {
			if _, hasUUID := smg.aSessionsIndex[fldName][fldVal][uuid]; hasUUID {
				found = true
				delete(smg.aSessionsIndex[fldName][fldVal], uuid)
				if len(smg.aSessionsIndex[fldName][fldVal]) == 0 {
					delete(smg.aSessionsIndex[fldName], fldVal)
				}
				if len(smg.aSessionsIndex[fldName]) == 0 {
					delete(smg.aSessionsIndex, fldName)
				}
			}
		}
	}
	return found
}

// getSessionIDsMatchingIndexes will check inside indexes if it can find sessionIDs matching all filters
// matchedIndexes returns map[matchedFieldName]possibleMatchedFieldVal so we optimize further to avoid checking them
func (smg *SMGeneric) getSessionIDsMatchingIndexes(fltrs map[string]string) (utils.StringMap, map[string]string) {
	smg.aSIMux.RLock()
	defer smg.aSIMux.RUnlock()
	sessionIDxes := smg.aSessionsIndex // Clone here and unlock sooner if getting slow
	matchedIndexes := make(map[string]string)
	var matchingSessions utils.StringMap
	checkNr := 0
	for fltrName, fltrVal := range fltrs {
		checkNr += 1
		if _, hasFldName := sessionIDxes[fltrName]; !hasFldName {
			continue
		}
		if _, hasFldVal := sessionIDxes[fltrName][fltrVal]; !hasFldVal {
			matchedIndexes[fltrName] = utils.META_NONE
			continue
		}
		matchedIndexes[fltrName] = fltrVal
		if checkNr == 1 { // First run will init the MatchingSessions
			matchingSessions = sessionIDxes[fltrName][fltrVal]
			continue
		}
		// Higher run, takes out non matching indexes
		for sessID := range sessionIDxes[fltrName][fltrVal] {
			if _, hasUUID := matchingSessions[sessID]; !hasUUID {
				delete(matchingSessions, sessID)
			}
		}
	}
	return matchingSessions.Clone(), matchedIndexes
}

func (smg *SMGeneric) getSessionIDsForPrefix(prefix string) []string {
	smg.aSessionsMux.Lock()
	defer smg.aSessionsMux.Unlock()
	sessionIDs := make([]string, 0)
	for sessionID := range smg.activeSessions {
		if strings.HasPrefix(sessionID, prefix) {
			sessionIDs = append(sessionIDs, sessionID)
		}
	}
	return sessionIDs
}

// Returns sessions/derived for a specific uuid
func (smg *SMGeneric) getASession(uuid string) []*SMGSession {
	smg.aSessionsMux.RLock()
	defer smg.aSessionsMux.RUnlock()
	return smg.activeSessions[uuid]
}

// Handle a new session, pass the connectionId so we can communicate on disconnect request
func (smg *SMGeneric) sessionStart(evStart SMGenericEvent, clntConn rpcclient.RpcClientConnection) error {
	sessionId := evStart.GetUUID()
	processed, err := engine.Guardian.Guard(func() (interface{}, error) { // Lock it on UUID level
		var sessionRuns []*engine.SessionRun
		if err := smg.rater.Call("Responder.GetSessionRuns", evStart.AsStoredCdr(smg.cgrCfg, smg.timezone), &sessionRuns); err != nil {
			return true, err
		} else if len(sessionRuns) == 0 {
			return true, nil
		}
		stopDebitChan := make(chan struct{})
		for _, sessionRun := range sessionRuns {
			s := &SMGSession{EventStart: evStart, RunID: sessionRun.DerivedCharger.RunID, Timezone: smg.timezone,
				rater: smg.rater, cdrsrv: smg.cdrsrv, CD: sessionRun.CallDescriptor, clntConn: clntConn}
			smg.recordASession(sessionId, s)
			//utils.Logger.Info(fmt.Sprintf("<SMGeneric> Starting session: %s, runId: %s", sessionId, s.runId))
			if smg.cgrCfg.SmGenericConfig.DebitInterval != 0 {
				s.stopDebit = stopDebitChan
				go s.debitLoop(smg.cgrCfg.SmGenericConfig.DebitInterval)
			}
		}
		return true, nil
	}, smg.cgrCfg.LockingTimeout, sessionId)
	if processed == nil || processed == false {
		utils.Logger.Err("<SMGeneric> Cannot start session, empty reply")
		return utils.ErrServerError
	}
	return err
}

// End a session from outside
func (smg *SMGeneric) sessionEnd(sessionId string, usage time.Duration) error {
	_, err := engine.Guardian.Guard(func() (interface{}, error) { // Lock it on UUID level
		ss := smg.getASession(sessionId)
		if len(ss) == 0 { // Not handled by us
			return nil, nil
		}
		if !smg.unrecordASession(sessionId) { // Unreference it early so we avoid concurrency
			return nil, nil // Did not find the session so no need to close it anymore
		}
		for idx, s := range ss {
			s.TotalUsage = usage // save final usage as totalUsage
			if idx == 0 && s.stopDebit != nil {
				close(s.stopDebit) // Stop automatic debits
			}
			aTime, err := s.EventStart.GetAnswerTime(utils.META_DEFAULT, smg.cgrCfg.DefaultTimezone)
			if err != nil || aTime.IsZero() {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not retrieve answer time for session: %s, runId: %s, aTime: %+v, error: %v",
					sessionId, s.RunID, aTime, err))
				continue // Unanswered session
			}
			if err := s.close(aTime.Add(usage)); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not close session: %s, runId: %s, error: %s", sessionId, s.RunID, err.Error()))
			}
			if err := s.saveOperations(sessionId); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not save session: %s, runId: %s, error: %s", sessionId, s.RunID, err.Error()))
			}
		}
		return nil, nil
	}, time.Duration(2)*time.Second, sessionId)
	return err
}

// Used when an update will relocate an initial session (eg multiple data streams)
func (smg *SMGeneric) sessionRelocate(sessionID, initialID string) error {
	_, err := engine.Guardian.Guard(func() (interface{}, error) { // Lock it on initialID level
		if utils.IsSliceMember([]string{sessionID, initialID}, "") { // Not allowed empty params here
			return nil, utils.ErrMandatoryIeMissing
		}
		ssNew := smg.getASession(sessionID) // Already relocated
		if len(ssNew) != 0 {
			return nil, nil
		}
		ss := smg.getASession(initialID)
		if len(ss) == 0 { // No need of relocation
			return nil, utils.ErrNotFound
		}
		for i, s := range ss {
			s.EventStart[utils.ACCID] = sessionID // Overwrite initialSessionID with new one
			smg.recordASession(sessionID, s)
			if i == 0 {
				smg.unrecordASession(initialID)
			}
		}
		return nil, nil
	}, time.Duration(2)*time.Second, initialID)
	return err
}

// Methods to apply on sessions, mostly exported through RPC/Bi-RPC
//Calculates maximum usage allowed for gevent
func (smg *SMGeneric) MaxUsage(gev SMGenericEvent) (time.Duration, error) {
	gev[utils.EVENT_NAME] = utils.CGR_AUTHORIZATION
	storedCdr := gev.AsStoredCdr(config.CgrConfig(), smg.timezone)
	var maxDur float64
	if err := smg.rater.Call("Responder.GetDerivedMaxSessionTime", storedCdr, &maxDur); err != nil {
		return time.Duration(0), err
	}
	return time.Duration(maxDur), nil
}

func (smg *SMGeneric) LCRSuppliers(gev SMGenericEvent) ([]string, error) {
	gev[utils.EVENT_NAME] = utils.CGR_LCR_REQUEST
	cd, err := gev.AsLcrRequest().AsCallDescriptor(smg.timezone)
	cd.CgrID = gev.GetCgrId(smg.timezone)
	if err != nil {
		return nil, err
	}
	var lcr engine.LCRCost
	if err = smg.rater.Call("Responder.GetLCR", &engine.AttrGetLcr{CallDescriptor: cd}, &lcr); err != nil {
		return nil, err
	}
	if lcr.HasErrors() {
		lcr.LogErrors()
		return nil, errors.New("LCR_COMPUTE_ERROR")
	}
	return lcr.SuppliersSlice()
}

// replicateSessionsForEvent will replicate session based on configuration
func (smg *SMGeneric) replicateSessionsForEvent(gev SMGenericEvent) (err error) {
	if smg.cgrCfg.SmGenericConfig.DebitInterval != 0 {
		return
	}
	smg.aSessionsMux.RLock()
	aSessions := smg.activeSessions[gev.GetUUID()]
	smg.aSessionsMux.RUnlock()
	if len(aSessions) == 0 {
		return
	}
	var wg sync.WaitGroup
	for _, rplConn := range smg.smgReplConns {
		if rplConn.sync {
			wg.Add(1)
		}
		go func(conn rpcclient.RpcClientConnection, sync bool, ss []*SMGSession) {
			for _, s := range ss {
				var reply string
				if err := conn.Call("SMGenericV1.SetPassiveSession", s, &reply); err != nil {
					utils.Logger.Err(fmt.Sprintf("<SMGeneric> replicating session with id <%s>, run: <%s>:  received error: <%s>", gev.GetUUID(), s.RunID))
				}
			}
			if sync {
				wg.Done()
			}
		}(rplConn.conn, rplConn.sync, aSessions)
	}
	wg.Wait() // wait for synchronous replication to finish
	return
}

// Called on session start
func (smg *SMGeneric) InitiateSession(gev SMGenericEvent, clnt rpcclient.RpcClientConnection) (time.Duration, error) {
	defer smg.replicateSessionsForEvent(gev)
	if err := smg.sessionStart(gev, clnt); err != nil {
		smg.sessionEnd(gev.GetUUID(), 0)
		return nilDuration, err
	}
	if smg.cgrCfg.SmGenericConfig.DebitInterval != 0 { // Session handled by debit loop
		return -1, nil
	}
	d, err := smg.UpdateSession(gev, clnt)
	if err != nil || d == 0 {
		smg.sessionEnd(gev.GetUUID(), 0)
	}
	return d, err
}

// Execute debits for usage/maxUsage
func (smg *SMGeneric) UpdateSession(gev SMGenericEvent, clnt rpcclient.RpcClientConnection) (time.Duration, error) {
	if smg.cgrCfg.SmGenericConfig.DebitInterval != 0 { // Not possible to update a session with debit loop active
		return 0, errors.New("ACTIVE_DEBIT_LOOP")
	}
	defer smg.replicateSessionsForEvent(gev)
	if initialID, err := gev.GetFieldAsString(utils.InitialOriginID); err == nil {
		err := smg.sessionRelocate(gev.GetUUID(), initialID)
		if err == utils.ErrNotFound { // Session was already relocated, create a new  session with this update
			err = smg.sessionStart(gev, clnt)
		}
		if err != nil {
			return nilDuration, err
		}
	}
	smg.resetTerminatorTimer(gev.GetUUID(), gev.GetSessionTTL(), gev.GetSessionTTLLastUsed(), gev.GetSessionTTLUsage())
	var lastUsed *time.Duration
	if evLastUsed, err := gev.GetLastUsed(utils.META_DEFAULT); err == nil {
		lastUsed = &evLastUsed
	} else if err != utils.ErrNotFound {
		return nilDuration, err
	}
	evMaxUsage, err := gev.GetMaxUsage(utils.META_DEFAULT, smg.cgrCfg.SmGenericConfig.MaxCallDuration)
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrMandatoryIeMissing
		}
		return nilDuration, err
	}
	aSessions := smg.getASession(gev.GetUUID())
	if len(aSessions) == 0 {
		utils.Logger.Err(fmt.Sprintf("<SMGeneric> SessionUpdate with no active sessions for event: <%s>", gev.GetUUID()))
		return nilDuration, utils.ErrServerError
	}
	for _, s := range aSessions {
		if maxDur, err := s.debit(evMaxUsage, lastUsed); err != nil {
			return nilDuration, err
		} else if maxDur < evMaxUsage {
			evMaxUsage = maxDur
		}
	}
	return evMaxUsage, nil
}

// Called on session end, should stop debit loop
func (smg *SMGeneric) TerminateSession(gev SMGenericEvent, clnt rpcclient.RpcClientConnection) error {
	defer smg.replicateSessionsForEvent(gev)
	if initialID, err := gev.GetFieldAsString(utils.InitialOriginID); err == nil {
		err := smg.sessionRelocate(gev.GetUUID(), initialID)
		if err == utils.ErrNotFound { // Session was already relocated, create a new  session with this update
			err = smg.sessionStart(gev, clnt)
		}
		if err != nil && err != utils.ErrMandatoryIeMissing {
			return err
		}
	}
	sessionIDs := []string{gev.GetUUID()}
	if sessionIDPrefix, err := gev.GetFieldAsString(utils.OriginIDPrefix); err == nil { // OriginIDPrefix is present, OriginID will not be anymore considered
		sessionIDs = smg.getSessionIDsForPrefix(sessionIDPrefix)
	}
	usage, errUsage := gev.GetUsage(utils.META_DEFAULT)
	var lastUsed time.Duration
	if errUsage != nil {
		if errUsage != utils.ErrNotFound {
			return errUsage
		}
		var err error
		lastUsed, err = gev.GetLastUsed(utils.META_DEFAULT)
		if err != nil {
			if err == utils.ErrNotFound {
				err = utils.ErrMandatoryIeMissing
			}
			return err
		}
	}
	var interimError error
	var hasActiveSession bool
	for _, sessionID := range sessionIDs {
		var s *SMGSession
		for _, s = range smg.getASession(sessionID) {
			break
		}
		if s == nil {
			continue // No session active, will not be able to close it anyway
		}
		hasActiveSession = true
		if errUsage != nil {
			usage = s.TotalUsage - s.LastUsage + lastUsed
		}
		if err := smg.sessionEnd(sessionID, usage); err != nil {
			interimError = err // Last error will be the one returned as API result
		}
	}
	if !hasActiveSession {
		return utils.ErrNoActiveSession
	}
	return interimError
}

// Processes one time events (eg: SMS)
func (smg *SMGeneric) ChargeEvent(gev SMGenericEvent) (maxDur time.Duration, err error) {
	var sessionRuns []*engine.SessionRun
	if err := smg.rater.Call("Responder.GetSessionRuns", gev.AsStoredCdr(smg.cgrCfg, smg.timezone), &sessionRuns); err != nil {
		return nilDuration, err
	} else if len(sessionRuns) == 0 {
		return nilDuration, nil
	}
	var maxDurInit bool // Avoid differences between default 0 and received 0
	for _, sR := range sessionRuns {
		cc := new(engine.CallCost)
		if err = smg.rater.Call("Responder.MaxDebit", sR.CallDescriptor, cc); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not Debit CD: %+v, RunID: %s, error: %s", sR.CallDescriptor, sR.DerivedCharger.RunID, err.Error()))
			break
		}
		sR.CallCosts = append(sR.CallCosts, cc) // Save it so we can revert on issues
		if ccDur := cc.GetDuration(); ccDur == 0 {
			err = utils.ErrInsufficientCredit
			break
		} else if !maxDurInit || ccDur < maxDur {
			maxDur = ccDur
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
				cd.CgrID = sR.CallDescriptor.CgrID
				cd.RunID = sR.CallDescriptor.RunID
				cd.Increments.Compress()
				//utils.Logger.Info(fmt.Sprintf("Refunding session run callcost: %s", utils.ToJSON(cd)))
				var response float64
				err := smg.rater.Call("Responder.RefundIncrements", cd, &response)
				if err != nil {
					return nilDuration, err
				}
			}
		}
		return nilDuration, err
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
			if err := smg.rater.Call("Responder.RefundRounding", cd, &response); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SM> ERROR failed to refund rounding: %v", err))
			}
		}
		var reply string
		smCost := &engine.SMCost{
			CGRID:       gev.GetCgrId(smg.timezone),
			CostSource:  utils.SESSION_MANAGER_SOURCE,
			RunID:       sR.DerivedCharger.RunID,
			OriginHost:  gev.GetOriginatorIP(utils.META_DEFAULT),
			OriginID:    gev.GetUUID(),
			CostDetails: cc,
		}
		if err := smg.cdrsrv.Call("CdrsV1.StoreSMCost", engine.AttrCDRSStoreSMCost{Cost: smCost, CheckDuplicate: true}, &reply); err != nil && !strings.HasSuffix(err.Error(), utils.ErrExists.Error()) {
			withErrors = true
			utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not save CC: %+v, RunID: %s error: %s", cc, sR.DerivedCharger.RunID, err.Error()))
		}
	}
	if withErrors {
		return nilDuration, ErrPartiallyExecuted
	}
	return maxDur, nil
}

func (smg *SMGeneric) ProcessCDR(gev SMGenericEvent) error {
	var reply string
	if err := smg.cdrsrv.Call("CdrsV1.ProcessCDR", gev.AsStoredCdr(smg.cgrCfg, smg.timezone), &reply); err != nil {
		return err
	}
	return nil
}

func (smg *SMGeneric) Connect() error {
	return nil
}

// Used by APIer to retrieve sessions
func (smg *SMGeneric) getASessions() map[string][]*SMGSession {
	smg.aSessionsMux.RLock()
	defer smg.aSessionsMux.RUnlock()
	return smg.activeSessions
}

func (smg *SMGeneric) ActiveSessions(fltrs map[string]string, count bool) (aSessions []*ActiveSession, counter int, err error) {
	aSessions = make([]*ActiveSession, 0) // Make sure we return at least empty list and not nil
	// Check first based on indexes so we can downsize the list of matching sessions
	matchingSessionIDs, checkedFilters := smg.getSessionIDsMatchingIndexes(fltrs)
	if len(matchingSessionIDs) == 0 && len(checkedFilters) != 0 {
		return
	}
	for fltrFldName := range fltrs {
		if _, alreadyChecked := checkedFilters[fltrFldName]; alreadyChecked && fltrFldName != utils.MEDI_RUNID { // Optimize further checks, RunID should stay since it can create bugs
			delete(fltrs, fltrFldName)
		}
	}
	var remainingSessions []*SMGSession // Survived index matching
	for sUUID, sGrp := range smg.getASessions() {
		if _, hasUUID := matchingSessionIDs[sUUID]; !hasUUID && len(checkedFilters) != 0 {
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
		aSessions = append(aSessions, s.AsActiveSession(smg.Timezone())) // Expensive for large number of sessions
	}
	return
}

// SetPassiveSession will add or set a previously found session in passive session list
func (smg *SMGeneric) setPassiveSession(s *SMGSession) {
	smg.pSessionsMux.Lock()
	if ss, hasID := smg.passiveSessions[s.EventStart.GetUUID()]; !hasID {
		smg.passiveSessions[s.EventStart.GetUUID()] = []*SMGSession{s}
	} else {
		var exists bool
		for i, oldS := range ss {
			if oldS.RunID == s.RunID {
				smg.passiveSessions[s.EventStart.GetUUID()][i] = s
				exists = true
				break
			}
		}
		if !exists {
			smg.passiveSessions[s.EventStart.GetUUID()] = append(smg.passiveSessions[s.EventStart.GetUUID()], s)
		}
	}
	smg.pSessionsMux.Unlock()
}

func (smg *SMGeneric) getPassiveSessions(originID, runID string) (pss map[string][]*SMGSession) {
	smg.pSessionsMux.RLock()
	if originID == "" {
		if len(smg.passiveSessions) != 0 {
			pss = smg.passiveSessions
		}
	} else {
		pSSlc := smg.passiveSessions[originID]
		if runID != "" {
			var found bool
			for _, s := range pSSlc {
				if s.RunID == runID {
					found = true
					pSSlc = []*SMGSession{s}
				}
			}
			if !found {
				pSSlc = []*SMGSession{}
			}
		}
		if len(pSSlc) != 0 {
			pss = map[string][]*SMGSession{originID: pSSlc}
		}
	}
	smg.pSessionsMux.RUnlock()
	return
}

func (smg *SMGeneric) Timezone() string {
	return smg.timezone
}

// System shutdown
func (smg *SMGeneric) Shutdown() error {
	for ssId := range smg.getASessions() { // Force sessions shutdown
		smg.sessionEnd(ssId, time.Duration(smg.cgrCfg.MaxCallDuration))
	}
	return nil
}

// RpcClientConnection interface
func (smg *SMGeneric) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return smg.CallBiRPC(nil, serviceMethod, args, reply) // Capture the version part out of original call
}

// Part of utils.BiRPCServer to help internal connections do calls over rpcclient.RpcClientConnection interface
func (smg *SMGeneric) CallBiRPC(clnt rpcclient.RpcClientConnection, serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// get method BiRPCV1.Method
	method := reflect.ValueOf(smg).MethodByName("BiRPC" + parts[0][len(parts[0])-2:] + parts[1]) // Inherit the version V1 in the method name and add prefix
	if !method.IsValid() {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// construct the params
	var clntVal reflect.Value
	if clnt == nil {
		clntVal = reflect.New(reflect.TypeOf(new(utils.BiRPCInternalClient))).Elem() // Kinda cheat since we make up a type here
	} else {
		clntVal = reflect.ValueOf(clnt)
	}
	params := []reflect.Value{clntVal, reflect.ValueOf(args), reflect.ValueOf(reply)}
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

func (smg *SMGeneric) BiRPCV1MaxUsage(clnt rpcclient.RpcClientConnection, ev SMGenericEvent, maxUsage *float64) error {
	maxUsageDur, err := smg.MaxUsage(ev)
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

/// Returns list of suppliers which can be used for the request
func (smg *SMGeneric) BiRPCV1LCRSuppliers(clnt rpcclient.RpcClientConnection, ev SMGenericEvent, suppliers *[]string) error {
	if supls, err := smg.LCRSuppliers(ev); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*suppliers = supls
	}
	return nil
}

// Called on session start, returns the maximum number of seconds the session can last
func (smg *SMGeneric) BiRPCV1InitiateSession(clnt rpcclient.RpcClientConnection, ev SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := smg.InitiateSession(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Interim updates, returns remaining duration from the rater
func (smg *SMGeneric) BiRPCV1UpdateSession(clnt rpcclient.RpcClientConnection, ev SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := smg.UpdateSession(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Called on session end, should stop debit loop
func (smg *SMGeneric) BiRPCV1TerminateSession(clnt rpcclient.RpcClientConnection, ev SMGenericEvent, reply *string) error {
	if err := smg.TerminateSession(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Called on individual Events (eg SMS)
func (smg *SMGeneric) BiRPCV1ChargeEvent(clnt rpcclient.RpcClientConnection, ev SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := smg.ChargeEvent(ev); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Called on session end, should send the CDR to CDRS
func (smg *SMGeneric) BiRPCV1ProcessCDR(clnt rpcclient.RpcClientConnection, ev SMGenericEvent, reply *string) error {
	if err := smg.ProcessCDR(ev); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

func (smg *SMGeneric) BiRPCV1ActiveSessions(clnt rpcclient.RpcClientConnection, attrs utils.AttrSMGGetActiveSessions, reply *[]*ActiveSession) error {
	aSessions, _, err := smg.ActiveSessions(attrs.AsMapStringString(), false)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = aSessions
	return nil
}

func (smg *SMGeneric) BiRPCV1ActiveSessionsCount(attrs utils.AttrSMGGetActiveSessions, reply *int) error {
	if _, count, err := smg.ActiveSessions(attrs.AsMapStringString(), true); err != nil {
		return err
	} else {
		*reply = count
	}
	return nil
}

// BiRPCV1SetPassiveSession used for replicating SMGSessions
func (smg *SMGeneric) BiRPCV1SetPassiveSession(s SMGSession, reply *string) error {
	smg.setPassiveSession(&s)
	*reply = utils.OK
	return nil
}

type AttrGetPassiveSessions struct {
	OriginID string
	RunID    string
}

func (smg *SMGeneric) BiRPCV1GetPassiveSessions(attrs AttrGetPassiveSessions, pSessions *map[string][]*SMGSession) error {
	if attrs.RunID != "" && attrs.OriginID == "" {
		return utils.ErrMandatoryIeMissing
	}
	pSS := smg.getPassiveSessions(attrs.OriginID, attrs.RunID)
	if len(pSS) == 0 {
		return utils.ErrNotFound
	}
	*pSessions = pSS
	return nil
}
