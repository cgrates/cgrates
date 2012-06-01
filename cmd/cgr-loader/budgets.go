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
package main

import (
	"log"
	"os"
	"encoding/csv"
)

var (
	volumeDiscounts = make(map[string][]string)
	volumeRates     = make(map[string][]string)
	inboundBonuses  = make(map[string][]string)
	outboundBonuses = make(map[string][]string)
	recurrentDebits = make(map[string][]string)
	recurrentTopups = make(map[string][]string)
	balanceProfiles = make(map[string][]string)
)

func loadVolumeDicounts() {
	fp, err := os.Open(*volumediscountsFn)
	if err != nil {
		log.Printf("Could not open volume discounts file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		volumeDiscounts[tag] = append(volumeDiscounts[tag], record[1:]...)
		log.Print(tag, volumeDiscounts[tag])
	}
}

func loadVolumeRates() {
	fp, err := os.Open(*volumeratesFn)
	if err != nil {
		log.Printf("Could not open volume rates file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		volumeRates[tag] = append(volumeRates[tag], record[1:]...)
		log.Print(tag, volumeRates[tag])
	}
}

func loadInboundBonuses() {
	fp, err := os.Open(*inboundbonusesFn)
	if err != nil {
		log.Printf("Could not open inbound bonueses file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		inboundBonuses[tag] = append(inboundBonuses[tag], record[1:]...)
		log.Print(tag, inboundBonuses[tag])
	}
}

func loadOutboundBonuses() {
	fp, err := os.Open(*inboundbonusesFn)
	if err != nil {
		log.Printf("Could not open outbound bonueses file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		outboundBonuses[tag] = append(outboundBonuses[tag], record[1:]...)
		log.Print(tag, outboundBonuses[tag])
	}
}

func loadRecurrentDebits() {
	fp, err := os.Open(*recurrentdebitsFn)
	if err != nil {
		log.Printf("Could not open recurent debits file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		recurrentDebits[tag] = append(recurrentDebits[tag], record[1:]...)
		log.Print(tag, recurrentDebits[tag])
	}
}

func loadRecurrentTopups() {
	fp, err := os.Open(*recurrenttopupsFn)
	if err != nil {
		log.Printf("Could not open recurent topups file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Tag" {
			// skip header line
			continue
		}
		recurrentTopups[tag] = append(recurrentTopups[tag], record[1:]...)
		log.Print(tag, recurrentTopups[tag])
	}
}

func loadBalanceProfiles() {
	fp, err := os.Open(*balanceprofilesFn)
	if err != nil {
		log.Printf("Could not open balance profiles file: %v", err)
		return
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	csvReader.Comma = sep
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		tag := record[0]
		if tag == "Account" {
			// skip header line
			continue
		}
		balanceProfiles[tag] = append(balanceProfiles[tag], record...)
		log.Print(tag, balanceProfiles[tag])
	}
}
