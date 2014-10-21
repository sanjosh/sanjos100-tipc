// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package net

//  incorporate http://tipc.sourceforge.net/doc/Programmers_Guide.txt

import (
	"io"
	"os"
	"syscall"
	"time"
	"encoding/binary"
)

// convert system sockaddr to net.TIPCAddr
func sockaddrToTIPC(sa syscall.Sockaddr) Addr {
	switch sa := sa.(type) {
	case *syscall.SockaddrTIPC:

        sub_ser := make([]byte, 4)
        sub_inst := make([]byte, 4)
        sub_dom := make([]byte, 4)

        for index := 0; index < 4; index ++ {
            sub_ser[index] = sa.Addr[index]
            sub_inst[index] = sa.Addr[index + 4]
            sub_dom[index] = sa.Addr[index + 8]
        }
        var service = binary.LittleEndian.Uint32(sub_ser)
        var instance = binary.LittleEndian.Uint32(sub_inst)
        var domain = binary.LittleEndian.Uint32(sub_dom)

		return &TIPCAddr{AddrType: sa.AddrType, Scope: sa.Scope, Service:service, Instance:instance, Domain:domain}
	}
	return nil
}

func (a *TIPCAddr) family() int {
	return syscall.AF_TIPC
}

func (a *TIPCAddr) isWildcard() bool {
	return false
}

// convert net.TIPCAddr to system sockaddr
func (a *TIPCAddr) sockaddr(family int) (syscall.Sockaddr, error) {
	if a == nil {
		return nil, nil
	}
    f := new (syscall.SockaddrTIPC)
    f.AddrType = a.AddrType
    f.Scope =  a.Scope
    sub_ser := make([]byte, 4)
    sub_inst := make([]byte, 4)
    sub_dom := make([]byte, 4)
    binary.LittleEndian.PutUint32(sub_ser, a.Service)
    binary.LittleEndian.PutUint32(sub_inst, a.Instance)
    binary.LittleEndian.PutUint32(sub_dom, a.Domain)
    for index := 0; index < 4; index ++ {
        f.Addr[index] = sub_ser[index] 
        f.Addr[4 + index] = sub_inst[index] 
        f.Addr[8 + index] = sub_dom[index] 
    }
    return f, nil
}

// TIPCConn is an implementation of the Conn interface for TIPC network
// connections.
type TIPCConn struct {
	conn
}

func newTIPCConn(fd *netFD) *TIPCConn {
	c := &TIPCConn{conn{fd}}
	c.SetNoDelay(true)
	return c
}

// ReadFrom implements the io.ReaderFrom ReadFrom method.
func (c *TIPCConn) ReadFrom(r io.Reader) (int64, error) {
	if n, err, handled := sendFile(c.fd, r); handled {
		return n, err
	}
	return genericReadFrom(c, r)
}

// CloseRead shuts down the reading side of the TIPC connection.
// Most callers should just use Close.
func (c *TIPCConn) CloseRead() error {
	if !c.ok() {
		return syscall.EINVAL
	}
	return c.fd.closeRead()
}

// CloseWrite shuts down the writing side of the TIPC connection.
// Most callers should just use Close.
func (c *TIPCConn) CloseWrite() error {
	if !c.ok() {
		return syscall.EINVAL
	}
	return c.fd.closeWrite()
}

// SetLinger sets the behavior of Close on a connection which still
// has data waiting to be sent or to be acknowledged.
//
// If sec < 0 (the default), the operating system finishes sending the
// data in the background.
//
// If sec == 0, the operating system discards any unsent or
// unacknowledged data.
//
// If sec > 0, the data is sent in the background as with sec < 0. On
// some operating systems after sec seconds have elapsed any remaining
// unsent data may be discarded.
func (c *TIPCConn) SetLinger(sec int) error {
	if !c.ok() {
		return syscall.EINVAL
	}
	return setLinger(c.fd, sec)
}

// SetKeepAlive sets whether the operating system should send
// keepalive messages on the connection.
func (c *TIPCConn) SetKeepAlive(keepalive bool) error {
	if !c.ok() {
		return syscall.EINVAL
	}
	return setKeepAlive(c.fd, keepalive)
}

// SetKeepAlivePeriod sets period between keep alives.
func (c *TIPCConn) SetKeepAlivePeriod(d time.Duration) error {
	if !c.ok() {
		return syscall.EINVAL
	}
	return setKeepAlivePeriod(c.fd, d)
}

