/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package main

import (
	"log"
	"time"
)

var (
	cm = NewChanMutex()
)

type Command struct {
	name string
	data string
}

func (c *Command) Execute() (err error) {
	ch := cm.pipe[c.name]
	ch <- c
	log.Print(c.data)
	time.Sleep(1 * time.Second)
	log.Print("end ", c.data)
	<-ch
	return
}

type ChanMutex struct {
	pipe map[string]chan *Command
}

func NewChanMutex() *ChanMutex {
	return &ChanMutex{pipe: make(map[string]chan *Command)}
}

func (cm *ChanMutex) Execute(c *Command) {
	if _, exists := cm.pipe[c.name]; !exists {
		cm.pipe[c.name] = make(chan *Command, 1)
	}
	c.Execute()
}

func main() {
	go cm.Execute(&Command{"vdf:rif", "prima rif"})
	go cm.Execute(&Command{"vdf:dan", "prima Dan"})
	go cm.Execute(&Command{"vdf:rif", "a doua rif"})
	time.Sleep(5 * time.Second)
}
