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
	"sync"
)

var (
	cm = NewChanMutex()
)

type Command struct {
	name string
	data string
}

func (c *Command) Execute() {
	ch := cm.pipe[c.name]
	ch <- c
	log.Print(c.data)
	time.Sleep(1 * time.Second)
	log.Print("end ", c.data)
	<-ch
}

type ChanMutex struct {
	pipe map[string]chan *Command
	mu   sync.Mutex
}

func NewChanMutex() *ChanMutex {
	return &ChanMutex{pipe: make(map[string]chan *Command)}
}

func (cm *ChanMutex) Execute(c *Command) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	ch, exists := cm.pipe[c.name]
	if !exists {
		log.Print("make chanell for: ", c.name)
		ch = make(chan *Command, 1)
		cm.pipe[c.name] = ch
	}
	go c.Execute()
}

func main() {
	cm.Execute(&Command{"vdf:rif", "prima rif"})
	cm.Execute(&Command{"vdf:dan", "prima Dan"})
	cm.Execute(&Command{"vdf:rif", "a doua rif"})
	time.Sleep(8 * time.Second)
}
