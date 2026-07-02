/*
Real-time Online/Offline Charging System (OCS) for Telecom/ISP environments
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

package rpcconsole

import "strings"

// Alias turns an RPC name into a console alias: strip the service version
// suffix, lowercase the method's leading capitals. AdminSv1.SetAccount ->
// admins.setAccount, IPsV1.AllocateIP -> ips.allocateIP.
func Alias(name string) string {
	service, method, ok := strings.Cut(name, ".")
	if !ok {
		return name
	}
	return serviceAlias(service) + "." + methodAlias(method)
}

// serviceAlias strips a trailing v1 and lowercases the rest. The S in Sv1 stays,
// giving the plural: AccountSv1 -> AccountS -> accounts.
func serviceAlias(service string) string {
	if len(service) >= 2 && strings.EqualFold(service[len(service)-2:], "v1") {
		service = service[:len(service)-2]
	}
	return strings.ToLower(service)
}

// methodAlias lowercases the leading capitals. When they're followed by a
// lowercase letter the last one is kept as the next word's start:
// SetAccount -> setAccount, STIRAuthenticate -> stirAuthenticate.
func methodAlias(method string) string {
	i := 0
	for i < len(method) && isUpper(method[i]) {
		i++
	}
	if i == 0 {
		return method
	}
	if i > 1 && i < len(method) && isLower(method[i]) {
		i--
	}
	return strings.ToLower(method[:i]) + method[i:]
}

func isUpper(b byte) bool { return 'A' <= b && b <= 'Z' }
func isLower(b byte) bool { return 'a' <= b && b <= 'z' }
