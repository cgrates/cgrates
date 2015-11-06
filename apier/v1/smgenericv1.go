package v1

import (
	"time"

	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
)

// Exports RPC from SMGeneric
type SMGenericV1 struct {
	sm *sessionmanager.GenericSessionManager
}

// Returns MaxUsage (for calls in seconds), -1 for no limit
func (self *SMGenericV1) GetMaxUsage(ev sessionmanager.GenericEvent, maxUsage *float64) error {
	maxUsageDur, err := self.sm.GetMaxUsage(ev)
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

// Called on session start, returns the maximum number of seconds the session can last
func (self *SMGenericV1) SessionStart(ev sessionmanager.GenericEvent, maxUsage *float64) error {
	if err := self.sm.SessionStart(ev); err != nil {
		return utils.NewErrServerError(err)
	}
	return self.GetMaxUsage(ev, maxUsage)
}

// Interim updates, returns remaining duration from the rater
func (self *SMGenericV1) SessionUpdate(ev sessionmanager.GenericEvent, maxUsage *float64) error {
	if err := self.sm.SessionUpdate(ev); err != nil {
		return utils.NewErrServerError(err)
	}
	return self.GetMaxUsage(ev, maxUsage)
}

// Called on session end, should stop debit loop
func (self *SMGenericV1) SessionEnd(ev sessionmanager.GenericEvent, reply *string) error {
	if err := self.sm.SessionEnd(ev); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Called on session end, should send the CDR to CDRS
func (self *SMGenericV1) ProcessCdr(ev sessionmanager.GenericEvent, reply *string) error {
	if err := self.sm.ProcessCdr(ev); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
