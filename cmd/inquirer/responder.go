package main

import (
		"github.com/rif/cgrates/timespans"
)

type Responder byte

/*
RPC method thet provides the external RPC interface for getting the rating information.
*/
func (r *Responder) Get(arg timespans.CallDescriptor, replay *timespans.CallCost) error {
	*replay = *CallRater(&arg)
	return nil
}
