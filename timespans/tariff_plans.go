/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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
package timespans

import (
	// "log"
	"strconv"
	"strings"
)

/*
Structure describing a tariff plan's number of bonus items. It is uset to restore
these numbers to the user budget every month.
*/
type TariffPlan struct {
	Id                       string
	SmsCredit                float64
	Traffic                  float64
	ReceivedCallSecondsLimit float64
	RecivedCallBonus         *RecivedCallBonus
	MinuteBuckets            []*MinuteBucket
	VolumeDiscountThresholds []*VolumeDiscount
}

/*
Structure that holds the thresholds and  for which 
*/
type VolumeDiscount struct {
	Volume   float64
	Discount float64 // procentage
}

type RecivedCallBonus struct {
	Credit        float64
	SmsCredit     float64
	Traffic       float64
	MinuteBuckets *MinuteBucket
}

/*
Serializes the tariff plan for the storage. Used for key-value storages.
*/
func (tp *TariffPlan) store() (result string) {
	result += strconv.FormatFloat(tp.SmsCredit, 'f', -1, 64) + ";"
	result += strconv.FormatFloat(tp.Traffic, 'f', -1, 64) + ";"
	result += strconv.FormatFloat(tp.ReceivedCallSecondsLimit, 'f', -1, 64) + ";"
	for i, mb := range tp.MinuteBuckets {
		if i > 0 {
			result += ","
		}
		result += mb.store()
	}
	result += ";"
	for i, vd := range tp.VolumeDiscountThresholds {
		if i > 0 {
			result += ","
		}
		result += strconv.FormatFloat(vd.Volume, 'f', -1, 64) + "|" + strconv.FormatFloat(vd.Discount, 'f', -1, 64)
	}
	result = strings.TrimRight(result, ";")
	return
}

/*
De-serializes the tariff plan for the storage. Used for key-value storages.
*/
func (tp *TariffPlan) restore(input string) {
	elements := strings.Split(input, ";")
	tp.SmsCredit, _ = strconv.ParseFloat(elements[0], 64)
	tp.Traffic, _ = strconv.ParseFloat(elements[1], 64)
	tp.ReceivedCallSecondsLimit, _ = strconv.ParseFloat(elements[2], 64)
	for _, mbs := range strings.Split(elements[3], ",") {
		mb := &MinuteBucket{}
		mb.restore(mbs)
		tp.MinuteBuckets = append(tp.MinuteBuckets, mb)
	}
	if len(elements) > 4 {
		for _, vdss := range strings.Split(elements[4], ",") {
			vd := &VolumeDiscount{}
			vds := strings.Split(vdss, "|")
			vd.Volume, _ = strconv.ParseFloat(vds[0], 64)
			vd.Discount, _ = strconv.ParseFloat(vds[1], 64)
			tp.VolumeDiscountThresholds = append(tp.VolumeDiscountThresholds, vd)
		}
	}
}
