/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package engine

// TPReader is the data source for TPLoader
type TPReader interface {
	// Read will read one record from data source
	Read() (any, error)
}

// TPLoader will read a record from TPReader and write it out to dataManager
type TPLoader struct {
	srcType    string       // needed by Load for choosing destination
	dataReader TPReader     // provides data to load
	dm         *DataManager // writes data to load
}
