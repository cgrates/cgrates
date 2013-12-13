/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package engine

import (
	"fmt"
	"sort"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// Can hold different units as seconds or monetary
type Balance struct {
	Uuid           string
	Value          float64
	ExpirationDate time.Time
	Weight         float64
	GroupIds       []string
	DestinationId  string
	RateSubject    string
	precision      int
}

func (b *Balance) Equal(o *Balance) bool {
	if b.DestinationId == "" {
		b.DestinationId = utils.ANY
	}
	if o.DestinationId == "" {
		o.DestinationId = utils.ANY
	}
	return b.ExpirationDate.Equal(o.ExpirationDate) &&
		b.Weight == o.Weight &&
		b.DestinationId == o.DestinationId &&
		b.RateSubject == o.RateSubject
}

// the default balance has no destinationid, Expirationdate or ratesubject
func (b *Balance) IsDefault() bool {
	return (b.DestinationId == "" || b.DestinationId == utils.ANY) && b.RateSubject == "" && b.ExpirationDate.IsZero()
}

func (b *Balance) IsExpired() bool {
	return !b.ExpirationDate.IsZero() && b.ExpirationDate.Before(time.Now())
}

func (b *Balance) HasDestination() bool {
	return b.DestinationId != "" && b.DestinationId != utils.ANY
}

func (b *Balance) MatchDestination(destinationId string) bool {
	return !b.HasDestination() || b.DestinationId == destinationId
}

func (b *Balance) Clone() *Balance {
	return &Balance{
		Uuid:           b.Uuid,
		Value:          b.Value, // this value is in seconds
		DestinationId:  b.DestinationId,
		ExpirationDate: b.ExpirationDate,
		Weight:         b.Weight,
		RateSubject:    b.RateSubject,
	}
}

// Returns the available number of seconds for a specified credit
func (b *Balance) GetMinutesForCredit(cd *CallDescriptor, initialCredit float64) (duration time.Duration, credit float64) {
	duration = time.Duration(b.Value) * time.Second
	credit = initialCredit
	cc, err := b.GetCost(cd)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error getting new cost for balance subject: %v", err))
		return 0, credit
	}
	if cc.Cost > 0 {
		duration = 0
		for _, ts := range cc.Timespans {
			ts.createIncrementsSlice()
			for _, incr := range ts.Increments {
				if incr.Cost <= credit {
					credit -= incr.Cost
					duration += incr.Duration
				} else {
					return
				}
			}
		}
	}
	return
}

func (b *Balance) GetCost(cd *CallDescriptor) (*CallCost, error) {
	if b.RateSubject != "" {
		cd.Subject = b.RateSubject
		cd.Account = cd.Subject
		return cd.GetCost()
	}
	cc := cd.CreateCallCost()
	cc.Cost = 0
	return cc, nil
}

