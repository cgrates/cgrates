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

package utils

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	math_rand "math/rand"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	boolGenerator *boolGen

	rfc3339Rule                  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.+$`)
	sqlRule                      = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}$`)
	utcFormat                    = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[T]\d{2}:\d{2}:\d{2}$`)
	gotimeRule                   = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}\.?\d*\s[+,-]\d+\s\w+$`)
	gotimeRule2                  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}\.?\d*\s[+,-]\d+\s[+,-]\d+$`)
	fsTimestamp                  = regexp.MustCompile(`^\d{16}$`)
	astTimestamp                 = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d*[+,-]\d+$`)
	unixTimestampRule            = regexp.MustCompile(`^\d{10}$`)
	unixTimestampMilisecondsRule = regexp.MustCompile(`^\d{13}$`)
	unixTimestampNanosecondsRule = regexp.MustCompile(`^\d{19}$`)
	oneLineTimestampRule         = regexp.MustCompile(`^\d{14}$`)
	oneSpaceTimestampRule        = regexp.MustCompile(`^\d{2}\.\d{2}.\d{4}\s{1}\d{2}:\d{2}:\d{2}$`)
	eamonTimestampRule           = regexp.MustCompile(`^\d{2}/\d{2}/\d{4}\s{1}\d{2}:\d{2}:\d{2}$`)
	broadsoftTimestampRule       = regexp.MustCompile(`^\d{14}\.\d{3}`)
)

func init() {
	boolGenerator = newBoolGen()
}

// BoolGenerator return the boolean generator
func BoolGenerator() *boolGen {
	return boolGenerator
}

func NewCounter(start, limit int64) *Counter {
	return &Counter{
		value: start,
		limit: limit,
	}
}

type Counter struct {
	value, limit int64
	sync.Mutex
}

func (c *Counter) Next() int64 {
	c.Lock()
	defer c.Unlock()
	c.value += 1
	if c.limit > 0 && c.value > c.limit {
		c.value = 0
	}
	return c.value
}

func (c *Counter) Value() int64 {
	c.Lock()
	defer c.Unlock()
	return c.value
}

// Returns first non empty string out of vals. Useful to extract defaults
func FirstNonEmpty(vals ...string) string {
	for _, val := range vals {
		if len(val) != 0 {
			return val
		}
	}
	return EmptyString
}

func FirstIntNonEmpty(vals ...int) int {
	for _, val := range vals {
		if val != 0 {
			return val
		}
	}
	return 0
}

func FirstDurationNonEmpty(vals ...time.Duration) time.Duration {
	for _, val := range vals {
		if val != 0 {
			return val
		}
	}
	return 0
}

// Sha1 generate the SHA1 hash from any string
// the order of string matters
func Sha1(attrs ...string) string {
	hasher := sha1.New()
	for _, attr := range attrs {
		hasher.Write([]byte(attr))
	}
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// helper function for uuid generation
func GenUUID() string {
	b := make([]byte, 16)
	io.ReadFull(rand.Reader, b)
	b[6] = (b[6] & 0x0F) | 0x40
	b[8] = (b[8] &^ 0x40) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10],
		b[10:])
}

// UUIDSha1Prefix generates a prefix of the sha1 applied to an UUID
// prefix 8 is chosen since the probability of colision starts being minimal after 7 characters (see git commits)
func UUIDSha1Prefix() string {
	return Sha1(GenUUID())[:7]
}

