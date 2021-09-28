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

package engine

// type used for showing sane data cost
type DataCost struct {
	Category, Tenant, Subject, Account,
	Destination, ToR string
	Cost             float64
	DataSpans        []*DataSpan
	deductConnectFee bool
}
type DataSpan struct {
	DataStart, DataEnd float64
	Cost               float64
	ratingInfo         *RatingInfo
	RateInterval       *RateInterval
	DataIndex          float64 // the data transfer so far till DataEnd
	Increments         []*DataIncrement
	MatchedSubject, MatchedPrefix,
	MatchedDestId, RatingPlanId string
}

type DataIncrement struct {
	Amount         float64
	Cost           float64
	BalanceInfo    *DebitInfo // need more than one for units with cost
	CompressFactor int
}