func (b *Balance) DebitMinutes(cc *CallCost, count bool, ub *UserBalance, moneyBalances BalanceChain) error {
	for tsIndex := 0; tsIndex < len(cc.Timespans); tsIndex++ {
		ts := cc.Timespans[tsIndex]
		if ts.Increments == nil {
			ts.createIncrementsSlice()
		}
		if paid, _ := ts.IsPaid(); paid {
			continue
		}
		tsWasSplit := false
		for incrementIndex, increment := range ts.Increments {
			if tsWasSplit {
				break
			}
			if increment.paid {
				continue
			}
			if b.RateSubject == ZEROSECOND || b.RateSubject == "" {
				amount := increment.Duration.Seconds()
				if b.Value >= amount {
					b.Value -= amount
					increment.SetMinuteBalance(b.Uuid)
					increment.MinuteInfo = &MinuteInfo{cc.Destination, amount, 0}
					increment.paid = true
					if count {
						ub.countUnits(&Action{BalanceId: MINUTES, Direction: cc.Direction, Balance: &Balance{Value: amount, DestinationId: cc.Destination}})
					}
				}
				continue
			}
			if b.RateSubject == ZEROMINUTE {
				amount := time.Minute.Seconds()
				if b.Value >= amount { // balance has at least 60 seconds
					newTs := ts
					if incrementIndex != 0 {
						// if increment it's not at the begining we must split the timespan
						newTs = ts.SplitByIncrement(incrementIndex)
					}
					newTs.RoundToDuration(time.Minute)
					newTs.RateInterval = &RateInterval{
						Rating: &RIRate{
							Rates: RateGroups{
								&Rate{
									GroupIntervalStart: 0,
									Value:              0,
									RateIncrement:      time.Minute,
									RateUnit:           time.Minute,
								},
							},
						},
					}
					newTs.createIncrementsSlice()
					// insert the new timespan
					if newTs != ts {
						tsIndex++
						cc.Timespans = append(cc.Timespans, nil)
						copy(cc.Timespans[tsIndex+1:], cc.Timespans[tsIndex:])
						cc.Timespans[tsIndex] = newTs
						tsWasSplit = true
					}
					cc.Timespans.RemoveOverlapedFromIndex(tsIndex)
					b.Value -= amount
					newTs.Increments[0].SetMinuteBalance(b.Uuid)
					newTs.Increments[0].MinuteInfo = &MinuteInfo{cc.Destination, amount, 0}
					newTs.Increments[0].paid = true
					if count {
						ub.countUnits(&Action{BalanceId: MINUTES, Direction: cc.Direction, Balance: &Balance{Value: amount, DestinationId: cc.Destination}})
					}
				}
				continue
			}
			// get the new rate
			cd := cc.CreateCallDescriptor()
			cd.Subject = b.RateSubject
			cd.TimeStart = ts.GetTimeStartForIncrement(incrementIndex)
			cd.TimeEnd = cc.Timespans[len(cc.Timespans)-1].TimeEnd
			cd.CallDuration = cc.Timespans[len(cc.Timespans)-1].CallDuration
			newCC, err := b.GetCost(cd)
			if err != nil {
				Logger.Err(fmt.Sprintf("Error getting new cost for balance subject: %v", err))
				continue
			}
			//debit new callcost
			var paidTs []*TimeSpan
			for _, nts := range newCC.Timespans {
				nts.createIncrementsSlice()
				paidTs = append(paidTs, nts)
				for _, nInc := range nts.Increments {
					// debit minutes and money
					seconds := nInc.Duration.Seconds()
					cost := nInc.Cost
					var moneyBal *Balance
					for _, mb := range moneyBalances {
						if mb.Value >= cost {
							moneyBal = mb
							break
						}
					}
					if moneyBal != nil && b.Value >= seconds {
						b.Value -= seconds
						moneyBal.Value -= cost

						nInc.SetMinuteBalance(b.Uuid)
						nInc.SetMoneyBalance(moneyBal.Uuid)
						nInc.MinuteInfo = &MinuteInfo{newCC.Destination, seconds, 0}
						nInc.paid = true
						if count {
							ub.countUnits(&Action{BalanceId: MINUTES, Direction: newCC.Direction, Balance: &Balance{Value: seconds, DestinationId: newCC.Destination}})
							ub.countUnits(&Action{BalanceId: CREDIT, Direction: newCC.Direction, Balance: &Balance{Value: cost, DestinationId: newCC.Destination}})
						}
					} else {
						increment.paid = false
						break
					}
				}
			}
			newTs := ts.SplitByIncrement(incrementIndex)
			increment.paid = (&cc.Timespans).OverlapWithTimeSpans(paidTs, newTs, tsIndex)
			tsWasSplit = increment.paid
		}
	}
	return nil
}

