package main

import (
	"fmt"
)

type Responder int

func (r *Responder) Get(args string, replay *string) error {		
	*replay = fmt.Sprintf("{'response': %s}", callRater(args))
	return nil
}

