package main

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


