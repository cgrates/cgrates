// +build integration

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
package config

// func TestNewCgrJsonCfgFromHttp(t *testing.T) {
// 	addr := "https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/tutmongo/cgrates.json"
// 	expVal, err := NewCgrJsonCfgFromFile(path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo", "cgrates.json"))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if _, err = net.DialTimeout("tcp", addr, time.Second); err != nil { // check if site is up
// 		return
// 	}

// 	if rply, err := NewCgrJsonCfgFromHttp(addr); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(expVal, rply) {
// 		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expVal), utils.ToJSON(rply))
// 	}

// }

// func TestNewCGRConfigFromPath(t *testing.T) {
// 	for key, val := range map[string]string{"LOGGER": "*syslog", "LOG_LEVEL": "6", "TLS_VERIFY": "false", "ROUND_DEC": "5",
// 		"DB_ENCODING": "*msgpack", "TP_EXPORT_DIR": "/var/spool/cgrates/tpe", "FAILED_POSTS_DIR": "/var/spool/cgrates/failed_posts",
// 		"DF_TENANT": "cgrates.org", "TIMEZONE": "Local"} {
// 		os.Setenv(key, val)
// 	}
// 	addr := "https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/a.json;https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/b/b.json;https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/c.json;https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/multifiles/d.json"
// 	expVal, err := NewCGRConfigFromPath(path.Join("/usr", "share", "cgrates", "conf", "samples", "multifiles"))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if _, err = net.DialTimeout("tcp", addr, time.Second); err != nil { // check if site is up
// 		return
// 	}

// 	if rply, err := NewCGRConfigFromPath(addr); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(expVal, rply) {
// 		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expVal), utils.ToJSON(rply))
// 	}

// }

/* Needs to be rewritten with a static config
func TestCgrCfgV1ReloadConfigSection(t *testing.T) {
	expected := map[string]interface{}{
		"Enabled": true,
		"Readers": []interface{}{
			map[string]interface{}{
				"ConcurrentReqs": 0.,
				"Content_fields": nil,
				"Continue":       false,
				"FieldSep":       "",
				"Filters":        nil,
				"Flags":          nil,
				"Header_fields":  nil,
				"ID":             "file_reader1",
				"ProcessedPath":  "/tmp/ers/out",
				"RunDelay":       -1.,
				"SourceID":       "",
				"SourcePath":     "/tmp/ers/in",
				"Tenant":         nil,
				"Timezone":       "",
				"Trailer_fields": nil,
				"Type":           "*file_csv",
				"XmlRootPath":    "",
			},
		},
		"SessionSConns": []interface{}{
			map[string]interface{}{
				"Address":     "*internal",
				"Synchronous": false,
				"TLS":         false,
				"Transport":   "",
			},
		},
	}

	cfg, _ := NewDefaultCGRConfig()
	var reply string
	var rcv map[string]interface{}

	if err := cfg.V1ReloadConfig(&ConfigReloadWithArgDispatcher{
		Path:    "/usr/share/cgrates/conf/samples/ers",
		Section: ERsJson,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expected: %s ,received: %s", utils.OK, reply)
	}

	if err := cfg.V1GetConfigSection(&StringWithArgDispatcher{Section: ERsJson}, &rcv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
*/