// SetNoDelay controls whether the operating system should delay
// packet transmission in hopes of sending fewer packets (Nagle's
// algorithm).  The default is true (no delay), meaning that data is
// sent as soon as possible after a Write.
func (c *TIPCConn) SetNoDelay(noDelay bool) error {
	if !c.ok() {
		return syscall.EINVAL
	}
	return setNoDelay(c.fd, noDelay)
}

// DialTIPC connects to the remote address raddr on the network net,
// which must be "tipc".  If laddr is not nil, it is
// used as the local address for the connection.
func DialTIPC(net string, laddr, raddr *TIPCAddr) (*TIPCConn, error) {
	switch net {
	case "tipc":
	default:
		return nil, &OpError{Op: "dial", Net: net, Addr: raddr, Err: UnknownNetworkError(net)}
	}
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: net, Addr: nil, Err: errMissingAddress}
	}
	return dialTIPC(net, laddr, raddr, noDeadline)
}

// How do we use SOCK_SEQPACKET, SOCK_RDM, etc supported by TIPC 
// SANDEEP TIPC - see udpsock_posix.go
// SANDEEP also add support for multicast
// who calls ListenPacket ?
func dialTIPC(net string, laddr, raddr *TIPCAddr, deadline time.Time) (*TIPCConn, error) {
	fd, err := socket(net, syscall.AF_TIPC, syscall.SOCK_STREAM, 0, false, laddr, raddr, deadline)

	if err != nil {
		return nil, &OpError{Op: "dial", Net: net, Addr: raddr, Err: err}
	}
	return newTIPCConn(fd), nil
}

// TIPCListener is a TIPC network listener.  Clients should typically
// use variables of type Listener instead of assuming TIPC.
type TIPCListener struct {
	fd *netFD
}

// AcceptTIPC accepts the next incoming call and returns the new
// connection.
func (l *TIPCListener) AcceptTIPC() (*TIPCConn, error) {
	if l == nil || l.fd == nil {
		return nil, syscall.EINVAL
	}
	fd, err := l.fd.accept()
	if err != nil {
		return nil, err
	}
	return newTIPCConn(fd), nil
}

// Accept implements the Accept method in the Listener interface; it
// waits for the next call and returns a generic Conn.
func (l *TIPCListener) Accept() (c Conn, err error) {
	return l.AcceptTIPC()
}

// Close stops listening on the TIPC address.
// Already Accepted connections are not closed.
func (l *TIPCListener) Close() error {
	if l == nil || l.fd == nil {
		return syscall.EINVAL
	}
	return l.fd.Close()
}

// Addr returns the listener's network address, a *TIPCAddr.
func (l *TIPCListener) Addr() Addr { return l.fd.laddr }

// SetDeadline sets the deadline associated with the listener.
// A zero time value disables the deadline.
func (l *TIPCListener) SetDeadline(t time.Time) error {
	if l == nil || l.fd == nil {
		return syscall.EINVAL
	}
	return l.fd.setDeadline(t)
}

// File returns a copy of the underlying os.File, set to blocking
// mode.  It is the caller's responsibility to close f when finished.
// Closing l does not affect f, and closing f does not affect l.
//
// The returned os.File's file descriptor is different from the
// connection's.  Attempting to change properties of the original
// using this duplicate may or may not have the desired effect.
func (l *TIPCListener) File() (f *os.File, err error) { return l.fd.dup() }

// ListenTIPC announces on the TIPC address laddr and returns a TIPC
// listener.  Net must be "tipc".  If laddr has a
// port of 0, ListenTIPC will choose an available port.  The caller can
// use the Addr method of TIPCListener to retrieve the chosen address.
func ListenTIPC(net string, laddr *TIPCAddr) (*TIPCListener, error) {
	if net != "tipc" {
		return nil, &OpError{Op: "listen", Net: net, Addr: laddr, Err: UnknownNetworkError(net)}
	}
	if laddr == nil {
		return nil, &OpError{Op: "listen", Net: net, Addr: laddr, Err: UnknownNetworkError(net)}
	}

	fd, err := socket(net, syscall.AF_TIPC, syscall.SOCK_STREAM, 0, false,laddr, nil, noDeadline)

	if err != nil {
		return nil, &OpError{Op: "listen", Net: net, Addr: laddr, Err: err}
	}

	return &TIPCListener{fd}, nil
}
