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

/*
import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type FailoverPoster struct {
	sync.Mutex
	mp MessageProvider // e.g kafka
}

func NewFailoverPoster(dataType, tenant, nodeID, conn, format, fldPostDir string,
	logLvl, attempts int) *FailoverPoster {
	fldPst := new(FailoverPoster)
	switch dataType {
	case MetaKafkaLog:
		fldPst.mp = NewExportLogger(nodeID, tenant, logLvl, conn, format, attempts, fldPostDir)
	}
	return fldPst
}

func (fldPst *FailoverPoster) AddFailedMessage(content interface{}) (err error) {
	fldPst.Lock()
	meta := fldPst.mp.GetMeta()
	filePath := filepath.Join(meta[FailedPostsDir].(string), meta[Format].(string)+
		PipeSep+MetaKafkaLog+GOBSuffix)
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
	failPoster := &FailoverPosterData{
		MetaData: meta,
		Content:  content.(*CGREvent),
	}
	enc := gob.NewEncoder(fileOut)
	err = enc.Encode(failPoster)
	fileOut.Close()
	fldPst.Unlock()
	return
}

type MessageProvider interface {
	GetContent(filePath string) (string, error)
	GetMeta() map[string]interface{}
}

func NewMessageProvider(dataType string) (MessageProvider, error) {
	switch dataType {
	case MetaKafkaLog:
		return new(ExportLogger), nil
	default:
		return nil, fmt.Errorf("Invalid Message Provider type in order to read the failed posts")
	}
}


func (fldPst *FailoverPoster) GetContent(filePath string) (string, error) {

}

func (fldPst *FailoverPoster) GetMeta() string {
	return EmptyString
}


// FailoverPosterData will keep the data and the content of the failed post. It is used when we read from gob file to know these info
type FailoverPosterData struct {
	MetaData map[string]interface{}
	Content  *CGREvent
} */