// Round return rounded version of x with prec precision.
//
// Special cases are:
//
//	Round(±0) = ±0
//	Round(±Inf) = ±Inf
//	Round(NaN) = NaN
func Round(x float64, prec int, method string) float64 {
	var rounder float64
	maxPrec := 7 // define a max precision to cut float errors
	if maxPrec < prec {
		maxPrec = prec
	}
	pow := math.Pow(10, float64(prec))
	intermed := x * pow
	_, frac := math.Modf(intermed)

	switch method {
	case MetaRoundingUp:
		if frac >= math.Pow10(-maxPrec) { // Max precision we go, rest is float chaos
			rounder = math.Ceil(intermed)
		} else {
			rounder = math.Floor(intermed)
		}
	case MetaRoundingDown:
		rounder = math.Floor(intermed)
	case MetaRoundingMiddle:
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

// RoundStatDuration is used in engine package for stat metrics that has duration (e.g acd metric, tcd metric, etc...)
func RoundStatDuration(x time.Duration, prec int) time.Duration {
	return x.Round(time.Duration(math.Pow10(9 - prec)))
}

func getAddDuration(tmStr string) (addDur time.Duration, err error) {
	eDurIdx := strings.Index(tmStr, "+")
	if eDurIdx == -1 {
		return
	}
	return time.ParseDuration(tmStr[eDurIdx+1:])
}

// ParseTimeDetectLayout returns the time from string
func ParseTimeDetectLayout(tmStr string, timezone string) (time.Time, error) {
	tmStr = strings.TrimSpace(tmStr)
	var nilTime time.Time
	if len(tmStr) == 0 || tmStr == MetaUnlimited {
		return nilTime, nil
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nilTime, err
	}
	switch {
	case tmStr == MetaUnlimited || tmStr == "":
	// leave it at zero
	case tmStr == MetaDaily:
		return time.Now().AddDate(0, 0, 1), nil // add one day
	case tmStr == MetaMonthly:
		return time.Now().AddDate(0, 1, 0), nil // add one month
	case tmStr == MetaMonthlyEstimated:
		return monthlyEstimated(time.Now())
	case tmStr == MetaYearly:
		return time.Now().AddDate(1, 0, 0), nil // add one year

	case strings.HasPrefix(tmStr, "*month_end"):
		expDate := GetEndOfMonth(time.Now())
		extraDur, err := getAddDuration(tmStr)
		if err != nil {
			return nilTime, err
		}
		expDate = expDate.Add(extraDur)
		return expDate, nil
	case strings.HasPrefix(tmStr, "*mo"): // add one month and extra duration
		extraDur, err := getAddDuration(tmStr)
		if err != nil {
			return nilTime, err
		}
		return time.Now().AddDate(0, 1, 0).Add(extraDur), nil
	case astTimestamp.MatchString(tmStr):
		return time.Parse("2006-01-02T15:04:05.999999999-0700", tmStr)
	case rfc3339Rule.MatchString(tmStr):
		return time.Parse(time.RFC3339, tmStr)
	case gotimeRule.MatchString(tmStr):
		return time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", tmStr)
	case gotimeRule2.MatchString(tmStr):
		return time.Parse("2006-01-02 15:04:05.999999999 -0700 -0700", tmStr)
	case sqlRule.MatchString(tmStr):
		return time.ParseInLocation("2006-01-02 15:04:05", tmStr, loc)
	case fsTimestamp.MatchString(tmStr):
		if tmstmp, err := strconv.ParseInt(tmStr+"000", 10, 64); err != nil {
			return nilTime, err
		} else {
			return time.Unix(0, tmstmp).In(loc), nil
		}
	case unixTimestampRule.MatchString(tmStr):
		//error never happens because of regex
		tmstmp, _ := strconv.ParseInt(tmStr, 10, 64)
		return time.Unix(tmstmp, 0).In(loc), nil
	case unixTimestampMilisecondsRule.MatchString(tmStr):
		//error never happens because of regex
		tmstmp, _ := strconv.ParseInt(tmStr, 10, 64)
		return time.Unix(0, tmstmp*int64(time.Millisecond)).In(loc), nil

	case unixTimestampNanosecondsRule.MatchString(tmStr):
		if tmstmp, err := strconv.ParseInt(tmStr, 10, 64); err != nil {
			return nilTime, err
		} else {
			return time.Unix(0, tmstmp).In(loc), nil
		}
	case tmStr == "0" || len(tmStr) == 0: // Time probably missing from request
		return nilTime, nil
	case oneLineTimestampRule.MatchString(tmStr):
		return time.ParseInLocation("20060102150405", tmStr, loc)
	case oneSpaceTimestampRule.MatchString(tmStr):
		return time.ParseInLocation("02.01.2006  15:04:05", tmStr, loc)
	case eamonTimestampRule.MatchString(tmStr):
		return time.ParseInLocation("02/01/2006 15:04:05", tmStr, loc)
	case broadsoftTimestampRule.MatchString(tmStr):
		return time.ParseInLocation("20060102150405.999", tmStr, loc)
	case tmStr == MetaNow:
		return time.Now(), nil
	case strings.HasPrefix(tmStr, "+"):
		tmStr = strings.TrimPrefix(tmStr, "+")
		if tmStrTmp, err := time.ParseDuration(tmStr); err != nil {
			return nilTime, err
		} else {
			return time.Now().Add(tmStrTmp), nil
		}
	case strings.HasPrefix(tmStr, "-"):
		tmStr = strings.TrimPrefix(tmStr, "-")
		if tmStrTmp, err := time.ParseDuration(tmStr); err != nil {
			return nilTime, err
		} else {
			return time.Now().Add(-tmStrTmp), nil
		}
	case utcFormat.MatchString(tmStr):
		return time.ParseInLocation("2006-01-02T15:04:05", tmStr, loc)

	}
	return nilTime, errors.New("Unsupported time format")
}

// monthlyEstimated calculates the next month's estimated end date by adjusting for month-end variations
// and handling edge cases like different month lengths and leap years
func monthlyEstimated(t1 time.Time) (time.Time, error) {
	initialMnt := t1.Month()
	tAfter := t1.AddDate(0, 1, 0)
	for tAfter.Month()-initialMnt > 1 {
		tAfter = tAfter.AddDate(0, 0, -1)
	}
	return tAfter, nil
}

// RoundDuration returns a number equal or larger than the amount that exactly
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

func SplitSuffix(suffix string) []string {
	length := len(suffix)
	subs := make([]string, length)
	max := len(suffix) - 1
	for i := 0; i < length; i++ {
		subs[i] = suffix[max-i:]
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
func ParseDurationWithSecs(durStr string) (d time.Duration, err error) {
	if durStr == "" {
		return
	}
	if _, err = strconv.ParseFloat(durStr, 64); err == nil { // Seconds format considered
		durStr += "s"
	}
	return time.ParseDuration(durStr)
}

// Parses duration, considers s as time unit if not provided, seconds as float to specify subunits
func ParseDurationWithNanosecs(durStr string) (d time.Duration, err error) {
	if durStr == "" {
		return
	}
	if durStr == MetaUnlimited {
		durStr = "-1"
	}
	if _, err = strconv.ParseFloat(durStr, 64); err == nil { // Seconds format considered
		durStr += "ns"
	}
	return time.ParseDuration(durStr)
}

// returns the minimum duration between the two
func MinDuration(d1, d2 time.Duration) time.Duration {
	if d1 < d2 {
		return d1
	}
	return d2
}

// ParseZeroRatingSubject will parse the subject in the balance
// returns duration if able to extract it from subject
// returns error if not able to parse duration (ie: if ratingSubject is standard one)
func ParseZeroRatingSubject(tor, rateSubj string, defaultRateSubj map[string]string, isUnitBal bool) (time.Duration, error) {
	rateSubj = strings.TrimSpace(rateSubj)
	if !isUnitBal && rateSubj == EmptyString {
		return 0, errors.New("no rating subject for monetary")
	}
	if rateSubj == EmptyString ||
		rateSubj == MetaAny {
		var hasToR bool
		if rateSubj, hasToR = defaultRateSubj[tor]; !hasToR {
			rateSubj = defaultRateSubj[MetaAny]
		}
	}
	if !strings.HasPrefix(rateSubj, MetaRatingSubjectPrefix) {
		return 0, errors.New("malformed rating subject: " + rateSubj)
	}
	durStr := rateSubj[len(MetaRatingSubjectPrefix):]
	if val, err := strconv.ParseFloat(durStr, 64); err == nil { // No time unit, postpend
		return time.Duration(val), nil // just return the float value converted(this should be faster than reparsing the string)
	}
	return time.ParseDuration(durStr)
}

func ConcatenatedKey(keyVals ...string) string {
	return strings.Join(keyVals, ConcatenatedKeySep)
}

func SplitConcatenatedKey(key string) []string {
	return strings.Split(key, ConcatenatedKeySep)
}

func InfieldJoin(vals ...string) string {
	return strings.Join(vals, InfieldSep)
}

func InfieldSplit(val string) []string {
	return strings.Split(val, InfieldSep)
}

// Splited Unzip in small functions to have better coverage
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
			continue
		}
		err = unzipFile(f, path, f.Mode())
		if err != nil {
			return err
		}
	}
	return err
}

type zipFile interface {
	Open() (io.ReadCloser, error)
}

func unzipFile(f zipFile, path string, fm os.FileMode) (err error) {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	err = copyFile(rc, path, fm)
	rc.Close()
	if err != nil {
		return err
	}
	return nil
}

func copyFile(rc io.ReadCloser, path string, fm os.FileMode) (err error) {
	f, err := os.OpenFile(
		path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, rc)
	return
}

// Fib returns successive Fibonacci numbers.
func Fib() func() int {
	a, b := 0, 1
	return func() int {
		a, b = b, a+b

		// Prevent int overflow by keeping b as the maximum valid Fibonacci number.
		if b < a {
			b = a
		}
		return a
	}
}

// FibDuration returns successive Fibonacci numbers as time.Duration with the
// unit specified by durationUnit or maxDuration if it is exceeded
func FibDuration(durationUnit, maxDuration time.Duration) func() time.Duration {
	fib := Fib()
	return func() time.Duration {
		fibNrAsDuration := time.Duration(fib())

		// Handle potential overflow when multiplying by durationUnit.
		if fibNrAsDuration > (math.MaxInt / durationUnit) {
			fibNrAsDuration = math.MaxInt
		} else {
			fibNrAsDuration *= durationUnit
		}

		// Cap the duration to maxDuration if specified.
		if maxDuration > 0 && maxDuration < fibNrAsDuration {
			return maxDuration
		}
		return fibNrAsDuration
	}
}

// Utilities to provide pointers where we need to define ad-hoc
func StringPointer(str string) *string {
	if str == MetaZero {
		str = EmptyString
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

func StringMapPointer(sm StringMap) *StringMap {
	return &sm
}

func MapStringStringPointer(mp map[string]string) *map[string]string {
	return &mp
}
func MapStringSlicePointer(mp map[string][]string) *map[string][]string {
	return &mp
}

func TimePointer(t time.Time) *time.Time {
	return &t
}

func DurationPointer(d time.Duration) *time.Duration {
	return &d
}

func SliceStringPointer(d []string) *[]string {
	return &d
}

func ToIJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", " ")
	return string(b)
}

func ToJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// Simple object cloner, b should be a pointer towards a value into which we want to decode
func Clone(a, b any) error {
	buff := new(bytes.Buffer)
	enc := gob.NewEncoder(buff)
	dec := gob.NewDecoder(buff)
	if err := enc.Encode(a); err != nil {
		return err
	}
	return dec.Decode(b)
}

// Used as generic function logic for various fields

// Attributes
//
//	source - the base source
//	width - the field width
//	strip - if present it will specify the strip strategy, when missing strip will not be allowed
//	padding - if present it will specify the padding strategy to use, left, right, zeroleft, zeroright
func FmtFieldWidth(fieldID, source string, width int, strip, padding string, mandatory bool) (string, error) {
	if mandatory && len(source) == 0 {
		return "", fmt.Errorf("Empty source value for fieldID: <%s>", fieldID)
	}
	if width == 0 { // Disable width processing if not defined
		return source, nil
	}
	if len(source) == width { // the source is exactly the maximum length
		return source, nil
	}
	if len(source) > width { //the source is bigger than allowed
		if len(strip) == 0 {
			return "", fmt.Errorf("Source %s is bigger than the width %d, no strip defied, fieldID: <%s>", source, width, fieldID)
		}
		if strip == MetaRight {
			return source[:width], nil
		} else if strip == MetaXRight {
			return source[:width-1] + "x", nil // Suffix with x to mark prefix
		} else if strip == MetaLeft {
			diffIndx := len(source) - width
			return source[diffIndx:], nil
		} else if strip == MetaXLeft { // Prefix one x to mark stripping
			diffIndx := len(source) - width
			return "x" + source[diffIndx+1:], nil
		}
	} else { //the source is smaller as the maximum allowed
		if len(padding) == 0 {
			return "", fmt.Errorf("Source %s is smaller than the width %d, no padding defined, fieldID: <%s>", source, width, fieldID)
		}
		var paddingFmt string
		switch padding {
		case MetaRight:
			paddingFmt = fmt.Sprintf("%%-%ds", width)
		case MetaLeft:
			paddingFmt = fmt.Sprintf("%%%ds", width)
		case MetaZeroLeft:
			paddingFmt = fmt.Sprintf("%%0%ds", width)
		}
		if len(paddingFmt) != 0 {
			return fmt.Sprintf(paddingFmt, source), nil
		}
	}
	return source, nil
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

// formats number in K,M,G, etc.
func SizeFmt(num float64, suffix string) string {
	if suffix == EmptyString {
		suffix = "B"
	}
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(num) < 1024.0 {
			return fmt.Sprintf("%3.1f%s%s", num, unit, suffix)
		}
		num /= 1024.0
	}
	return fmt.Sprintf("%.1f%s%s", num, "Yi", suffix)
}

func TimeIs0h(t time.Time) bool {
	return t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0
}

func ParseHierarchyPath(path string, sep string) HierarchyPath {
	if path == EmptyString {
		return nil
	}
	if sep == EmptyString {
		for _, sep = range []string{"/", NestingSep} {
			if strings.Contains(path, sep) {
				break
			}
		}
	}
	path = strings.Trim(path, sep) // Need to strip if prefix of suffiy (eg: paths with /) so we can properly split
	return HierarchyPath(strings.Split(path, sep))
}

// HierarchyPath is used in various places to represent various path hierarchies (eg: in Diameter groups, XML trees)
type HierarchyPath []string

// AsString converts HierarchyPath to a string.
func (hP HierarchyPath) AsString(sep string, isAbsolute bool) string {
	var strHP strings.Builder

	// If isAbsolute is true and the HierarchyPath slice is empty, sep will be returned.
	// This will indicate the start of the absolute path.
	if isAbsolute {
		strHP.WriteString(sep)
	}

	if len(hP) == 0 {
		// If isAbsolute is false and HierarchyPath is empty, return '.' to represent the current directory in a relative path.
		// This convention avoids errors (e.g., "expr expression is nil") when retrieving elements from an empty path.
		if !isAbsolute {
			return "."
		}
		return strHP.String()
	}

	for i, elem := range hP {
		if i != 0 {
			strHP.WriteString(sep)
		}
		strHP.WriteString(elem)
	}
	return strHP.String()
}

// Clone returns a deep copy of HierarchyPath
func (h HierarchyPath) Clone() (cln HierarchyPath) {
	if h == nil {
		return
	}
	return slices.Clone(h)
}

// Mask a number of characters in the suffix of the destination
func MaskSuffix(dest string, maskLen int) string {
	destLen := len(dest)
	if maskLen < 0 {
		return dest
	} else if maskLen > destLen {
		maskLen = destLen
	}
	dest = dest[:destLen-maskLen]
	for i := 0; i < maskLen; i++ {
		dest += MaskChar
	}
	return dest
}

// Sortable Int64Slice
type Int64Slice []int64

func (slc Int64Slice) Len() int {
	return len(slc)
}
func (slc Int64Slice) Swap(i, j int) {
	slc[i], slc[j] = slc[j], slc[i]
}
func (slc Int64Slice) Less(i, j int) bool {
	return slc[i] < slc[j]
}

func GetCGRVersion() (vers string, err error) {
	vers = fmt.Sprintf("%s@%s", CGRateS, Version)
	if GitCommitDate == "" || GitCommitHash == "" {
		return vers, nil
	}
	var matched bool

	/*
		Git v2.45 relevant release note:
		 * The output format for dates "iso-strict" has been tweaked to show
		   a time in the Zulu timezone with "Z" suffix, instead of "+00:00".
	*/

	// Parse the git commit date, which might be in different formats depending on the git version
	trimmedCommitDate := strings.TrimSpace(GitCommitDate)
	commitDate, err := time.Parse("2006-01-02T15:04:05Z", trimmedCommitDate)
	if err != nil {
		// Failed to parse iso-strict date format for git version 2.45+. Try to parse with the previous format.
		var fallbackErr error
		commitDate, fallbackErr = time.Parse("2006-01-02T15:04:05-07:00", trimmedCommitDate)
		if fallbackErr != nil {
			// Both parsing attempts failed, group the errors together.
			err = fmt.Errorf(
				"failed to parse date:\ngit2.45+ iso-strict format: %w\nprevious iso-strict format: %w",
				err, fallbackErr)
		} else {
			err = nil // successfully parsed with fallback format
		}
	}
	if err != nil && GitCommitDate != "%cI" { // ignore error for old git versions where %cI is not defined
		return vers, fmt.Errorf("version build error: %w", err)
	}
	matched, err = regexp.MatchString("^[0-9a-f]{12,}$", GitCommitHash)
	if err != nil {
		return vers, fmt.Errorf("version build error: commit hash compilation failed: %v", err)
	} else if !matched {
		return vers, fmt.Errorf("version build error: commit hash does not match expected format")
	}
	commitHash := GitCommitHash
	//CGRateS@v0.10.1~dev-20200110075344-7572e7b11e00
	return fmt.Sprintf("%s@%s-%s-%s", CGRateS, Version, commitDate.UTC().Format("20060102150405"), commitHash[:12]), nil
}

// NewTenantID parses a string in the format of "tenant:ID" and returns
// a TenantID struct. If the separator is not found, the entire string
// is treated as the ID.
func NewTenantID(input string) *TenantID {
	tenant, id, sepFound := strings.Cut(input, ConcatenatedKeySep)
	if !sepFound {
		return &TenantID{ID: input}
	}
	return &TenantID{Tenant: tenant, ID: id}
}

type PaginatorWithTenant struct {
	Tenant string
	Paginator
}

type TenantWithAPIOpts struct {
	Tenant  string
	APIOpts map[string]any
}

type TenantID struct {
	Tenant string
	ID     string
}

type TenantIDWithAPIOpts struct {
	*TenantID
	APIOpts map[string]any
}

func (tID *TenantID) TenantID() string {
	return ConcatenatedKey(tID.Tenant, tID.ID)
}

func (tID *TenantIDWithAPIOpts) TenantIDConcatenated() string {
	return ConcatenatedKey(tID.Tenant, tID.ID)
}

// CachedRPCResponse is used to cache a RPC response
type CachedRPCResponse struct {
	Result any
	Error  error
}

func ReverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func GetUrlRawArguments(dialURL string) (out map[string]string) {
	out = make(map[string]string)
	idx := strings.IndexRune(dialURL, '?')
	if idx == -1 {
		return
	}
	strParams := dialURL[idx+1:]
	if len(strParams) == 0 {
		return
	}
	vecParams := strings.Split(strParams, "&")
	for _, paramPair := range vecParams {
		idx := strings.IndexRune(paramPair, '=')
		if idx == -1 {
			continue
		}
		out[paramPair[:idx]] = paramPair[idx+1:]
	}
	return
}

// WarnExecTime is used when we need to meassure the execution of specific functions
// and warn when the total duration is higher than expected
// should be usually called with defer, ie: defer WarnExecTime(time.Now(), "MyTestFunc", 2*time.Second)
func WarnExecTime(startTime time.Time, logID string, maxDur time.Duration) {
	totalDur := time.Since(startTime)
	if totalDur > maxDur {
		Logger.Warning(fmt.Sprintf("<%s> execution took: <%s>", logID, totalDur))
	}
}

// endchan := LongExecTimeDetector("mesaj", 5*time.Second)
// defer func() { close(endchan) }()
func LongExecTimeDetector(logID string, maxDur time.Duration) (endchan chan struct{}) {
	endchan = make(chan struct{}, 1)
	go func() {
		select {
		case <-time.After(maxDur):
			Logger.Warning(fmt.Sprintf("<%s> execution more than: <%s>", logID, maxDur))
		case <-endchan:
		}
	}()
	return
}

type StringWithAPIOpts struct {
	APIOpts map[string]any
	Tenant  string
	Arg     string
}

func CastRPCErr(err error) error {
	if err != nil {
		if _, has := ErrMap[err.Error()]; has {
			return ErrMap[err.Error()]
		}
	}
	return err
}

// RandomInteger returns a random 64-bit integer between min and max values
func RandomInteger(min, max int64) int64 {
	return math_rand.Int63n(max-min) + min
}

type LoadIDsWithAPIOpts struct {
	LoadIDs map[string]int64
	Tenant  string
	APIOpts map[string]any
}

// IsURL returns if the path is an URL
func IsURL(path string) bool {
	return strings.HasPrefix(path, "https://") ||
		strings.HasPrefix(path, "http://")
}

// GetIndexesArg the API argumets to specify an index
type GetIndexesArg struct {
	IdxItmType string
	TntCtx     string
	IdxKey     string
	Tenant     string
	APIOpts    map[string]any
}

// SetIndexesArg the API arguments needed for seting an index
type SetIndexesArg struct {
	IdxItmType string
	TntCtx     string
	Indexes    map[string]StringSet
	Tenant     string
	APIOpts    map[string]any
}

type DurationArgs struct {
	Duration time.Duration
	APIOpts  map[string]any
	Tenant   string
}

type DirectoryArgs struct {
	DirPath string
	APIOpts map[string]any
	Tenant  string
}

// AESEncrypt will encrypt the provided txt using the encKey and AES algorithm
func AESEncrypt(txt, encKey string) (encrypted string, err error) {
	key, _ := hex.DecodeString(encKey)
	var blk cipher.Block
	if blk, err = aes.NewCipher(key); err != nil {
		return
	}
	var aesGCM cipher.AEAD
	aesGCM, _ = cipher.NewGCM(blk)
	nonce := make([]byte, aesGCM.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	return fmt.Sprintf("%x", aesGCM.Seal(nonce, nonce, []byte(txt), nil)), nil
}

// AESDecrypt will decrypt the provided encrypted txt using the encKey and AES algorithm
func AESDecrypt(encrypted string, encKey string) (txt string, err error) {

	key, _ := hex.DecodeString(encKey)
	enc, _ := hex.DecodeString(encrypted)

	var blk cipher.Block
	if blk, err = aes.NewCipher(key); err != nil {
		return
	}

	var aesGCM cipher.AEAD
	aesGCM, _ = cipher.NewGCM(blk)
	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]
	var plaintext []byte
	plaintext, err = aesGCM.Open(nil, nonce, ciphertext, nil)
	return string(plaintext), err
}

// Hash generates the hash text
func ComputeHash(dataKeys ...string) (lns string, err error) {
	var hashByts []byte
	hashByts, err = bcrypt.GenerateFromPassword(
		[]byte(ConcatenatedKey(dataKeys...)),
		bcrypt.MinCost)
	return string(hashByts), err
}

// VerifyHash matches the data hash with the dataKeys ha
func VerifyHash(hash string, dataKeys ...string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash),
		[]byte(ConcatenatedKey(dataKeys...)))
	return err == nil
}

