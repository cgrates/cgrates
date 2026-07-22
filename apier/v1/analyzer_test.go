// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"testing"

	"github.com/cgrates/cgrates/analyzers"
)

func TestNewAnalyzerSv1(t *testing.T) {
	analyzerService := &analyzers.AnalyzerService{}
	analyzerSv1 := NewAnalyzerSv1(analyzerService)
	if analyzerSv1 == nil {
		t.Errorf("expected non-nil AnalyzerSv1, got nil")
	}
	if analyzerSv1.aS != analyzerService {
		t.Errorf("expected AnalyzerService to be %v, got %v", analyzerService, analyzerSv1.aS)
	}
}
