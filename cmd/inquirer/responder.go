package main

type Responder int

func (r *Responder) Get(args string, replay *string) error {		
	*replay = CallRater(args)
	return nil
}

