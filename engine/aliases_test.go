package engine

import "testing"

func init() {
	aliasService = NewAliasHandler(accountingStorage)
}
func TestAliasesGetAlias(t *testing.T) {
	alias := Alias{}
	err := aliasService.GetAlias(Alias{
		Direction: "*out",
		Tenant:    "cgrates.org",
		Category:  "call",
		Account:   "dan",
		Subject:   "dan",
		Context:   "*rating",
	}, &alias)
	if err != nil ||
		len(alias.Values) != 2 ||
		len(alias.Values[0].Pairs) != 2 {
		t.Error("Error getting alias: ", err, alias)
	}
}

func TestAliasesGetMatchingAlias(t *testing.T) {
	var response string
	err := aliasService.GetMatchingAlias(AttrMatchingAlias{
		Direction:   "*out",
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "dan",
		Subject:     "dan",
		Context:     "*rating",
		Destination: "444",
		Target:      "Subject",
		Original:    "rif",
	}, &response)
	if err != nil || response != "rif1" {
		t.Error("Error getting alias: ", err, response)
	}
}

func TestAliasesLoadAlias(t *testing.T) {
	var response string
	cd := &CallDescriptor{
		Direction:   "*out",
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "rif",
		Subject:     "rif",
		Destination: "444",
		ExtraFields: map[string]string{
			"Cli":   "0723",
			"Other": "stuff",
		},
	}
	err := LoadAlias(
		&AttrMatchingAlias{
			Direction:   "*out",
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "dan",
			Subject:     "dan",
			Context:     "*rating",
			Destination: "444",
		}, cd, "ExtraFields")
	if err != nil || cd == nil {
		t.Error("Error getting alias: ", err, response)
	}
	if cd.Subject != "rif1" ||
		cd.ExtraFields["Cli"] != "0724" {
		t.Errorf("Aliases failed to change interface: %+v", cd)
	}
}
