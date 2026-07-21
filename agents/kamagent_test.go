// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package agents

import (
	"testing"

	"github.com/cgrates/cgrates/sessions"
)

func TestKAsSessionSClientIface(t *testing.T) {
	_ = sessions.BiRPCClient(new(KamailioAgent))
}
