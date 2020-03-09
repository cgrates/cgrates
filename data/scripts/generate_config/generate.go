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
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/cgrates/cgrates/config"
)

func writeDefaultCofig(fileName string) error {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	rows := strings.Split(config.CGRATES_CFG_JSON, "\n")[1:] // remove first empty row
	for i, row := range rows {
		if i == 0 || i == len(rows)-1 { // do not comment first and last row
			fmt.Fprintln(f, row)
			continue
		}
		if withoutSpace := strings.TrimSpace(row); len(withoutSpace) == 0 || strings.HasPrefix(row, "//") { // do not comment empty rows and alerady commented ones
			fmt.Fprintln(f, row)
			continue
		}
		fmt.Fprintf(f, "// %s\n", row)
	}
	return nil
}

// used only to generate the commented configuration file
func main() {
	generateFlags := flag.NewFlagSet("generate", flag.ContinueOnError)
	cfgPath := generateFlags.String("config_path", path.Join("/usr", "share", "cgrates", "conf", "cgrates", "cgrates.json"), "The file path for generated configuration.")
	if err := generateFlags.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Generating configuration file ...")
	if err := os.Remove(*cfgPath); err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
	if err := writeDefaultCofig(*cfgPath); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Done writing file at path: %s\n", *cfgPath)
}
