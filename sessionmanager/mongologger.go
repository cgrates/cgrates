/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package sessionmanager

import (
	"fmt"
	"github.com/cgrates/cgrates/timespans"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type LogEntry struct {
	UUID     string
	CallCost *timespans.CallCost
}

type MongoLogger struct {
	Col *mgo.Collection
}

func (ml *MongoLogger) Log(uuid string, cc *timespans.CallCost) {
	if ml.Col == nil {
		//timespans.Logger.Warning("Cannot write log to database.")
		return
	}

	err := ml.Col.Insert(&LogEntry{uuid, cc})
	if err != nil {
		timespans.Logger.Err(fmt.Sprintf("failed to execute insert statement: %v", err))
	}
}

func (ml *MongoLogger) GetLog(uuid string)(cc *timespans.CallCost, err error) {
	result := new(LogEntry)
	err = ml.Col.Find(bson.M{"uuid": uuid}).One(result)
	cc = result.CallCost
    return 
}