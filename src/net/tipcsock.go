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

const addressDelimiter = ';'
const rangeDelimiter = '-'

func JoinServiceInstance(service, instance uint32) (string) {
    var s = strconv.FormatUint(uint64(service), 10)
    var i = strconv.FormatUint(uint64(instance), 10)
    return s + string(addressDelimiter) + i
}

func JoinServiceInstanceRange(service, low, high uint32) (string) {
    var s = strconv.FormatUint(uint64(service), 10)
    var l = strconv.FormatUint(uint64(low), 10)
    var h = strconv.FormatUint(uint64(high), 10)
    return s + string(addressDelimiter) + l + string(rangeDelimiter) + h
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
		return "<tipcnil>"
	}
	if a.AddrType == TIPC_ADDR_NAME {
		return JoinServiceInstance(a.Service, a.Instance)
	} else if a.AddrType == TIPC_ADDR_NAMESEQ {
		return JoinServiceInstanceRange(a.Service, a.Instance, a.Domain)
    }
	return "<tipcundef>"
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
	if net != "tipc" {
		return nil, UnknownNetworkError(net)
	}

    var addrSep = last(addr, addressDelimiter)

    if addrSep < 0 {
        return nil, UnknownNetworkError(net)
    }

    service, err := strconv.ParseUint(addr[:addrSep], 10, 32)

    if err != nil {
        return nil, UnknownNetworkError(net)
    }

    var rangeSep = last(addr, rangeDelimiter)

    if rangeSep < 0 {

        instance, err := strconv.ParseUint(addr[addrSep+1:], 10, 32)

        if err != nil {
            return nil, UnknownNetworkError(net)
        }

	    var x = TIPCAddr{AddrType: TIPC_ADDR_NAME, Scope: TIPC_ZONE_SCOPE, Service: uint32(service), Instance: uint32(instance), Domain: 0}

        return &x, nil

    } else {

        startRange, err := strconv.ParseUint(addr[addrSep + 1:rangeSep], 10, 32)
        if err != nil {
            return nil, UnknownNetworkError(net)
        }

        endRange, err := strconv.ParseUint(addr[rangeSep + 1:], 10, 32)
        if err != nil {
            return nil, UnknownNetworkError(net)
        }

	    var x = TIPCAddr{AddrType: TIPC_ADDR_NAMESEQ, Scope: TIPC_ZONE_SCOPE, Service: uint32(service), Instance: uint32(startRange), Domain:uint32(endRange)}
        return &x, nil
    }

}
