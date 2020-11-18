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

package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// NewRjReader creates a new rjReader from a io.Reader
func NewRjReader(rdr io.Reader) (r *rjReader, err error) {
	var b []byte
	b, err = ioutil.ReadAll(rdr)
	if err != nil {
		return
	}
	return NewRjReaderFromBytes(b), nil
}

// NewRjReaderFromBytes creates a new rjReader from a slice of bytes
func NewRjReaderFromBytes(b []byte) *rjReader {
	return &rjReader{buf: b}
}

// isNewLine check if byte is new line
func isNewLine(c byte) bool {
	return c == '\n' || c == '\r'
}

// isWhiteSpace check if byte is white space
func isWhiteSpace(c byte) bool {
	return c == ' ' || c == '\t' || isNewLine(c) || c == 0
}

//ReadEnv reads the enviorment variable
func ReadEnv(key string) (string, error) { //it shod print a warning not a error
	if env := os.Getenv(key); env != "" {
		return env, nil
	}
	return "", utils.ErrEnvNotFound(key)
}

// isAlfanum check if byte is number or letter
func isAlfanum(bit byte) bool {
	return (bit >= 'a' && bit <= 'z') ||
		(bit >= 'A' && bit <= 'Z') ||
		(bit >= '0' && bit <= '9')
}

// structure that implements io.Reader to read json files ignoring C style comments and replacing *env:
type rjReader struct {
	buf        []byte
	isInString bool // ignore character in strings
	indx       int  // used to parse the buffer
	envOff     bool
}

// Read implementation
func (rjr *rjReader) Read(p []byte) (n int, err error) {
	for n = range p {
		p[n], err = rjr.ReadByte()
		if !rjr.envOff &&
			p[n] == '*' &&
			rjr.checkMeta() {
			if err = rjr.replaceEnv(rjr.indx - 1); err != nil {
				return
			}
			p[n] = rjr.buf[rjr.indx-1] // replace with first value
		}
		if err != nil {
			return
		}
	}
	n++ //because it starts from 0
	return
}

// Close implementation
func (rjr *rjReader) Close() error {
	rjr.buf = nil
	return nil
}

// ReadByte implementation
func (rjr *rjReader) ReadByte() (bit byte, err error) {
	if rjr.isInString { //ignore commas in strings
		return rjr.ReadByteWC()
	}
	bit, err = rjr.ReadByteWC()
	if err != nil {
		return
	}
	if bit == utils.CSV_SEP {
		var bit2 byte
		bit2, err = rjr.PeekByteWC()
		if err != nil {
			return
		}
		if bit2 == ']' || bit2 == '}' {
			return rjr.ReadByteWC()
		}
	}
	return
}

func (rjr *rjReader) UnreadByte() (err error) {
	if rjr.indx <= 0 {
		return bufio.ErrInvalidUnreadByte
	}
	rjr.indx--
	return
}

// returns true if the file was parsed completly
func (rjr *rjReader) isEndOfFile() bool {
	return rjr.indx >= len(rjr.buf)
}

// consumeComent consumes the comment based on the peeked byte
func (rjr *rjReader) consumeComent(pkbit byte) (bool, error) {
	switch pkbit {
	case '/':
		for !rjr.isEndOfFile() {
			bit := rjr.buf[rjr.indx]
			rjr.indx++
			if isNewLine(bit) { //read until newline or EOF
				return true, nil
			}
		}
		return true, io.EOF
	case '*':
		rjr.indx += 3 // increase to ignore peeked bytes
		for !rjr.isEndOfFile() {
			if rjr.buf[rjr.indx] == '/' &&
				rjr.buf[rjr.indx-1] == '*' {
				rjr.indx++
				return true, nil
			}
			rjr.indx++
		}
		return true, utils.ErrJsonIncompleteComment
	}
	return false, nil
}

//readFirstNonWhiteSpace reads first non white space byte
func (rjr *rjReader) readFirstNonWhiteSpace() (bit byte, err error) {
	for !rjr.isEndOfFile() {
		bit = rjr.buf[rjr.indx]
		rjr.indx++
		if !isWhiteSpace(bit) {
			return
		}
	}
	return 0, io.EOF
}

// ReadByteWC reads next byte skiping the comments
func (rjr *rjReader) ReadByteWC() (bit byte, err error) {
	if rjr.isEndOfFile() {
		return 0, io.EOF
	}
	if rjr.isInString {
		bit = rjr.buf[rjr.indx]
		rjr.indx++
	} else if bit, err = rjr.readFirstNonWhiteSpace(); err != nil {
		return
	}
	if bit == '"' {
		rjr.isInString = !rjr.isInString
		return
	}
	if !rjr.isInString && !rjr.isEndOfFile() && bit == '/' {
		var pkbit byte
		var isComment bool
		pkbit = rjr.buf[rjr.indx]
		isComment, err = rjr.consumeComent(pkbit)
		if err != nil && err != io.EOF {
			return
		}
		if isComment {
			return rjr.ReadByteWC()
		}
	}
	return
}

// PeekByteWC peeks next byte skiping the comments
func (rjr *rjReader) PeekByteWC() (bit byte, err error) {
	for !rjr.isEndOfFile() {
		bit = rjr.buf[rjr.indx]
		if !rjr.isInString && rjr.indx+1 < len(rjr.buf) && bit == '/' { //try consume comment
			var pkbit byte
			var isComment bool
			pkbit = rjr.buf[rjr.indx+1]
			isComment, err = rjr.consumeComent(pkbit)
			if err != nil {
				return
			}
			if isComment {
				return rjr.PeekByteWC()
			}
			return
		}
		if !isWhiteSpace(bit) {
			return
		}
		rjr.indx++
	}
	return 0, io.EOF
}

