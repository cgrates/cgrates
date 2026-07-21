// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ugorji/go/codec"
	"go.mongodb.org/mongo-driver/bson"
)

// NewMarshaler returns the marshaler type selected by mrshlerStr
func NewMarshaler(mrshlerStr string) (ms Marshaler, err error) {
	switch mrshlerStr {
	case MsgPack:
		ms = NewCodecMsgpackMarshaler()
	case JSON:
		ms = new(JSONMarshaler)
	default:
		err = fmt.Errorf("Unsupported marshaler: %v", mrshlerStr)
	}
	return
}

type Marshaler interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

type JSONMarshaler struct{}

func (JSONMarshaler) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONMarshaler) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

type BSONMarshaler struct{}

func (BSONMarshaler) Marshal(v any) ([]byte, error) {
	return bson.Marshal(v)
}

func (BSONMarshaler) Unmarshal(data []byte, v any) error {
	return bson.Unmarshal(data, v)
}

type JSONBufMarshaler struct{}

func (JSONBufMarshaler) Marshal(v any) (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(v)
	data = buf.Bytes()
	return
}

func (JSONBufMarshaler) Unmarshal(data []byte, v any) error {
	return json.NewDecoder(bytes.NewBuffer(data)).Decode(v)
}

type CodecMsgpackMarshaler struct {
	mh *codec.MsgpackHandle
}

func NewCodecMsgpackMarshaler() *CodecMsgpackMarshaler {
	mh := new(codec.MsgpackHandle)
	mh.MapType = reflect.TypeOf(map[string]any(nil))
	mh.RawToString = true
	// mh.TimeNotBuiltin = true
	return &CodecMsgpackMarshaler{mh}
}

func (cmm *CodecMsgpackMarshaler) Marshal(v any) (b []byte, err error) {
	enc := codec.NewEncoderBytes(&b, cmm.mh)
	err = enc.Encode(v)
	return
}

func (cmm *CodecMsgpackMarshaler) Unmarshal(data []byte, v any) error {
	dec := codec.NewDecoderBytes(data, cmm.mh)
	return dec.Decode(v)
}

type BincMarshaler struct {
	bh *codec.BincHandle
}

func NewBincMarshaler() *BincMarshaler {
	return &BincMarshaler{new(codec.BincHandle)}
}

func (bm *BincMarshaler) Marshal(v any) (b []byte, err error) {
	enc := codec.NewEncoderBytes(&b, bm.bh)
	err = enc.Encode(v)
	return
}

func (bm *BincMarshaler) Unmarshal(data []byte, v any) error {
	dec := codec.NewDecoderBytes(data, bm.bh)
	return dec.Decode(v)
}

type GOBMarshaler struct{}

func (GOBMarshaler) Marshal(v any) (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = gob.NewEncoder(buf).Encode(v)
	data = buf.Bytes()
	return
}

func (GOBMarshaler) Unmarshal(data []byte, v any) error {
	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(v)
}
