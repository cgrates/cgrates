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
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Session type holding the call information fields, a session delegate for specific
// actions and a channel to signal end of the debit loop.
type Session struct {
	eventStart     engine.Event  // Store the original event who started this session so we can use it's info later (eg: disconnect, cgrid)
	stopDebit      chan struct{} // Channel to communicate with debit loops when closing the session
	sessionManager SessionManager
	connId         string // Reference towards connection id on the session manager side.
	warnMinDur     time.Duration
	sessionRuns    []*engine.SessionRun
}

func (s *Session) GetSessionRun(runid string) *engine.SessionRun {
	for _, sr := range s.sessionRuns {
		if sr.DerivedCharger.RunID == runid {
			return sr
		}
	}
	return nil
}

func (s *Session) SessionRuns() []*engine.SessionRun {
	return s.sessionRuns
}

// Creates a new session and in case of prepaid starts the debit loop for each of the session runs individually
func NewSession(ev engine.Event, connId string, sm SessionManager) *Session {
	s := &Session{eventStart: ev,
		stopDebit:      make(chan struct{}),
		sessionManager: sm,
		connId:         connId,
	}
	if err := sm.Rater().Call("Responder.GetSessionRuns", ev.AsStoredCdr(s.sessionManager.Timezone()), &s.sessionRuns); err != nil || len(s.sessionRuns) == 0 {
		return nil
	}
	for runIdx := range s.sessionRuns {
		go s.debitLoop(runIdx) // Send index of the just appended sessionRun
	}
	return s
}

// the debit loop method (to be stoped by sending somenthing on stopDebit channel)
func (s *Session) debitLoop(runIdx int) {
	nextCd := s.sessionRuns[runIdx].CallDescriptor
	nextCd.CgrID = s.eventStart.GetCgrId(s.sessionManager.Timezone())
	index := 0.0
	debitPeriod := s.sessionManager.DebitInterval()
	for {
		select {
		case <-s.stopDebit:
			return
		default:
		}
		if index > 0 { // first time use the session start time
			nextCd.TimeStart = nextCd.TimeEnd
		}
		nextCd.TimeEnd = nextCd.TimeStart.Add(debitPeriod)
		nextCd.LoopIndex = index
		nextCd.DurationIndex += debitPeriod // first presumed duration
		cc := new(engine.CallCost)
		if err := s.sessionManager.Rater().Call("Responder.MaxDebit", nextCd, cc); err != nil {
			utils.Logger.Err(fmt.Sprintf("Could not complete debit opperation: %v", err))
			if err.Error() == utils.ErrUnauthorizedDestination.Error() {
				s.sessionManager.DisconnectSession(s.eventStart, s.connId, UNAUTHORIZED_DESTINATION)
				return
			}
			s.sessionManager.DisconnectSession(s.eventStart, s.connId, SYSTEM_ERROR)
			return
		}
		if cc.GetDuration() == 0 {
			s.sessionManager.DisconnectSession(s.eventStart, s.connId, INSUFFICIENT_FUNDS)
			return
		}
		if s.warnMinDur != time.Duration(0) && cc.GetDuration() <= s.warnMinDur {
			s.sessionManager.WarnSessionMinDuration(s.eventStart.GetUUID(), s.connId)
		}
		s.sessionRuns[runIdx].CallCosts = append(s.sessionRuns[runIdx].CallCosts, cc)
		nextCd.TimeEnd = cc.GetEndTime() // set debited timeEnd
		// update call duration with real debited duration
		nextCd.DurationIndex -= debitPeriod
		nextCd.DurationIndex += cc.GetDuration()
		nextCd.MaxCostSoFar += cc.Cost
		time.Sleep(cc.GetDuration())
		index++
	}
}

// Stops the debit loop
func (s *Session) Close(ev engine.Event) error {
	close(s.stopDebit) // Close the channel so all the sessionRuns listening will be notified
	if _, err := ev.GetEndTime(utils.META_DEFAULT, s.sessionManager.Timezone()); err != nil {
		utils.Logger.Err("Error parsing event stop time.")
		for idx := range s.sessionRuns {
			s.sessionRuns[idx].CallDescriptor.TimeEnd = s.sessionRuns[idx].CallDescriptor.TimeStart.Add(s.sessionRuns[idx].CallDescriptor.DurationIndex)
		}
	}

	// Costs refunds
	for _, sr := range s.SessionRuns() {
		if len(sr.CallCosts) == 0 {
			continue // why would we have 0 callcosts
		}
		//utils.Logger.Debug(fmt.Sprintf("ALL CALLCOSTS: %s", utils.ToJSON(sr.CallCosts)))
		lastCC := sr.CallCosts[len(sr.CallCosts)-1]
		lastCC.Timespans.Decompress()
		// put credit back
		startTime, err := ev.GetAnswerTime(sr.DerivedCharger.AnswerTimeField, s.sessionManager.Timezone())
		if err != nil {
			utils.Logger.Crit("Error parsing prepaid call start time from event")
			return err
		}
		duration, err := ev.GetDuration(sr.DerivedCharger.UsageField)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("Error parsing call duration from event: %s", err.Error()))
			return err
		}
		hangupTime := startTime.Add(duration)
		//utils.Logger.Debug(fmt.Sprintf("BEFORE REFUND: %s", utils.ToJSON(lastCC)))
		err = s.Refund(lastCC, hangupTime)
		//utils.Logger.Debug(fmt.Sprintf("AFTER REFUND: %s", utils.ToJSON(lastCC)))
		if err != nil {
			return err
		}
	}
	go s.SaveOperations()
	return nil
}

