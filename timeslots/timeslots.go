package timeslots

import (
	"time"
)

type BilingUnit int

type RatingProfile struct {
	StartTime time.Time
	ConnectFee float32
	Price float32
	BillingUnit BilingUnit
}

type ActivationPeriod struct {
	ActivationTime time.Time
	RatingProfiles []RatingProfile
}

type Customer struct {
	Id string
	Prefix string
	ActivationPeriods []ActivationPeriod
}

const (
	SECONDS =iota
	COUNT
	BYTES
)

type CallDescription struct {
	TOR int
	CstmId, Subject, Destination string
	TimeStart, TimeEnd time.Time
}

type CallCost struct {
	TOR int
	CstmId, Subject, Prefix string
	Cost, ConnectFee float32
//	ratesInfo *RatingProfile
}

func GetCost(in *CallDescription, sg StorageGetter) (result *CallCost, err error) {
	return &CallCost{TOR: 1, CstmId:"",Subject:"", Prefix:"", Cost:1, ConnectFee:1}, nil
}

