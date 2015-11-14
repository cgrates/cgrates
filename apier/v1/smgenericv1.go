package v1

import (
	"time"

	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func NewSMGenericV1(sm *sessionmanager.SMGeneric) *SMGenericV1 {
	return &SMGenericV1{sm: sm}
}

// Exports RPC from SMGeneric
type SMGenericV1 struct {
	sm *sessionmanager.SMGeneric
}

// Returns MaxUsage (for calls in seconds), -1 for no limit
func (self *SMGenericV1) GetMaxUsage(ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	maxUsageDur, err := self.sm.GetMaxUsage(ev, nil)
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

// Returns list of suppliers which can be used for the request
func (self *SMGenericV1) GetLcrSuppliers(ev sessionmanager.SMGenericEvent, suppliers *[]string) error {
	if supls, err := self.sm.GetLcrSuppliers(ev, nil); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*suppliers = supls
	}
	return nil
}

// Called on session start, returns the maximum number of seconds the session can last
func (self *SMGenericV1) SessionStart(ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := self.sm.SessionStart(ev, nil); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Interim updates, returns remaining duration from the rater
func (self *SMGenericV1) SessionUpdate(ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := self.sm.SessionUpdate(ev, nil); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Called on session end, should stop debit loop
func (self *SMGenericV1) SessionEnd(ev sessionmanager.SMGenericEvent, reply *string) error {
	if err := self.sm.SessionEnd(ev, nil); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Called on session end, should send the CDR to CDRS
func (self *SMGenericV1) ProcessCdr(ev sessionmanager.SMGenericEvent, reply *string) error {
	if err := self.sm.ProcessCdr(ev); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// rpcclient.RpcClientConnection interface
func (self *SMGenericV1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	switch serviceMethod {
	case "SMGenericV1.GetLcrSuppliers":
		argsConverted, canConvert := args.(sessionmanager.SMGenericEvent)
		if !canConvert {
			return rpcclient.ErrWrongArgsType
		}
		replyConverted, canConvert := args.(*[]string)
		if !canConvert {
			return rpcclient.ErrWrongReplyType
		}
		return self.GetLcrSuppliers(argsConverted, replyConverted)
		/*smg.handlers["SMGenericV1.GetLcrSuppliers"] = smg.GetLcrSuppliers
		smg.handlers["SMGenericV1.SessionStart"] = smg.SessionStart
		smg.handlers["SMGenericV1.SessionUpdate"] = smg.SessionUpdate
		smg.handlers["SMGenericV1.SessionEnd"] = smg.SessionEnd
		smg.handlers["SMGenericV1.ProcessCdr"] = smg.ProcessCdr
		*/
	}
	return rpcclient.ErrUnsupporteServiceMethod
}