func (b *Balance) DebitMoney(cc *CallCost, count bool, ub *UserBalance) error {
	for tsIndex := 0; tsIndex < len(cc.Timespans); tsIndex++ {
		ts := cc.Timespans[tsIndex]
		if ts.Increments == nil {
			ts.createIncrementsSlice()
		}
		if paid, _ := ts.IsPaid(); paid {
			continue
		}
		tsWasSplit := false
		for incrementIndex, increment := range ts.Increments {
			if tsWasSplit {
				break
			}
			if increment.paid {
				continue
			}
			// check standard subject tags
			if b.RateSubject == "" {
				amount := increment.Cost
				if b.Value >= amount {
					b.Value -= amount
					increment.SetMoneyBalance(b.Uuid)
					increment.paid = true
					if count {
						ub.countUnits(&Action{BalanceId: CREDIT, Direction: cc.Direction, Balance: &Balance{Value: amount, DestinationId: cc.Destination}})
					}
				}
			} else {
				// get the new rate
				cd := cc.CreateCallDescriptor()
				cd.Subject = b.RateSubject
				cd.TimeStart = ts.GetTimeStartForIncrement(incrementIndex)
				cd.TimeEnd = cc.Timespans[len(cc.Timespans)-1].TimeEnd
				cd.CallDuration = cc.Timespans[len(cc.Timespans)-1].CallDuration
				newCC, err := b.GetCost(cd)
				if err != nil {
					Logger.Err(fmt.Sprintf("Error getting new cost for balance subject: %v", err))
					continue
				}
				//debit new callcost
				var paidTs []*TimeSpan
				for _, nts := range newCC.Timespans {
					nts.createIncrementsSlice()
					paidTs = append(paidTs, nts)
					for _, nInc := range nts.Increments {
						// debit money
						amount := nInc.Cost
						if b.Value >= amount {
							b.Value -= amount
							nInc.SetMoneyBalance(b.Uuid)
							nInc.paid = true
							if count {
								ub.countUnits(&Action{BalanceId: CREDIT, Direction: newCC.Direction, Balance: &Balance{Value: amount, DestinationId: newCC.Destination}})
							}
						} else {
							increment.paid = false
							break
						}
					}
				}
				newTs := ts.SplitByIncrement(incrementIndex)
				increment.paid = (&cc.Timespans).OverlapWithTimeSpans(paidTs, newTs, tsIndex)
				tsWasSplit = increment.paid
			}
		}
	}
	return nil
}

/*
Structure to store minute buckets according to weight, precision or price.
*/
type BalanceChain []*Balance

func (bc BalanceChain) Len() int {
	return len(bc)
}

func (bc BalanceChain) Swap(i, j int) {
	bc[i], bc[j] = bc[j], bc[i]
}

func (bc BalanceChain) Less(j, i int) bool {
	return bc[i].Weight < bc[j].Weight ||
		bc[i].precision < bc[j].precision
}

func (bc BalanceChain) Sort() {
	sort.Sort(bc)
}

func (bc BalanceChain) GetTotalValue() (total float64) {
	for _, b := range bc {
		if !b.IsExpired() {
			total += b.Value
		}
	}
	return
}

func (bc BalanceChain) Debit(amount float64) float64 {
	bc.Sort()
	for i, b := range bc {
		if b.IsExpired() {
			continue
		}
		if b.Value >= amount || i == len(bc)-1 { // if last one go negative
			b.Value -= amount
			break
		}
		b.Value = 0
		amount -= b.Value
	}
	return bc.GetTotalValue()
}

func (bc BalanceChain) Equal(o BalanceChain) bool {
	if len(bc) != len(o) {
		return false
	}
	bc.Sort()
	o.Sort()
	for i := 0; i < len(bc); i++ {
		if !bc[i].Equal(o[i]) {
			return false
		}
	}
	return true
}

func (bc BalanceChain) Clone() BalanceChain {
	var newChain BalanceChain
	for _, b := range bc {
		newChain = append(newChain, b.Clone())
	}
	return newChain
}

func (bc BalanceChain) GetBalance(uuid string) *Balance {
	for _, balance := range bc {
		if balance.Uuid == uuid {
			return balance
		}
	}
	return nil
}

func (bc BalanceChain) HasBalance(balance *Balance) bool {
	for _, b := range bc {
		if b.Equal(balance) {
			return true
		}
	}
	return false
}
