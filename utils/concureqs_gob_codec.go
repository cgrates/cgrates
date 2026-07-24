// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

// Most of the logic follows standard library implementation in this file
package utils

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
	"net/rpc"
)

type concReqsGobServerCodec struct {
	rwc       io.ReadWriteCloser
	dec       *gob.Decoder
	enc       *gob.Encoder
	encBuf    *bufio.Writer
	closed    bool
	allocated bool // populated if we have allocated a channel for concurrent requests
}

func NewConcReqsGobServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	buf := bufio.NewWriter(conn)
	return &concReqsGobServerCodec{
		rwc:    conn,
		dec:    gob.NewDecoder(conn),
		enc:    gob.NewEncoder(buf),
		encBuf: buf,
	}
}

func (c *concReqsGobServerCodec) ReadRequestHeader(r *rpc.Request) error {
	return c.dec.Decode(r)
}

func (c *concReqsGobServerCodec) ReadRequestBody(body any) error {
	if err := ConReqs.Allocate(); err != nil {
		return err
	}
	c.allocated = true
	return c.dec.Decode(body)
}

func (c *concReqsGobServerCodec) WriteResponse(r *rpc.Response, body any) (err error) {
	if c.allocated {
		defer func() {
			ConReqs.Deallocate()
			c.allocated = false
		}()
	}
	if err = c.enc.Encode(r); err != nil {
		if c.encBuf.Flush() == nil {
			// Gob couldn't encode the header. Should not happen, so if it does,
			// shut down the connection to signal that the connection is broken.
			log.Println("rpc: gob error encoding response:", err)
			c.Close()
		}
		return
	}
	if err = c.enc.Encode(body); err != nil {
		if c.encBuf.Flush() == nil {
			// Was a gob problem encoding the body but the header has been written.
			// Shut down the connection to signal that the connection is broken.
			log.Println("rpc: gob error encoding body:", err)
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *concReqsGobServerCodec) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}