func (s *Session) Refund(lastCC *engine.CallCost, hangupTime time.Time) error {
	end := lastCC.Timespans[len(lastCC.Timespans)-1].TimeEnd
	refundDuration := end.Sub(hangupTime)
	//utils.Logger.Debug(fmt.Sprintf("HANGUPTIME: %s REFUNDDURATION: %s", hangupTime.String(), refundDuration.String()))
	var refundIncrements engine.Increments
	for i := len(lastCC.Timespans) - 1; i >= 0; i-- {
		ts := lastCC.Timespans[i]
		tsDuration := ts.GetDuration()
		if refundDuration <= tsDuration {

			lastRefundedIncrementIndex := -1
			for j := len(ts.Increments) - 1; j >= 0; j-- {
				increment := ts.Increments[j]
				if increment.Duration <= refundDuration {
					refundIncrements = append(refundIncrements, increment)
					refundDuration -= increment.Duration
					lastRefundedIncrementIndex = j
				} else {
					break //increment duration is larger, cannot refund increment
				}
			}
			if lastRefundedIncrementIndex == 0 {
				lastCC.Timespans[i] = nil
				lastCC.Timespans = lastCC.Timespans[:i]
			} else {
				ts.SplitByIncrement(lastRefundedIncrementIndex)
				ts.Cost = ts.CalculateCost()
			}
			break // do not go to other timespans
		} else {
			refundIncrements = append(refundIncrements, ts.Increments...)
			// remove the timespan entirely
			lastCC.Timespans[i] = nil
			lastCC.Timespans = lastCC.Timespans[:i]
			// continue to the next timespan with what is left to refund
			refundDuration -= tsDuration
		}
	}
	// show only what was actualy refunded (stopped in timespan)
	// utils.Logger.Info(fmt.Sprintf("Refund duration: %v", initialRefundDuration-refundDuration))
	if len(refundIncrements) > 0 {
		cd := &engine.CallDescriptor{
			CgrID:       s.eventStart.GetCgrId(s.sessionManager.Timezone()),
			Direction:   lastCC.Direction,
			Tenant:      lastCC.Tenant,
			Category:    lastCC.Category,
			Subject:     lastCC.Subject,
			Account:     lastCC.Account,
			Destination: lastCC.Destination,
			TOR:         lastCC.TOR,
			Increments:  refundIncrements,
		}
		cd.Increments.Compress()
		//utils.Logger.Info(fmt.Sprintf("Refunding duration %v with cd: %+v", refundDuration, cd))
		var response float64
		err := s.sessionManager.Rater().Call("Responder.RefundIncrements", cd, &response)
		if err != nil {
			return err
		}
	}
	//utils.Logger.Debug(fmt.Sprintf("REFUND INCR: %s", utils.ToJSON(refundIncrements)))
	lastCC.Cost -= refundIncrements.GetTotalCost()
	lastCC.UpdateRatedUsage()
	lastCC.Timespans.Compress()
	return nil
}

// Nice print for session
func (s *Session) String() string {
	sDump, _ := json.Marshal(s)
	return string(sDump)
}

// Saves call_costs for each session run
func (s *Session) SaveOperations() {
	for _, sr := range s.sessionRuns {
		if len(sr.CallCosts) == 0 {
			break // There are no costs to save, ignore the operation
		}
		firstCC := sr.CallCosts[0]
		for _, cc := range sr.CallCosts[1:] {
			firstCC.Merge(cc)
		}
		firstCC.Timespans.Compress()

		firstCC.Round()
		roundIncrements := firstCC.GetRoundIncrements()
		if len(roundIncrements) != 0 {
			cd := firstCC.CreateCallDescriptor()
			cd.Increments = roundIncrements
			var response float64
			if err := s.sessionManager.Rater().Call("Responder.RefundRounding", cd, &response); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SM> ERROR failed to refund rounding: %v", err))
			}
		}
		smCost := &engine.SMCost{
			CGRID:       s.eventStart.GetCgrId(s.sessionManager.Timezone()),
			CostSource:  utils.SESSION_MANAGER_SOURCE,
			RunID:       sr.DerivedCharger.RunID,
			OriginHost:  s.eventStart.GetOriginatorIP(utils.META_DEFAULT),
			OriginID:    s.eventStart.GetUUID(),
			CostDetails: firstCC,
		}
		var reply string
		if err := s.sessionManager.CdrSrv().Call("CdrsV1.StoreSMCost", engine.AttrCDRSStoreSMCost{Cost: smCost, CheckDuplicate: true}, &reply); err != nil {
			// this is a protection against the case when the close event is missed for some reason
			// when the cdr arrives to cdrserver because our callcost is not there it will be rated
			// as postpaid. When the close event finally arives we have to refund everything
			if err == utils.ErrExists {
				s.Refund(firstCC, firstCC.Timespans[0].TimeStart)
			} else {
				utils.Logger.Err(fmt.Sprintf("<SM> ERROR failed to log call cost: %v", err))
			}
		}
	}
}

