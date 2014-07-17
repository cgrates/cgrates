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
	XMLName        xml.Name               `xml:"document"`
	Type           string                 `xml:"type,attr"`
	Configurations []*CgrXmlConfiguration `xml:"configuration"`
	cdrcs          map[string]*CgrXmlCdrcCfg
	cdres          map[string]*CgrXmlCdreCfg // Cahe cdrexporter instances, key will be the ID
}

// Storage for raw configuration
type CgrXmlConfiguration struct {
	XMLName   xml.Name `xml:"configuration"`
	Section   string   `xml:"section,attr"`
	Id        string   `xml:"id,attr"`
	RawConfig []byte   `xml:",innerxml"` // Used to store the configuration struct, as raw so we can store different types
}

func (cfgInst *CgrXmlConfiguration) rawConfigElement() []byte {
	rawConfig := append([]byte("<element>"), cfgInst.RawConfig...) // Encapsulate the rawConfig in one element so we can Unmarshall into one struct
	rawConfig = append(rawConfig, []byte("</element>")...)
	return rawConfig
}

func (xmlCfg *CgrXmlCfgDocument) cacheAll() error {
	for _, cacheFunc := range []func() error{xmlCfg.cacheCdrcCfgs, xmlCfg.cacheCdreCfgs} {
		if err := cacheFunc(); err != nil {
			return err
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
			if err := fld.PopulateRSRField(); err != nil {
				return fmt.Errorf("Populating field %s, error: %s", fld.Id, err.Error())
			}
		}
		cdrcCfg.setDefaults()
		xmlCfg.cdrcs[cfgInst.Id] = cdrcCfg
	}
	return nil
}

// Avoid building from raw config string always, so build cache here
func (xmlCfg *CgrXmlCfgDocument) cacheCdreCfgs() error {
	xmlCfg.cdres = make(map[string]*CgrXmlCdreCfg)
	for _, cfgInst := range xmlCfg.Configurations {
		if cfgInst.Section != utils.CDRE {
			continue
		}
		cdreCfg := new(CgrXmlCdreCfg)
		if err := xml.Unmarshal(cfgInst.rawConfigElement(), cdreCfg); err != nil {
			return err
		} else if cdreCfg == nil {
			return fmt.Errorf("Could not unmarshal CgrXmlCdreCfg: %s", cfgInst.Id)
		}
		if cdreCfg.Header != nil {
			// Cache rsr fields
			for _, fld := range cdreCfg.Header.Fields {
				if err := fld.populateRSRField(); err != nil {
					return fmt.Errorf("Populating field %s, error: %s", fld.Name, err.Error())
				}
				if err := fld.populateFltrRSRField(); err != nil {
					return fmt.Errorf("Populating field %s, error: %s", fld.Name, err.Error())
				}
			}
		}
		if cdreCfg.Content != nil {
			// Cache rsr fields
			for _, fld := range cdreCfg.Content.Fields {
				if err := fld.populateRSRField(); err != nil {
					return fmt.Errorf("Populating field %s, error: %s", fld.Name, err.Error())
				}
				if err := fld.populateFltrRSRField(); err != nil {
					return fmt.Errorf("Populating field %s, error: %s", fld.Name, err.Error())
				}
			}
		}
		if cdreCfg.Trailer != nil {
			// Cache rsr fields
			for _, fld := range cdreCfg.Trailer.Fields {
				if err := fld.populateRSRField(); err != nil {
					return fmt.Errorf("Populating field %s, error: %s", fld.Name, err.Error())
				}
				if err := fld.populateFltrRSRField(); err != nil {
					return fmt.Errorf("Populating field %s, error: %s", fld.Name, err.Error())
				}
			}
		}
		xmlCfg.cdres[cfgInst.Id] = cdreCfg
	}
	return nil
}

// Return instances or filtered instance of cdrefw configuration
func (xmlCfg *CgrXmlCfgDocument) GetCdreCfgs(instName string) map[string]*CgrXmlCdreCfg {
	if len(instName) != 0 {
		if cfg, hasIt := xmlCfg.cdres[instName]; !hasIt {
			return nil
		} else {
			return map[string]*CgrXmlCdreCfg{instName: cfg}
		}
	}
	return xmlCfg.cdres
}

// Return instances or filtered instance of cdrc configuration
func (xmlCfg *CgrXmlCfgDocument) GetCdrcCfgs(instName string) map[string]*CgrXmlCdrcCfg {
	if len(instName) != 0 {
		if cfg, hasIt := xmlCfg.cdrcs[instName]; !hasIt {
			return nil
		} else {
			return map[string]*CgrXmlCdrcCfg{instName: cfg} // Filtered
		}
	}
	return xmlCfg.cdrcs // Unfiltered
}