// newBoolGen initialize an efficient boolean generator
func newBoolGen() *boolGen {
	return &boolGen{src: math_rand.NewSource(time.Now().UnixNano())}
}

// boolGen is an efficient boolean generator
type boolGen struct {
	src       math_rand.Source
	cache     int64
	remaining int
}

// RandomBool generate a random boolean
func (b *boolGen) RandomBool() bool {
	if b.remaining == 0 {
		b.cache, b.remaining = b.src.Int63(), 63
	}
	result := b.cache&0x01 == 1
	b.cache >>= 1
	b.remaining--
	return result
}

// GenerateDBItemOpts will create the options for DB replication
// if they are empty they should be omitted
func GenerateDBItemOpts(apiKey, routeID, cache, rmtHost string) (mp map[string]any) {
	mp = make(map[string]any)
	if apiKey != EmptyString {
		mp[OptsAPIKey] = apiKey
	}
	if routeID != EmptyString {
		mp[OptsRouteID] = routeID
	}
	if cache != EmptyString {
		mp[CacheOpt] = cache
	}
	if rmtHost != EmptyString {
		mp[RemoteHostOpt] = rmtHost
	}
	return
}

// SplitPath splits filter rules based on the specified separator
func SplitPath(rule string, sep byte, n int) []string {
	var nest, pos int
	if n <= 0 {
		n = bytes.Count([]byte(rule), []byte{sep}) + 1
	}
	if n == 1 {
		return []string{rule}
	}
	splt := make([]string, 0, n)
	shouldBreak := false
	for i, b := range rule {
		switch byte(b) {
		case sep:
			if nest != 0 {
				continue // skip separators inside nested structures
			}
			splt = append(splt, rule[pos:i])
			pos = i + 1
			if len(splt) == n-1 {
				shouldBreak = true
			}
		case IdxStart[0]:
			nest++ // entering nested structure
		case IdxEnd[0]:
			nest-- // exiting nested structure
		case HashtagSep[0]:
			if sep == NestingSep[0] {
				shouldBreak = true
			}
		}
		if shouldBreak {
			break
		}

	}
	splt = append(splt, rule[pos:]) // add last element
	return splt
}

type PanicMessageArgs struct {
	Tenant  string
	APIOpts map[string]any
	Message string
}

// ParseBinarySize converts string byte sizes (b, kb, mb, gb) to byte int64
func ParseBinarySize(size string) (int64, error) {
	var num float64
	var unit string

	_, err := fmt.Sscanf(size, "%f%s", &num, &unit)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %s", size)
	}

	multipliers := map[string]int64{
		"B":  1,
		"KB": 1 << 10,
		"K":  1 << 10,
		"MB": 1 << 20,
		"M":  1 << 20,
		"GB": 1 << 30,
		"G":  1 << 30,
	}

	if mult, ok := multipliers[strings.ToUpper(unit)]; ok {
		return int64(num * float64(mult)), nil
	}

	return 0, fmt.Errorf("unknown unit: %s", unit)
}
