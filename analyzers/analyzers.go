// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package analyzers

import (
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

// NewAnalyzerService initializes a AnalyzerService
func NewAnalyzerService() (*AnalyzerService, error) {
	return &AnalyzerService{}, nil
}

// AnalyzerService is the service handling analyzer
type AnalyzerService struct {
}

// ListenAndServe will initialize the service
func (aS *AnalyzerService) ListenAndServe(exitChan chan bool) error {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.AnalyzerS))
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// Shutdown is called to shutdown the service
func (aS *AnalyzerService) Shutdown() error {
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown initialized", utils.AnalyzerS))
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown complete", utils.AnalyzerS))
	return nil
}
