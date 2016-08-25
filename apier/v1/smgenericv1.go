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
func (self *SMGenericV1) MaxUsage(ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	maxUsageDur, err := self.sm.MaxUsage(ev, nil)
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
func (self *SMGenericV1) LCRSuppliers(ev sessionmanager.SMGenericEvent, suppliers *[]string) error {
	if supls, err := self.sm.LCRSuppliers(ev, nil); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*suppliers = supls
	}
	return nil
}

// Called on session start, returns the maximum number of seconds the session can last
func (self *SMGenericV1) InitiateSession(ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := self.sm.InitiateSession(ev, nil); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Interim updates, returns remaining duration from the rater
func (self *SMGenericV1) UpdateSession(ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := self.sm.UpdateSession(ev, nil); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Called on session end, should stop debit loop
func (self *SMGenericV1) TerminateSession(ev sessionmanager.SMGenericEvent, reply *string) error {
	if err := self.sm.TerminateSession(ev, nil); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Called on individual Events (eg SMS)
func (self *SMGenericV1) ChargeEvent(ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := self.sm.ChargeEvent(ev, nil); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Called on session end, should send the CDR to CDRS
func (self *SMGenericV1) ProcessCDR(ev sessionmanager.SMGenericEvent, reply *string) error {
	if err := self.sm.ProcessCDR(ev); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

func (self *SMGenericV1) ActiveSessions(attrs utils.AttrSMGGetActiveSessions, reply *[]*sessionmanager.ActiveSession) error {
	aSessions, err := self.sm.ActiveSessions(attrs.AsMapStringString())
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = aSessions
	return nil
}

func (self *SMGenericV1) ActiveSessionsCount(attrs utils.AttrSMGGetActiveSessions, reply *int) error {
	var aSessions []*sessionmanager.ActiveSession
	if err := self.ActiveSessions(attrs, &aSessions); err != nil {
		return err
	}
	*reply = len(aSessions)
	return nil
}

// rpcclient.RpcClientConnection interface
func (self *SMGenericV1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	switch serviceMethod {
	case "SMGenericV1.MaxUsage":
		argsConverted, canConvert := args.(sessionmanager.SMGenericEvent)
		if !canConvert {
			return rpcclient.ErrWrongArgsType
		}
		replyConverted, canConvert := reply.(*float64)
		if !canConvert {
			return rpcclient.ErrWrongReplyType
		}
		self.MaxUsage(argsConverted, replyConverted)
	case "SMGenericV1.LCRSuppliers":
		argsConverted, canConvert := args.(sessionmanager.SMGenericEvent)
		if !canConvert {
			return rpcclient.ErrWrongArgsType
		}
		replyConverted, canConvert := reply.(*[]string)
		if !canConvert {
			return rpcclient.ErrWrongReplyType
		}
		return self.LCRSuppliers(argsConverted, replyConverted)
	case "SMGenericV1.InitiateSession":
		argsConverted, canConvert := args.(sessionmanager.SMGenericEvent)
		if !canConvert {
			return rpcclient.ErrWrongArgsType
		}
		replyConverted, canConvert := reply.(*float64)
		if !canConvert {
			return rpcclient.ErrWrongReplyType
		}
		return self.InitiateSession(argsConverted, replyConverted)
	case "SMGenericV1.UpdateSession":
		argsConverted, canConvert := args.(sessionmanager.SMGenericEvent)
		if !canConvert {
			return rpcclient.ErrWrongArgsType
		}
		replyConverted, canConvert := reply.(*float64)
		if !canConvert {
			return rpcclient.ErrWrongReplyType
		}
		return self.UpdateSession(argsConverted, replyConverted)
	case "SMGenericV1.TerminateSession":
		argsConverted, canConvert := args.(sessionmanager.SMGenericEvent)
		if !canConvert {
			return rpcclient.ErrWrongArgsType
		}
		replyConverted, canConvert := reply.(*string)
		if !canConvert {
			return rpcclient.ErrWrongReplyType
		}
		return self.TerminateSession(argsConverted, replyConverted)
	case "SMGenericV1.ChargeEvent":
		argsConverted, canConvert := args.(sessionmanager.SMGenericEvent)
		if !canConvert {
			return rpcclient.ErrWrongArgsType
		}
		replyConverted, canConvert := reply.(*float64)
		if !canConvert {
			return rpcclient.ErrWrongReplyType
		}
		return self.ChargeEvent(argsConverted, replyConverted)
	case "SMGenericV1.ProcessCDR":
		argsConverted, canConvert := args.(sessionmanager.SMGenericEvent)
		if !canConvert {
			return rpcclient.ErrWrongArgsType
		}
		replyConverted, canConvert := reply.(*string)
		if !canConvert {
			return rpcclient.ErrWrongReplyType
		}
		return self.ProcessCDR(argsConverted, replyConverted)
	case "SMGenericV1.ActiveSessions":
		argsConverted, canConvert := args.(utils.AttrSMGGetActiveSessions)
		if !canConvert {
			return rpcclient.ErrWrongArgsType
		}
		replyConverted, canConvert := reply.(*[]*sessionmanager.ActiveSession)
		if !canConvert {
			return rpcclient.ErrWrongReplyType
		}
		return self.ActiveSessions(argsConverted, replyConverted)

	case "SMGenericV1.ActiveSessionsCount":
		argsConverted, canConvert := args.(utils.AttrSMGGetActiveSessions)
		if !canConvert {
			return rpcclient.ErrWrongArgsType
		}
		replyConverted, canConvert := reply.(*int)
		if !canConvert {
			return rpcclient.ErrWrongReplyType
		}
		return self.ActiveSessionsCount(argsConverted, replyConverted)
	}
	return rpcclient.ErrUnsupporteServiceMethod
}
