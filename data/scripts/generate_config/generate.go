// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func writeDefaultConfig(fileName string) error {
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
	if err := writeDefaultConfig(*cfgPath); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Done writing file at path: %s\n", *cfgPath)
}
