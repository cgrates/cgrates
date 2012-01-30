package main

type Responder byte

/*
RPC method thet provides the external RPC interface for getting the rating information.
*/
func (r *Responder) Get(args string, replay *string) error {	
	*replay = CallRater(args)
	return nil
}

