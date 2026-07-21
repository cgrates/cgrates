// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package rpcconsole

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// Client is an RPC connection plus the method list and descriptors it fetches
// from the engine on connect.
type Client struct {
	cl      *rpcclient.RPCClient
	methods []string                           // sorted "Service.Method" names
	descs   map[string]*utils.MethodDescriptor // descriptor by method name
	aliases map[string]string                  // alias -> RPC name
}

// NewClient fetches the methods the engine serves and returns a ready Client.
func NewClient(cl *rpcclient.RPCClient) (*Client, error) {
	var mds []utils.MethodDescriptor
	if err := cl.Call(context.Background(), utils.CoreSv1DescribeMethods,
		&utils.DescribeMethodsArgs{
			SkipServices: []string{"_goRPC_", utils.ReplicatorSv1},
		}, &mds); err != nil {
		return nil, err
	}
	c := &Client{
		cl:      cl,
		methods: make([]string, len(mds)),
		descs:   make(map[string]*utils.MethodDescriptor, len(mds)),
		aliases: make(map[string]string, len(mds)),
	}
	for i := range mds {
		c.methods[i] = mds[i].Method
		c.descs[mds[i].Method] = &mds[i]
		c.aliases[Alias(mds[i].Method)] = mds[i].Method
	}
	return c, nil
}

// Methods returns the sorted "Service.Method" names the engine serves.
func (c *Client) Methods() []string {
	return c.methods
}

// Resolve maps an alias to its RPC name. A raw RPC name passes through unchanged.
func (c *Client) Resolve(method string) string {
	if rpc, ok := c.aliases[method]; ok {
		return rpc
	}
	return method
}

// Describe returns the method's descriptor, or nil if the engine doesn't serve
// it. Takes an alias or the RPC name.
func (c *Client) Describe(method string) *utils.MethodDescriptor {
	return c.descs[c.Resolve(method)]
}

// Call runs method with params and returns the decoded reply. Takes an alias or
// the RPC name.
func (c *Client) Call(method string, params any) (any, error) {
	var reply any
	if err := c.cl.Call(context.Background(), c.Resolve(method), params, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}