func (s *Session) AsActiveSessions() []*ActiveSession {
	var aSessions []*ActiveSession
	sTime, _ := s.eventStart.GetSetupTime(utils.META_DEFAULT, s.sessionManager.Timezone())
	aTime, _ := s.eventStart.GetAnswerTime(utils.META_DEFAULT, s.sessionManager.Timezone())
	usage, _ := s.eventStart.GetDuration(utils.META_DEFAULT)
	pdd, _ := s.eventStart.GetPdd(utils.META_DEFAULT)
	for _, sessionRun := range s.sessionRuns {
		aSession := &ActiveSession{
			CgrId:       s.eventStart.GetCgrId(s.sessionManager.Timezone()),
			TOR:         utils.VOICE,
			OriginID:    s.eventStart.GetUUID(),
			CdrHost:     s.eventStart.GetOriginatorIP(utils.META_DEFAULT),
			CdrSource:   "FS_" + s.eventStart.GetName(),
			ReqType:     s.eventStart.GetReqType(utils.META_DEFAULT),
			Direction:   s.eventStart.GetDirection(utils.META_DEFAULT),
			Tenant:      s.eventStart.GetTenant(utils.META_DEFAULT),
			Category:    s.eventStart.GetCategory(utils.META_DEFAULT),
			Account:     s.eventStart.GetAccount(utils.META_DEFAULT),
			Subject:     s.eventStart.GetSubject(utils.META_DEFAULT),
			Destination: s.eventStart.GetDestination(utils.META_DEFAULT),
			SetupTime:   sTime,
			AnswerTime:  aTime,
			Usage:       usage,
			Pdd:         pdd,
			ExtraFields: s.eventStart.GetExtraFields(),
			Supplier:    s.eventStart.GetSupplier(utils.META_DEFAULT),
			SMId:        "UNKNOWN",
		}
		if sessionRun.DerivedCharger != nil {
			aSession.RunId = sessionRun.DerivedCharger.RunID
		}
		if sessionRun.CallDescriptor != nil {
			aSession.LoopIndex = sessionRun.CallDescriptor.LoopIndex
			aSession.DurationIndex = sessionRun.CallDescriptor.DurationIndex
			aSession.MaxRate = sessionRun.CallDescriptor.MaxRate
			aSession.MaxRateUnit = sessionRun.CallDescriptor.MaxRateUnit
			aSession.MaxCostSoFar = sessionRun.CallDescriptor.MaxCostSoFar
		}
		aSessions = append(aSessions, aSession)
	}
	return aSessions
}

func (s *Session) AsMapStringIface() (map[string]interface{}, error) {
	mp := make(map[string]interface{})
	v := reflect.ValueOf(s).Elem()
	for i := 0; i < v.NumField(); i++ {
		mp[v.Type().Field(i).Name] = v.Field(i).Interface()
	}
	return mp, nil
}

// Will be used when displaying active sessions via RPC
type ActiveSession struct {
	CgrId         string
	TOR           string            // type of record, meta-field, should map to one of the TORs hardcoded inside the server <*voice|*data|*sms|*generic>
	OriginID      string            // represents the unique accounting id given by the telecom switch generating the CDR
	CdrHost       string            // represents the IP address of the host generating the CDR (automatically populated by the server)
	CdrSource     string            // formally identifies the source of the CDR (free form field)
	ReqType       string            // matching the supported request types by the **CGRateS**, accepted values are hardcoded in the server <prepaid|postpaid|pseudoprepaid|rated>.
	Direction     string            // matching the supported direction identifiers of the CGRateS <*out>
	Tenant        string            // tenant whom this record belongs
	Category      string            // free-form filter for this record, matching the category defined in rating profiles.
	Account       string            // account id (accounting subsystem) the record should be attached to
	Subject       string            // rating subject (rating subsystem) this record should be attached to
	Destination   string            // destination to be charged
	SetupTime     time.Time         // set-up time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
	Pdd           time.Duration     // PDD value
	AnswerTime    time.Time         // answer time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
	Usage         time.Duration     // event usage information (eg: in case of tor=*voice this will represent the total duration of a call)
	Supplier      string            // Supplier information when available
	ExtraFields   map[string]string // Extra fields to be stored in CDR
	SMId          string
	SMConnId      string
	RunId         string
	LoopIndex     float64       // indicates the position of this segment in a cost request loop
	DurationIndex time.Duration // the call duration so far (till TimeEnd)
	MaxRate       float64
	MaxRateUnit   time.Duration
	MaxCostSoFar  float64
}
