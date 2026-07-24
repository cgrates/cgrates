// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
