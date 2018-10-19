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
	"bytes"
	"errors"
	// "fmt"
	"io"
	"os"

	"github.com/cgrates/cgrates/utils"
)

// Reads the enviorment variable
func ReadEnv(key string) (string, error) {
	if env := os.Getenv(key); env != "" {
		return env, nil
	}
	return "", errEnvNotFound
}

func NewRawJSONReader(r io.Reader) io.Reader {
	return &EnvReader{
		rd: &comentByteReader{
			rd: bufio.NewReader(r),
		},
	}
}

type comentByteReader struct {
	rd           *bufio.Reader
	isPosComment bool // for comment removal
	isInString   bool // ignore coment delimiters in string declarations
}

func isNL(c byte) bool {
	return c == '\n' || c == '\r'
}

func isWS(c byte) bool {
	return c == ' ' || c == '\t' || isNL(c)
}

func (b *comentByteReader) consumeNL() error {
	for {
		bit, err := b.rd.ReadByte()
		if err != nil || isNL(bit) {
			return err
		}
	}
}
func (b *comentByteReader) consumeMLC() error {
	stop := false
	for {
		bit, err := b.rd.ReadByte()
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF {
			return errors.New("Incomplete Comment")
		}

		if stop && bit == '/' {
			return nil
		}
		stop = bit == '*'
	}
}

func (b *comentByteReader) ReadByte() (byte, error) {
	bit, err := b.rd.ReadByte()
	if err != nil {
		return bit, err
	}
	if bit == '"' {
		b.isInString = !b.isInString
		return bit, err
	}
	if !b.isInString {
		if bit == '/' {
			bit2, err := b.rd.Peek(1)
			if err != nil {
				return bit, err
			}
			switch bit2[0] {
			case '/':
				err = b.consumeNL()
				if err == io.EOF {
					return 0, err
				}
				return '\n', err
			case '*':
				if err = b.consumeMLC(); err != nil {
					return 0, err
				}
				return b.rd.ReadByte()
			}
		}
	}
	return bit, err
}

type EnvReader struct {
	buf []byte
	rd  io.ByteReader // reader provided by the client
	r   int           // buf read positions
	m   int           // meta Ofset
}

var errNegativeRead = errors.New("reader returned negative count from Read")
var errEnvNotFound = errors.New("reader cant find enviormental variable")

func (b *EnvReader) readEnvName() (name []byte, bit byte, err error) { //0 if not set
	esc := []byte{' ', '\t', '\n', '\r', ',', '}', ']', '\'', '"', '/'}
	for { //read byte by byte
		bit, err := b.rd.ReadByte()
		if err != nil {
			return name, 0, err
		}
		if bytes.IndexByte(esc, bit) != -1 {
			return name, bit, nil
		}
		name = append(name, bit)
	}
}

func (b *EnvReader) replaceEnv(buf []byte, startEnv, midEnv int) (n int, err error) {
	key, bit, err := b.readEnvName()
	if err != nil && err != io.EOF {
		return 0, err
	}
	value, err := ReadEnv(string(key))
	if err != nil {
		return 0, err
	}

	if endEnv := midEnv + len(key); endEnv > len(b.buf) { // garbage
		b.buf = nil
	}

	i := 0
	for ; startEnv+i < len(buf) && i < len(value); i++ {
		buf[startEnv+i] = value[i]
	}

	if startEnv+i < len(buf) { //add the bit
		buf[startEnv+i] = bit
		for j := startEnv + i + 1; j <= midEnv && j < len(buf); j++ { //replace *env: if value < len("*env:")
			buf[j] = 0
		}
		return startEnv + i, nil
	}

	if i <= len(value) { //pune restul in buferul extra
		b.buf = make([]byte, len(value)-i+1)
		for j := 0; j+i < len(value); j++ {
			b.buf[j] = value[j+i]
		}
		b.buf[len(value)-i] = bit //add the bit
	}
	return len(buf), nil
}

func (b *EnvReader) checkMeta(bit byte) bool {
	if bit == utils.MetaEnv[b.m] {
		if bit == ':' {
			b.m = 0
			return true
		}
		b.m++
		return false
	}
	b.m = 0
	return false
}

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
				startEnv := b.r - len(utils.MetaEnv) + 1
				midEnv := b.r
				b.r, err = b.replaceEnv(p, startEnv, midEnv)
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
			startEnv := pOf - len(utils.MetaEnv) + 1
			midEnv := pOf
			pOf, err = b.replaceEnv(p, startEnv, midEnv)
			if err != nil {
				return pOf, err
			}
		}

	}
	if b.m != 0 { //continue to read if posible meta
		initMeta := b.m
		buf := make([]byte, len(utils.MetaEnv)-initMeta)
		for i := 0; b.m != 0; i++ {
			buf[i], err = b.rd.ReadByte()
			if err != nil {
				return i - 1, err
			}
			if isEnv := b.checkMeta(buf[i]); isEnv {
				startEnv := len(p) - initMeta
				midEnv := i
				i, err = b.replaceEnv(p, startEnv, midEnv)
				if err != nil {
					return i, err
				}
				buf = nil
			}

		}
		if len(buf) > 0 {
			b.buf = buf
		}
	}
	return len(p), nil

}

// func (b *EnvReader) commentGuard(bit byte) byte {
// 	// fmt.Println(b.isPosComment, string(bit))
// 	if b.isPosComment && bit == '/' {
// 		b.isPosComment = false
// 		return 1
// 	} else if b.isPosComment && bit == '*' {
// 		b.isPosComment = false
// 		return 2
// 	}
// 	b.isPosComment = bit == '/'
// 	return 0
// }
