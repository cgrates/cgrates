/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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

package config

import (
	"encoding/xml"
	"fmt"
	"github.com/cgrates/cgrates/utils"
	"io"
)

// Decodes a reader enforcing specific format of the configuration file
func ParseCgrXmlConfig(reader io.Reader) (*CgrXmlCfgDocument, error) {
	xmlConfig := new(CgrXmlCfgDocument)
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(xmlConfig); err != nil {
		return nil, err
	}
	if err := xmlConfig.cacheAll(); err != nil {
		return nil, err
	}
	return xmlConfig, nil
}

// Define a format for configuration file, one doc contains more configuration instances, identified by section, type and id
type CgrXmlCfgDocument struct {
	XMLName        xml.Name                    `xml:"document"`
	Type           string                      `xml:"type,attr"`
	Configurations []*CgrXmlConfiguration      `xml:"configuration"`
	cdrefws        map[string]*CgrXmlCdreFwCfg // Cache for processed fixed width config instances, key will be the id of the instance
	cdrcs          map[string]*CgrXmlCdrcCfg
}

// Storage for raw configuration
type CgrXmlConfiguration struct {
	XMLName   xml.Name `xml:"configuration"`
	Section   string   `xml:"section,attr"`
	Type      string   `xml:"type,attr"`
	Id        string   `xml:"id,attr"`
	RawConfig []byte   `xml:",innerxml"` // Used to store the configuration struct, as raw so we can store different types
}

func (cfgInst *CgrXmlConfiguration) rawConfigElement() []byte {
	rawConfig := append([]byte("<element>"), cfgInst.RawConfig...) // Encapsulate the rawConfig in one element so we can Unmarshall into one struct
	rawConfig = append(rawConfig, []byte("</element>")...)
	return rawConfig
}

func (xmlCfg *CgrXmlCfgDocument) cacheAll() error {
	for _, cacheFunc := range []func() error{xmlCfg.cacheCdreFWCfgs, xmlCfg.cacheCdrcCfgs} {
		if err := cacheFunc(); err != nil {
			return err
		}
	}
	return nil
}

// Avoid building from raw config string always, so build cache here
func (xmlCfg *CgrXmlCfgDocument) cacheCdreFWCfgs() error {
	xmlCfg.cdrefws = make(map[string]*CgrXmlCdreFwCfg)
	for _, cfgInst := range xmlCfg.Configurations {
		if cfgInst.Section == utils.CDRE || cfgInst.Type == utils.CDRE_FIXED_WIDTH {
			cdrefwCfg := new(CgrXmlCdreFwCfg)
			rawConfig := append([]byte("<element>"), cfgInst.RawConfig...) // Encapsulate the rawConfig in one element so we can Unmarshall into one struct
			rawConfig = append(rawConfig, []byte("</element>")...)
			if err := xml.Unmarshal(rawConfig, cdrefwCfg); err != nil {
				return err
			} else if cdrefwCfg == nil {
				return fmt.Errorf("Could not unmarshal CgrXmlCdreFwCfg: %s", cfgInst.Id)
			} else { // All good, cache the config instance
				xmlCfg.cdrefws[cfgInst.Id] = cdrefwCfg
			}
		}
	}
	return nil
}

// Avoid building from raw config string always, so build cache here
func (xmlCfg *CgrXmlCfgDocument) cacheCdrcCfgs() error {
	xmlCfg.cdrcs = make(map[string]*CgrXmlCdrcCfg)
	for _, cfgInst := range xmlCfg.Configurations {
		if cfgInst.Section != utils.CDRC {
			continue // Another type of config instance, not interesting to process
		}
		cdrcCfg := new(CgrXmlCdrcCfg)
		if err := xml.Unmarshal(cfgInst.rawConfigElement(), cdrcCfg); err != nil {
			return err
		} else if cdrcCfg == nil {
			return fmt.Errorf("Could not unmarshal config instance: %s", cfgInst.Id)
		}
		// Cache rsr fields
		for _, fld := range cdrcCfg.CdrFields {
			if err := fld.PopulateRSRFIeld(); err != nil {
				return fmt.Errorf("Populating field %s, error: %s", fld.Id, err.Error())
			}
		}
		xmlCfg.cdrcs[cfgInst.Id] = cdrcCfg
	}
	return nil
}

func (xmlCfg *CgrXmlCfgDocument) GetCdreFWCfg(instName string) (*CgrXmlCdreFwCfg, error) {
	if cfg, hasIt := xmlCfg.cdrefws[instName]; !hasIt {
		return nil, nil
	} else {
		return cfg, nil
	}
}

func (xmlCfg *CgrXmlCfgDocument) GetCdrcCfg(instName string) (*CgrXmlCdrcCfg, error) {
	if cfg, hasIt := xmlCfg.cdrcs[instName]; !hasIt {
		return nil, nil
	} else {
		return cfg, nil
	}
}
