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

package utils

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Returns first non empty string out of vals. Useful to extract defaults
func FirstNonEmpty(vals ...string) string {
	for _, val := range vals {
		if len(val) != 0 {
			return val
		}
	}
	return ""
}

func Sha1(attrs ...string) string {
	hasher := sha1.New()
	for _, attr := range attrs {
		hasher.Write([]byte(attr))
	}
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func NewTPid() string {
	return Sha1(GenUUID())
}

// helper function for uuid generation
func GenUUID() string {
	uuid := make([]byte, 16)
	n, err := rand.Read(uuid)
	if n != len(uuid) || err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	// TODO: verify the two lines implement RFC 4122 correctly
	uuid[8] = 0x80 // variant bits see page 5
	uuid[4] = 0x40 // version 4 Pseudo Random, see page 7

	return hex.EncodeToString(uuid)
}

// Round return rounded version of x with prec precision.
//
// Special cases are:
//	Round(±0) = ±0
//	Round(±Inf) = ±Inf
//	Round(NaN) = NaN
func Round(x float64, prec int, method string) float64 {
	var rounder float64
	maxPrec := 7 // define a max precison to cut float errors
	if maxPrec < prec {
		maxPrec = prec
	}
	pow := math.Pow(10, float64(prec))
	intermed := x * pow
	_, frac := math.Modf(intermed)

	switch method {
	case ROUNDING_UP:
		if frac >= math.Pow10(-maxPrec) { // Max precision we go, rest is float chaos
			rounder = math.Ceil(intermed)
		} else {
			rounder = math.Floor(intermed)
		}
	case ROUNDING_DOWN:
		rounder = math.Floor(intermed)
	case ROUNDING_MIDDLE:
		if frac >= 0.5 {
			rounder = math.Ceil(intermed)
		} else {
			rounder = math.Floor(intermed)
		}
	default:
		rounder = intermed
	}

	return rounder / pow
}

func ParseTimeDetectLayout(tmStr string) (time.Time, error) {
	var nilTime time.Time
	rfc3339Rule := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.+$`)
	sqlRule := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}$`)
	gotimeRule := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}\.?\d*\s[+,-]\d+\s\w+$`)
	fsTimestamp := regexp.MustCompile(`^\d{16}$`)
	unixTimestampRule := regexp.MustCompile(`^\d{10}$`)
	oneLineTimestampRule := regexp.MustCompile(`^\d{14}$`)
	twoSpaceTimestampRule := regexp.MustCompile(`^\d{2}\.\d{2}.\d{4}\s{2}\d{2}:\d{2}:\d{2}$`)
	switch {
	case rfc3339Rule.MatchString(tmStr):
		return time.Parse(time.RFC3339, tmStr)
	case gotimeRule.MatchString(tmStr):
		return time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", tmStr)
	case sqlRule.MatchString(tmStr):
		return time.Parse("2006-01-02 15:04:05", tmStr)
	case fsTimestamp.MatchString(tmStr):
		if tmstmp, err := strconv.ParseInt(tmStr+"000", 10, 64); err != nil {
			return nilTime, err
		} else {
			return time.Unix(0, tmstmp), nil
		}
	case unixTimestampRule.MatchString(tmStr):
		if tmstmp, err := strconv.ParseInt(tmStr, 10, 64); err != nil {
			return nilTime, err
		} else {
			return time.Unix(tmstmp, 0), nil
		}
	case tmStr == "0" || len(tmStr) == 0: // Time probably missing from request
		return nilTime, nil
	case oneLineTimestampRule.MatchString(tmStr):
		return time.Parse("20060102150405", tmStr)
	case twoSpaceTimestampRule.MatchString(tmStr):
		return time.Parse("02.01.2006  15:04:05", tmStr)
	}
	return nilTime, errors.New("Unsupported time format")
}

func ParseDate(date string) (expDate time.Time, err error) {
	date = strings.TrimSpace(date)
	switch {
	case date == "*unlimited" || date == "":
		// leave it at zero
	case strings.HasPrefix(date, "+"):
		d, err := time.ParseDuration(date[1:])
		if err != nil {
			return expDate, err
		}
		expDate = time.Now().Add(d)
	case date == "*daily":
		expDate = time.Now().AddDate(0, 0, 1) // add one day
	case date == "*monthly":
		expDate = time.Now().AddDate(0, 1, 0) // add one month
	case date == "*yearly":
		expDate = time.Now().AddDate(1, 0, 0) // add one year
	case strings.HasSuffix(date, "Z"):
		expDate, err = time.Parse(time.RFC3339, date)
	default:
		unix, err := strconv.ParseInt(date, 10, 64)
		if err != nil {
			return expDate, err
		}
		expDate = time.Unix(unix, 0)
	}
	return expDate, err
}

// returns a number equeal or larger than the amount that exactly
// is divisible to whole
func RoundTo(whole, amount time.Duration) time.Duration {
	a, w := float64(amount), float64(whole)
	if math.Mod(a, w) == 0 {
		return amount
	}
	return time.Duration((w - math.Mod(a, w)) + a)
}

func SplitPrefix(prefix string, minLength int) []string {
	length := int(math.Max(float64(len(prefix)-(minLength-1)), 0))
	subs := make([]string, length)
	max := len(prefix)
	for i := 0; i < length; i++ {
		subs[i] = prefix[:max-i]
	}
	return subs
}

func CopyHour(src, dest time.Time) time.Time {
	if src.Hour() == 0 && src.Minute() == 0 && src.Second() == 0 {
		return src
	}
	return time.Date(dest.Year(), dest.Month(), dest.Day(), src.Hour(), src.Minute(), src.Second(), src.Nanosecond(), src.Location())
}

// Parses duration, considers s as time unit if not provided
func ParseDurationWithSecs(durStr string) (time.Duration, error) {
	if _, err := strconv.Atoi(durStr); err == nil { // No suffix, default to seconds
		durStr += "s"
	}
	return time.ParseDuration(durStr)
}

func ParseDurationWithNanosecs(durStr string) (time.Duration, error) {
	if _, err := strconv.Atoi(durStr); err == nil { // No suffix, default to seconds
		durStr += "ns"
	}
	return time.ParseDuration(durStr)
}

func BalanceKey(tenant, account, direction string) string {
	return fmt.Sprintf("%s:%s:%s", direction, tenant, account)
}

// returns the minimum duration between the two
func MinDuration(d1, d2 time.Duration) time.Duration {
	if d1 < d2 {
		return d1
	}
	return d2
}

func ParseZeroRatingSubject(rateSubj string) (time.Duration, error) {
	if rateSubj == "" {
		rateSubj = ZERO_RATING_SUBJECT_PREFIX + "1s"
	}
	if !strings.HasPrefix(rateSubj, ZERO_RATING_SUBJECT_PREFIX) {
		return 0, errors.New("malformed rating subject: " + rateSubj)
	}
	durStr := rateSubj[len(ZERO_RATING_SUBJECT_PREFIX):]
	return time.ParseDuration(durStr)
}

func ConcatenatedKey(keyVals ...string) string {
	resKey := ""
	for idx, key := range keyVals {
		if idx != 0 {
			resKey += CONCATENATED_KEY_SEP
		}
		resKey += key
	}
	return resKey
}

func RatingSubjectAliasKey(tenant, subject string) string {
	return ConcatenatedKey(tenant, subject)
}

func AccountAliasKey(tenant, account string) string {
	return ConcatenatedKey(tenant, account)
}
