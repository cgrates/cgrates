/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package utils

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
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

/*func GenUUID() string {
	uuid := make([]byte, 16)
	n, err := rand.Read(uuid)
	if n != len(uuid) || err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	// TODO: verify the two lines implement RFC 4122 correctly
	uuid[8] = 0x80 // variant bits see page 5
	uuid[4] = 0x40 // version 4 Pseudo Random, see page 7

	return hex.EncodeToString(uuid)
}*/

// helper function for uuid generation
func GenUUID() string {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		log.Fatal(err)
	}
	b[6] = (b[6] & 0x0F) | 0x40
	b[8] = (b[8] &^ 0x40) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10],
		b[10:])
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

func ParseTimeDetectLayout(tmStr string, timezone string) (time.Time, error) {
	var nilTime time.Time
	if len(tmStr) == 0 {
		return nilTime, nil
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nilTime, err
	}
	rfc3339Rule := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.+$`)
	sqlRule := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}$`)
	gotimeRule := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}\.?\d*\s[+,-]\d+\s\w+$`)
	fsTimestamp := regexp.MustCompile(`^\d{16}$`)
	unixTimestampRule := regexp.MustCompile(`^\d{10}$`)
	oneLineTimestampRule := regexp.MustCompile(`^\d{14}$`)
	oneSpaceTimestampRule := regexp.MustCompile(`^\d{2}\.\d{2}.\d{4}\s{1}\d{2}:\d{2}:\d{2}$`)
	eamonTimestampRule := regexp.MustCompile(`^\d{2}/\d{2}/\d{4}\s{1}\d{2}:\d{2}:\d{2}$`)
	switch {
	case rfc3339Rule.MatchString(tmStr):
		return time.Parse(time.RFC3339, tmStr)
	case gotimeRule.MatchString(tmStr):
		return time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", tmStr)
	case sqlRule.MatchString(tmStr):
		return time.ParseInLocation("2006-01-02 15:04:05", tmStr, loc)
	case fsTimestamp.MatchString(tmStr):
		if tmstmp, err := strconv.ParseInt(tmStr+"000", 10, 64); err != nil {
			return nilTime, err
		} else {
			return time.Unix(0, tmstmp).In(loc), nil
		}
	case unixTimestampRule.MatchString(tmStr):
		if tmstmp, err := strconv.ParseInt(tmStr, 10, 64); err != nil {
			return nilTime, err
		} else {
			return time.Unix(tmstmp, 0).In(loc), nil
		}
	case tmStr == "0" || len(tmStr) == 0: // Time probably missing from request
		return nilTime, nil
	case oneLineTimestampRule.MatchString(tmStr):
		return time.ParseInLocation("20060102150405", tmStr, loc)
	case oneSpaceTimestampRule.MatchString(tmStr):
		return time.ParseInLocation("02.01.2006  15:04:05", tmStr, loc)
	case eamonTimestampRule.MatchString(tmStr):
		return time.ParseInLocation("02/01/2006 15:04:05", tmStr, loc)
	case tmStr == "*now":
		return time.Now(), nil
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
	case date == "*end_month":
		expDate = GetEndOfMonth(time.Now())
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

// returns a number equal or larger than the amount that exactly
// is divisible to whole
func RoundDuration(whole, amount time.Duration) time.Duration {
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

// Parses duration, considers s as time unit if not provided, seconds as float to specify subunits
func ParseDurationWithSecs(durStr string) (time.Duration, error) {
	if durSecs, err := strconv.ParseFloat(durStr, 64); err == nil { // Seconds format considered
		durNanosecs := int(durSecs * NANO_MULTIPLIER)
		return time.Duration(durNanosecs), nil
	} else {
		return time.ParseDuration(durStr)
	}
}

func AccountKey(tenant, account string) string {
	return fmt.Sprintf("%s:%s", tenant, account)
}

// returns the minimum duration between the two
func MinDuration(d1, d2 time.Duration) time.Duration {
	if d1 < d2 {
		return d1
	}
	return d2
}

func ParseZeroRatingSubject(rateSubj string) (time.Duration, error) {
	rateSubj = strings.TrimSpace(rateSubj)
	if rateSubj == "" || rateSubj == ANY {
		rateSubj = ZERO_RATING_SUBJECT_PREFIX + "1s"
	}
	if !strings.HasPrefix(rateSubj, ZERO_RATING_SUBJECT_PREFIX) {
		return 0, errors.New("malformed rating subject: " + rateSubj)
	}
	durStr := rateSubj[len(ZERO_RATING_SUBJECT_PREFIX):]
	return time.ParseDuration(durStr)
}

func ConcatenatedKey(keyVals ...string) string {
	return strings.Join(keyVals, CONCATENATED_KEY_SEP)
}

func LCRKey(direction, tenant, category, account, subject string) string {
	return ConcatenatedKey(direction, tenant, category, account, subject)

}

func RatingSubjectAliasKey(tenant, subject string) string {
	return ConcatenatedKey(tenant, subject)
}

func AccountAliasKey(tenant, account string) string {
	return ConcatenatedKey(tenant, account)
}

func InfieldJoin(vals ...string) string {
	return strings.Join(vals, INFIELD_SEP)
}

func InfieldSplit(val string) []string {
	return strings.Split(val, INFIELD_SEP)
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// successive Fibonacci numbers.
func Fib() func() time.Duration {
	a, b := 0, 1
	return func() time.Duration {
		a, b = b, a+b
		return time.Duration(a) * time.Second
	}
}

// Utilities to provide pointers where we need to define ad-hoc
func StringPointer(str string) *string {
	if str == ZERO {
		str = ""
		return &str
	}
	return &str
}

func IntPointer(i int) *int {
	return &i
}

func Int64Pointer(i int64) *int64 {
	return &i
}

func Float64Pointer(f float64) *float64 {
	return &f
}

func BoolPointer(b bool) *bool {
	return &b
}

func StringSlicePointer(slc []string) *[]string {
	return &slc
}

func Float64SlicePointer(slc []float64) *[]float64 {
	return &slc
}

func StringMapPointer(sm StringMap) *StringMap {
	return &sm
}

func TimePointer(t time.Time) *time.Time {
	return &t
}

func ReflectFuncLocation(handler interface{}) (file string, line int) {
	f := runtime.FuncForPC(reflect.ValueOf(handler).Pointer())
	entry := f.Entry()
	return f.FileLine(entry)
}

func ToIJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", " ")
	return string(b)
}

func ToJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func LogFull(v interface{}) {
	log.Print(ToIJSON(v))
}

// Used to convert from generic interface type towards string value
func ConvertIfaceToString(fld interface{}) (string, bool) {
	var strVal string
	var converted bool
	switch fld.(type) {
	case string:
		strVal = fld.(string)
		converted = true
	case int:
		strVal = strconv.Itoa(fld.(int))
		converted = true
	case int64:
		strVal = strconv.FormatInt(fld.(int64), 10)
		converted = true
	case bool:
		strVal = strconv.FormatBool(fld.(bool))
		converted = true
	case []uint8:
		var byteVal []byte
		if byteVal, converted = fld.([]byte); converted {
			strVal = string(byteVal)
		}
	default: // Maybe we are lucky and the value converts to string
		strVal, converted = fld.(string)
	}
	return strVal, converted
}

// Simple object cloner, b should be a pointer towards a value into which we want to decode
func Clone(a, b interface{}) error {
	buff := new(bytes.Buffer)
	enc := gob.NewEncoder(buff)
	dec := gob.NewDecoder(buff)
	if err := enc.Encode(a); err != nil {
		return err
	}
	if err := dec.Decode(b); err != nil {
		return err
	}
	return nil
}

// Used as generic function logic for various fields

// Attributes
//  source - the base source
//  width - the field width
//  strip - if present it will specify the strip strategy, when missing strip will not be allowed
//  padding - if present it will specify the padding strategy to use, left, right, zeroleft, zeroright
func FmtFieldWidth(source string, width int, strip, padding string, mandatory bool) (string, error) {
	if mandatory && len(source) == 0 {
		return "", errors.New("Empty source value")
	}
	if width == 0 { // Disable width processing if not defined
		return source, nil
	}
	if len(source) == width { // the source is exactly the maximum length
		return source, nil
	}
	if len(source) > width { //the source is bigger than allowed
		if len(strip) == 0 {
			return "", fmt.Errorf("Source %s is bigger than the width %d, no strip defied", source, width)
		}
		if strip == "right" {
			return source[:width], nil
		} else if strip == "xright" {
			return source[:width-1] + "x", nil // Suffix with x to mark prefix
		} else if strip == "left" {
			diffIndx := len(source) - width
			return source[diffIndx:], nil
		} else if strip == "xleft" { // Prefix one x to mark stripping
			diffIndx := len(source) - width
			return "x" + source[diffIndx+1:], nil
		}
	} else { //the source is smaller as the maximum allowed
		if len(padding) == 0 {
			return "", fmt.Errorf("Source %s is smaller than the width %d, no padding defined", source, width)
		}
		var paddingFmt string
		switch padding {
		case "right":
			paddingFmt = fmt.Sprintf("%%-%ds", width)
		case "left":
			paddingFmt = fmt.Sprintf("%%%ds", width)
		case "zeroleft":
			paddingFmt = fmt.Sprintf("%%0%ds", width)
		}
		if len(paddingFmt) != 0 {
			return fmt.Sprintf(paddingFmt, source), nil
		}
	}
	return source, nil
}

// Returns the string representation of iface or error if not convertible
func CastIfToString(iface interface{}) (strVal string, casts bool) {
	switch iface.(type) {
	case string:
		strVal = iface.(string)
		casts = true
	case int:
		strVal = strconv.Itoa(iface.(int))
		casts = true
	case int64:
		strVal = strconv.FormatInt(iface.(int64), 10)
		casts = true
	case float64:
		strVal = strconv.FormatFloat(iface.(float64), 'f', -1, 64)
		casts = true
	case bool:
		strVal = strconv.FormatBool(iface.(bool))
		casts = true
	case []uint8:
		var byteVal []byte
		if byteVal, casts = iface.([]byte); casts {
			strVal = string(byteVal)
		}
	default: // Maybe we are lucky and the value converts to string
		strVal, casts = iface.(string)
	}
	return strVal, casts
}

func GetEndOfMonth(ref time.Time) time.Time {
	if ref.IsZero() {
		return time.Now()
	}
	year, month, _ := ref.Date()
	if month == time.December {
		year++
		month = time.January
	} else {
		month++
	}
	eom := time.Date(year, month, 1, 0, 0, 0, 0, ref.Location())
	return eom.Add(-time.Second)
}
