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
	"net/rpc"
	"sync"
)

type RaterList struct {
	clientAddresses   []string
	clientConnections []*rpc.Client
	balancerIndex     int
	mu                sync.RWMutex
}

/*
Constructor for RateList holding one slice for addreses and one slice for connections.
*/
func NewRaterList() *RaterList {
	r := &RaterList{balancerIndex: 0} // leaving both slices to nil
	return r
}

/*
Adds a client to the two  internal slices.
*/
func (rl *RaterList) AddClient(address string, client *rpc.Client) {
	rl.clientAddresses = append(rl.clientAddresses, address)
	rl.clientConnections = append(rl.clientConnections, client)
	return
}

/*
Removes a client from the slices locking the readers and reseting the balancer index.
*/
func (rl *RaterList) RemoveClient(address string) {
	index := -1
	for i, v := range rl.clientAddresses {
		if v == address {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	rl.clientAddresses = append(rl.clientAddresses[:index], rl.clientAddresses[index+1:]...)
	rl.clientConnections = append(rl.clientConnections[:index], rl.clientConnections[index+1:]...)
	rl.balancerIndex = 0
}

/*
Returns a client for the specifed address.
*/
func (rl *RaterList) GetClient(address string) (*rpc.Client, bool) {
	for i, v := range rl.clientAddresses {
		if v == address {
			return rl.clientConnections[i], true
		}
	}
	return nil, false
}

/*
Returns the next available connection at each call looping at the end of connections.
*/
func (rl *RaterList) Balance() (result *rpc.Client) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.balancerIndex >= len(rl.clientAddresses) {
		rl.balancerIndex = 0
	}
	if len(rl.clientAddresses) > 0 {
		result = rl.clientConnections[rl.balancerIndex]
		rl.balancerIndex++
	}

	return
}