//checkMeta check if char mach with next char from MetaEnv if not reset the counting
func (rjr *rjReader) checkMeta() bool {
	return rjr.indx-1+len(utils.MetaEnv) < len(rjr.buf) &&
		utils.MetaEnv == string(rjr.buf[rjr.indx-1:rjr.indx-1+len(utils.MetaEnv)])
}

//readEnvName reads the enviorment key
func (rjr *rjReader) readEnvName(indx int) (name []byte, endindx int) { //0 if not set
	for indx < len(rjr.buf) { //read byte by byte
		bit := rjr.buf[indx]
		if !isAlfanum(bit) && bit != '_' { //[a-zA-Z_]+[a-zA-Z0-9_]*
			return name, indx
		}
		name = append(name, bit)
		indx++
	}
	return name, indx
}

//replaceEnv replaces the EnvMeta and enviorment key with  enviorment variable value in specific buffer
func (rjr *rjReader) replaceEnv(startEnv int) error {
	midEnv := len(utils.MetaEnv)
	key, endEnv := rjr.readEnvName(startEnv + midEnv)
	value, err := ReadEnv(string(key))
	if err != nil {
		return err
	}
	rjr.buf = append(rjr.buf[:startEnv], append([]byte(value), rjr.buf[endEnv:]...)...) // replace *env:ENV_STUFF with ENV_VALUE
	return nil
}

// warning: needs to read file again
func (rjr *rjReader) HandleJSONError(err error) error {
	var offset int64
	switch realErr := err.(type) {
	case nil:
		return nil
	case *json.InvalidUTF8Error, *json.UnmarshalFieldError: // deprecated
		return err
	case *json.InvalidUnmarshalError: // e.g. nil parameter
		return err
	case *json.SyntaxError:
		offset = realErr.Offset
	case *json.UnmarshalTypeError:
		offset = realErr.Offset
	default:
		return err
	}
	if offset == 0 {
		return fmt.Errorf("%s at line 0 around position 0", err.Error())
	}
	rjr.indx = 0

	line, character := rjr.getJSONOffsetLine(offset)
	return fmt.Errorf("%s around line %v and position %v\n line: %q", err.Error(), line, character,
		strings.Split(string(rjr.buf), "\n")[line-1])
}

func (rjr *rjReader) getJSONOffsetLine(offset int64) (line, character int64) {
	line = 1 // start line counting from 1
	var lastChar byte

	var i int64 = 0
	readString := func() error {
		for i < offset {
			if rjr.isEndOfFile() {
				return io.EOF
			}
			b := rjr.buf[rjr.indx]
			rjr.indx++
			i++
			if isNewLine(b) {
				line++
				character = 0
			} else {
				character++
			}
			if b == '"' {
				return nil
			}
		}
		return nil
	}
	readLineComment := func() error {
		for i < offset {
			if rjr.isEndOfFile() {
				return io.EOF
			}
			b := rjr.buf[rjr.indx]
			rjr.indx++
			if isNewLine(b) {
				line++
				character = 0
				return nil
			}
			character++
		}
		return nil
	}

	readComment := func() error {
		for i < offset {
			if rjr.isEndOfFile() {
				return io.EOF
			}
			b := rjr.buf[rjr.indx]
			rjr.indx++
			if isNewLine(b) {
				line++
				character = 0
			} else {
				character++
			}
			if b == '*' {
				if rjr.isEndOfFile() {
					return io.EOF
				}
				b := rjr.buf[rjr.indx]
				rjr.indx++
				if b == '/' {
					character++
					return nil
				}
				// Unread byte return error only if rjr.indx is small or equal to 0
				// The value is greater than 0 so it's safe to not check the error
				rjr.UnreadByte()
			}
		}
		return nil
	}

	for i < offset { // handle the parsing
		if rjr.isEndOfFile() {
			return
		}
		b := rjr.buf[rjr.indx]
		rjr.indx++
		character++
		if isNewLine(b) {
			line++
			character = 0
		}
		if (b == ']' || b == '}') && lastChar == utils.CSV_SEP {
			i-- //ignore utils.CSV_SEP if is followed by ] or }
		}
		if !isWhiteSpace(b) {
			i++
			lastChar = b
		}
		if b == '"' { // read "" value
			rerr := readString()
			if rerr != nil {
				break
			}
		}
		if b == '/' {
			if rjr.isEndOfFile() {
				return
			}
			b := rjr.buf[rjr.indx]
			rjr.indx++
			if b == '/' { // read //
				character++
				i--
				rerr := readLineComment()
				if rerr != nil {
					break
				}
			} else if b == '*' { // read /*
				character++
				i--
				rerr := readComment()
				if rerr != nil {
					break
				}
			} else {
				// Unread byte return error only if rjr.indx is small or equal to 0
				// The value is greater than 0 so it's safe to not check the error
				rjr.UnreadByte()
			}
		}
	}
	return
}

// Loads the json config out of rjReader
func (rjr *rjReader) Decode(cfg interface{}) (err error) {
	if err = json.NewDecoder(rjr).Decode(cfg); err != nil {
		return rjr.HandleJSONError(err)
	}
	return
}
