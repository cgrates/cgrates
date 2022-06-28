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
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
)

type FailoverPoster struct {
	MessageProvider
}

func NewFailoverPoster( /*addMsg *MessageProvider*/ ) *FailoverPoster {
	return new(FailoverPoster)
}

func (fldPst *FailoverPoster) AddMessage(failedPostDir, kafkaConn,
	kafkaTopic string, content interface{}) (err error) {
	filePath := filepath.Join(failedPostDir, kafkaTopic+PipeSep+MetaKafka+GOBSuffix)
	var fileOut *os.File
	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		fileOut, err = os.Create(filePath)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("<Kafka> failed to write logs to file <%s> because <%s>", filePath, err))
		}
	} else {
		fileOut, err = os.OpenFile(filePath, os.O_RDWR|os.O_APPEND, 0755)
		if err != nil {
			return err
		}
	}
	enc := gob.NewEncoder(fileOut)
	if err = enc.Encode(content); err != nil {
		return err
	}
	fileOut.Close()
	return
}

type MessageProvider interface {
	GetContent() string
	GetMeta() string
}

func (fldPst *FailoverPoster) GetContent() string {
	return EmptyString
}

func (fldPst *FailoverPoster) GetMeta() string {
	return EmptyString
}
