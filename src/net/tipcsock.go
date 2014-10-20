// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
    "strconv"
)

const (
    TIPC_ADDR_NAMESEQ = 1
    TIPC_ADDR_MCAST   = 1
    TIPC_ADDR_NAME    = 2
    TIPC_ADDR_ID      = 3
)

const (
    TIPC_ZONE_SCOPE     = 1
    TIPC_CLUSTER_SCOPE  = 2
    TIPC_NODE_SCOPE     = 3
)

func JoinServiceInstance(service, instance uint32) (string) {
    var s = strconv.FormatUint(uint64(service), 10)
    var i = strconv.FormatUint(uint64(instance), 10)
    return s + "." + i
}

func JoinServiceInstanceRange(service, low, high uint32) (string) {
    var s = strconv.FormatUint(uint64(service), 10)
    var l = strconv.FormatUint(uint64(low), 10)
    var h = strconv.FormatUint(uint64(high), 10)
    return s + "." + l + "-" + h
}

// TIPCAddr represents the address of a TIPC end point.
type TIPCAddr struct {
	AddrType   uint8  // only supporting TIPC_ADDR_NAME right now in Resolve
	Scope      int8   // only used in bind
	Service    uint32
	Instance   uint32
	Domain     uint32
}

// Network returns the address's network name, "tipc".
func (a *TIPCAddr) Network() string { return "tipc" }

func (a *TIPCAddr) String() string {
	if a == nil {
		return "<nil>"
	}
	if a.AddrType == TIPC_ADDR_NAME {
		return JoinServiceInstance(a.Service, a.Instance)
	} else if a.AddrType == TIPC_ADDR_NAMESEQ {
		return JoinServiceInstanceRange(a.Service, a.Instance, a.Domain)
    }
	return "<nil>"
}


func (a *TIPCAddr) toAddr() Addr {
	if a == nil {
		return nil
	}
	return a
}

// ResolveTIPCAddr parses addr as a TIPC address of the form "service:instance"
// and resolves a pair of service name and port name on the network net, 
// which must be "tipc".

func ResolveTIPCAddr(net, addr string) (*TIPCAddr, error) {
	switch net {
	case "tipc":
		net = "tipc"
	default:
		return nil, UnknownNetworkError(net)
	}
    var i = last(addr, '.')
    if i < 0 {
        return nil, UnknownNetworkError(net)
    }
    service, err := strconv.ParseUint(addr[:i], 10, 32)
    if err != nil {
        return nil, UnknownNetworkError(net)
    }
    instance, err := strconv.ParseUint(addr[i+1:], 10, 32)
    if err != nil {
        return nil, UnknownNetworkError(net)
    }

    // TODO add support for other types
	var x = TIPCAddr{AddrType:TIPC_ADDR_NAME, Scope: 0, Service: uint32(service), Instance: uint32(instance), Domain: 0}
    return &x, nil
}
