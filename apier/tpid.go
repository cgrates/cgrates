/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package apier

import (
	"github.com/cgrates/cgrates/utils"
)


type TPID struct{
	TPid	string // the id of tariff plan
	Active	bool   // Marks whether this tpid is the active one in datadb
}


// Return TPid associated with id queried, new tpid in case of empty id 
func (self *Apier) GetTPID( reqAttr string, reply *TPID) error {
	*reply = TPID{Active:false}
	if reqAttr == "" {
		reply.TPid = utils.NewTPid()
	} else {
		reply.TPid = reqAttr
	}
	return nil
}
		

