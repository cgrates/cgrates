package main

import (
	"time"
)

type Rating struct {
	ConnectFeeIn time.Time
	PriceIn float32
	ConnectFeeOut time.Time
	PriceOut float32
}

type Customer struct {
	Id string
	Prefix string
}
