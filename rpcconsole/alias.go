// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
