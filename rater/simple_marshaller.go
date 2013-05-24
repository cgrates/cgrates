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

package rater

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func Marshal(v interface{}) ([]byte, error) {
	switch i := v.(type) {
	case *ActivationPeriod:
		result, err := activationPeriodStore(i)
		return []byte(result), err
	case *RatingProfile:
		result, err := ratingProfileStore(i)
		return []byte(result), err
	}
	return nil, errors.New("Not supported type")
}

func Unmarshal(data []byte, v interface{}) error {
	switch i := v.(type) {
	case *ActivationPeriod:
		return activationPeriodRestore(string(data), i)
	case *RatingProfile:
		return ratingProfileRestore(string(data), i)
	}
	return errors.New("Not supported type")
}

func activationPeriodStore(ap *ActivationPeriod) (result string, err error) {
	result += strconv.FormatInt(ap.ActivationTime.UnixNano(), 10) + "|"
	for _, i := range ap.Intervals {
		result += i.store() + "|"
	}
	result = strings.TrimRight(result, "|")
	return
}

func activationPeriodRestore(input string, ap *ActivationPeriod) error {
	elements := strings.Split(input, "|")
	unixNano, _ := strconv.ParseInt(elements[0], 10, 64)
	ap.ActivationTime = time.Unix(0, unixNano).In(time.UTC)
	els := elements[1:]
	if len(els) > 1 {
		els = elements[1 : len(elements)-1]
	}
	for _, is := range els {
		i := &Interval{}
		i.restore(is)
		ap.Intervals = append(ap.Intervals, i)
	}
	return nil
}

func ratingProfileStore(rp *RatingProfile) (result string, err error) {
	result += rp.FallbackKey + ">"
	for k, aps := range rp.DestinationMap {
		result += k + "="
		for _, ap := range aps {
			aps, err := activationPeriodStore(ap)
			if err != nil {
				return result, err
			}
			result += aps + "<"
		}
		result = strings.TrimRight(result, "<")
		result += ">"
	}
	result = strings.TrimRight(result, ">")
	return
}

func ratingProfileRestore(input string, rp *RatingProfile) error {
	if rp.DestinationMap == nil {
		rp.DestinationMap = make(map[string][]*ActivationPeriod, 1)
	}
	elements := strings.Split(input, ">")
	rp.FallbackKey = elements[0]
	for _, kv := range elements[1:] {
		pair := strings.SplitN(kv, "=", 2)
		apList := strings.Split(pair[1], "<")
		var newAps []*ActivationPeriod
		for _, aps := range apList {
			ap := new(ActivationPeriod)
			err := activationPeriodRestore(aps, ap)
			if err != nil {
				return err
			}
			newAps = append(newAps, ap)
		}
		rp.DestinationMap[pair[0]] = newAps
	}
	return nil
}
