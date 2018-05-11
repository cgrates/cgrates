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

package main

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

func NewFileReaderTester(fPath, cgrAddr string, parallel, runs int, reqSep []byte) (frt *FileReaderTester, err error) {
	frt = &FileReaderTester{
		parallel: parallel, runs: runs,
		reqSep: reqSep,
	}
	if frt.rdr, err = os.Open(fPath); err != nil {
		return nil, err
	}
	if frt.conn, err = net.Dial("tcp", cgrAddr); err != nil {
		return nil, err
	}
	return
}

// TesterReader will read requests from file and post them remotely
type FileReaderTester struct {
	parallel int
	runs     int
	reqSep   []byte

	rdr      io.Reader
	conn     net.Conn
	connScnr *bufio.Scanner
}

func (frt *FileReaderTester) connSendReq(req []byte) (err error) {
	frt.conn.SetReadDeadline(time.Now().Add(time.Millisecond)) // will block most of the times on read
	if _, err = frt.conn.Write(req); err != nil {
		return
	}
	ioutil.ReadAll(frt.conn)
	return
}

// Test reads from rdr, split the content based on lineSep and sends individual lines to remote
func (frt *FileReaderTester) Test() (err error) {
	var fContent []byte
	if fContent, err = ioutil.ReadAll(frt.rdr); err != nil {
		return
	}
	reqs := bytes.Split(fContent, frt.reqSep)

	// parallel requests
	if frt.parallel > 0 {
		var wg sync.WaitGroup
		reqLimiter := make(chan struct{}, frt.parallel)
		for i := 0; i < frt.runs; i++ {
			wg.Add(1)
			go func(i int) {
				reqLimiter <- struct{}{} // block till buffer will allow
				if err := frt.connSendReq(reqs[rand.Intn(len(reqs))]); err != nil {
					log.Printf("ERROR: %s", err.Error())
				}
				<-reqLimiter // release one request from buffer
				wg.Done()
			}(i)
		}
		wg.Wait()
		return
	}

	// serial requests
	for i := 0; i < frt.runs; i++ {
		for _, req := range reqs {
			if err := frt.connSendReq(req); err != nil {
				log.Printf("ERROR: %s", err.Error())
			}
		}
	}
	return
}
