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

import "time"

type LoadInstance struct {
	LoadID           string // Unique identifier for the load
	RatingLoadID     string
	AccountingLoadID string
	//TariffPlanID     string    // Tariff plan identificator for the data loaded
	LoadTime time.Time // Time of load
}

type CacheFileInfo struct {
	Encoding string
	LoadInfo *LoadInstance
}

/*
func LoadCacheFileInfo(path string) (*CacheFileInfo, error) {
	// open data file
	dataFile, err := os.Open(filepath.Join(path, "cache.info"))
	defer dataFile.Close()
	if err != nil {
		Logger.Err("<cache decoder>: " + err.Error())
		return nil, err
	}

	filesInfo := &CacheFileInfo{}
	dataDecoder := json.NewDecoder(dataFile)
	err = dataDecoder.Decode(filesInfo)
	if err != nil {
		Logger.Err("<cache decoder>: " + err.Error())
		return nil, err
	}
	return filesInfo, nil
}

func SaveCacheFileInfo(path string, cfi *CacheFileInfo) error {
	if path == "" {
		return nil
	}
	// open data file
	// create a the path
	if err := os.MkdirAll(path, 0766); err != nil {
		Logger.Err("<cache encoder>:" + err.Error())
		return err
	}

	dataFile, err := os.Create(filepath.Join(path, "cache.info"))
	defer dataFile.Close()
	if err != nil {
		Logger.Err("<cache encoder>:" + err.Error())
		return err
	}

	// serialize the data
	dataEncoder := json.NewEncoder(dataFile)
	if err := dataEncoder.Encode(cfi); err != nil {
		Logger.Err("<cache encoder>:" + err.Error())
		return err
	}
	return nil
}

func CacheFileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	}
	return false
}
*/
