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
	"io"
	"os"

	"github.com/cgrates/cgrates/utils"
)

// NewRawJSONReader returns a raw JSON reader
func NewRawJSONReader(r io.Reader) io.Reader {
	return &EnvReader{
		rd: &rawJSON{
			rdr: bufio.NewReader(r),
		},
	}
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

// rawJSON is io.ByteReader interface to read JSON without comments, whitespaces and commas before ']' and '}'
type rawJSON struct {
	isInString bool // ignore character in strings
	rdr        *bufio.Reader
}

// ReadByte implementation
func (b *rawJSON) ReadByte() (bit byte, err error) {
	if b.isInString { //ignore commas in strings
		return b.ReadByteWC()
	}
	bit, err = b.ReadByteWC()
	if err != nil {
		return bit, err
	}
	if bit == ',' {
		bit2, err := b.PeekByteWC()
		if err != nil {
			return bit, err
		}
		if bit2 == ']' || bit2 == '}' {
			return b.ReadByteWC()
		}
	}
	return bit, err
}

// consumeComent consumes the comment based on the peeked byte
func (b *rawJSON) consumeComent(pkbit byte) (bool, error) {
	switch pkbit {
	case '/':
		for {
			bit, err := b.rdr.ReadByte()
			if err != nil || isNewLine(bit) { //read until newline or EOF
				return true, err
			}
		}
	case '*':
		for bit, err := b.rdr.ReadByte(); bit != '*'; bit, err = b.rdr.ReadByte() { //max 2 reads
			if err == io.EOF {
				return true, utils.ErrJsonIncompleteComment
			}
			if err != nil {
				return true, err
			}
		}
		simbolMeet := false
		for {
			bit, err := b.rdr.ReadByte()
			if err == io.EOF {
				return true, utils.ErrJsonIncompleteComment
			}
			if err != nil {
				return true, err
			}
			if simbolMeet && bit == '/' {
				return true, nil
			}
			simbolMeet = bit == '*'
		}
	}
	return false, nil
}

//readFirstNonWhiteSpace reads first non white space byte
func (b *rawJSON) readFirstNonWhiteSpace() (byte, error) {
	for {
		bit, err := b.rdr.ReadByte()
		if err != nil || !isWhiteSpace(bit) {
			return bit, err
		}
	}
}

// ReadByteWC reads next byte skiping the comments
func (b *rawJSON) ReadByteWC() (bit byte, err error) {
	if b.isInString {
		bit, err = b.rdr.ReadByte()
	} else {
		bit, err = b.readFirstNonWhiteSpace()
	}
	if err != nil {
		return bit, err
	}
	if bit == '"' {
		b.isInString = !b.isInString
		return bit, err
	}
	if !b.isInString && bit == '/' {
		bit2, err := b.rdr.Peek(1)
		if err != nil {
			return bit, err
		}
		isComment, err := b.consumeComent(bit2[0])
		if err != nil && err != io.EOF {
			return bit, err
		}
		if isComment {
			return b.ReadByteWC()
		}
	}
	return bit, err
}

// PeekByteWC peeks next byte skiping the comments
func (b *rawJSON) PeekByteWC() (byte, error) {
	for {
		bit, err := b.rdr.Peek(1)
		if err != nil {
			return bit[0], err
		}
		if !b.isInString && bit[0] == '/' { //try consume comment
			bit, err = b.rdr.Peek(2)
			if err != nil {
				return bit[0], err
			}
			isComment, err := b.consumeComent(bit[1])
			if err != nil {
				return bit[0], err
			}
			if isComment {
				return b.PeekByteWC()
			}
			return bit[0], err
		}
		if !isWhiteSpace(bit[0]) {
			return bit[0], err
		}
		bit2, err := b.rdr.ReadByte()
		if err != nil {
			return bit2, err
		}
	}
}

// EnvReader io.Reader interface to read JSON replacing the EnvMeta
type EnvReader struct {
	buf []byte
	rd  io.ByteReader // reader provided by the client
	r   int           // buf read positions
	m   int           // meta Ofset used to determine fi MetaEnv was meet
}

//readEnvName reads the enviorment key
func (b *EnvReader) readEnvName() (name []byte, bit byte, err error) { //0 if not set
	for { //read byte by byte
		bit, err := b.rd.ReadByte()
		if err != nil {
			return name, 0, err
		}
		if !isAlfanum(bit) && bit != '_' { //[a-zA-Z_]+[a-zA-Z0-9_]*
			return name, bit, nil
		}
		name = append(name, bit)
	}
}

//replaceEnv replaces the EnvMeta and enviorment key with  enviorment variable value in specific buffer
func (b *EnvReader) replaceEnv(buf []byte, startEnv, midEnv int) (n int, err error) {
	key, bit, err := b.readEnvName()
	if err != nil && err != io.EOF {
		return 0, err
	}
	value, err := ReadEnv(string(key))
	if err != nil {
		return 0, err
	}

	if endEnv := midEnv + len(key); endEnv > len(b.buf) { // for garbage colector
		b.buf = nil
	}

	i := 0
	for ; startEnv+i < len(buf) && i < len(value); i++ {
		buf[startEnv+i] = value[i]
	}

	if startEnv+i < len(buf) { //add the bit
		buf[startEnv+i] = bit
		for j := startEnv + i + 1; j <= midEnv && j < len(buf); j++ { //replace *env: if value < len("*env:")
			buf[j] = ' '
		}
		return startEnv + i, nil //return position of \"
	}

	if i <= len(value) { // add the remaining in the extra buffer
		b.buf = make([]byte, len(value)-i+1)
		for j := 0; j+i < len(value); j++ {
			b.buf[j] = value[j+i]
		}
		b.buf[len(value)-i] = bit //add the bit
	}
	return len(buf), nil
}

//checkMeta check if char mach with next char from MetaEnv if not reset the counting
func (b *EnvReader) checkMeta(bit byte) bool {
	if bit == utils.MetaEnv[b.m] {
		if bit == ':' {
			b.m = 0
			return true
		}
		b.m++
		return false
	}
	b.m = 0 //reset counting
	return false
}

// Read implementation
func (b *EnvReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	pOf := 0
	b.m = 0
	if len(b.buf) > 0 { //try read extra
		pOf = b.r
		for ; b.r < len(b.buf) && b.r-pOf < len(p); b.r++ {
			p[b.r-pOf] = b.buf[b.r]
			if isEnv := b.checkMeta(p[b.r-pOf]); isEnv {
				b.r, err = b.replaceEnv(p, b.r-len(utils.MetaEnv)+1, b.r)
				if err != nil {
					return b.r, err
				}
			}
		}
		pOf = b.r - pOf
		if pOf >= len(p) {
			return pOf, nil
		}
		if len(b.buf) <= b.r {
			b.buf = nil
			b.r = 0
		}
	}
	for ; pOf < len(p); pOf++ { //normal read
		p[pOf], err = b.rd.ReadByte()
		if err != nil {
			return pOf, err
		}
		if isEnv := b.checkMeta(p[pOf]); isEnv {
			pOf, err = b.replaceEnv(p, pOf-len(utils.MetaEnv)+1, pOf)
			if err != nil {
				return pOf, err
			}
		}
	}
	if b.m != 0 { //continue to read if posible meta
		initMeta := b.m
		buf := make([]byte, len(utils.MetaEnv)-initMeta)
		i := 0
		for ; b.m != 0; i++ {
			buf[i], err = b.rd.ReadByte()
			if err != nil {
				return i - 1, err
			}
			if isEnv := b.checkMeta(buf[i]); isEnv {
				i, err = b.replaceEnv(p, len(p)-initMeta, i)
				if err != nil {
					return i, err
				}
				buf = nil
			}
		}
		if len(buf) > 0 {
			b.buf = buf[:i]
		}
	}
	return len(p), nil
}
