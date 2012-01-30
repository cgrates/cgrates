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
	tOR int
	cstmId, subject, destination string
	timeStart, timeEnd time.Time
}

func GetCost(in *CallDescription, sg StorageGetter) (result string, err error) {
	return "", nil
}

