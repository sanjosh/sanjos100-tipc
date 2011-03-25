// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"os"
	"syscall"
)

func newFileFD(f *os.File) (*netFD, os.Error) {
	fd, errno := syscall.Dup(f.Fd())
	if errno != 0 {
		return nil, os.NewSyscallError("dup", errno)
	}

	proto, errno := syscall.GetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_TYPE)
	if errno != 0 {
		return nil, os.NewSyscallError("getsockopt", errno)
	}

	toAddr := sockaddrToTCP
	sa, _ := syscall.Getsockname(fd)
	switch sa.(type) {
	default:
		closesocket(fd)
		return nil, os.EINVAL
	case *syscall.SockaddrInet4:
		if proto == syscall.SOCK_DGRAM {
			toAddr = sockaddrToUDP
		} else if proto == syscall.SOCK_RAW {
			toAddr = sockaddrToIP
		}
	case *syscall.SockaddrInet6:
		if proto == syscall.SOCK_DGRAM {
			toAddr = sockaddrToUDP
		} else if proto == syscall.SOCK_RAW {
			toAddr = sockaddrToIP
		}
	case *syscall.SockaddrUnix:
		toAddr = sockaddrToUnix
		if proto == syscall.SOCK_DGRAM {
			toAddr = sockaddrToUnixgram
		} else if proto == syscall.SOCK_SEQPACKET {
			toAddr = sockaddrToUnixpacket
		}
	}
	laddr := toAddr(sa)
	sa, _ = syscall.Getpeername(fd)
	raddr := toAddr(sa)

	return newFD(fd, 0, proto, laddr.Network(), laddr, raddr)
}

// FileConn returns a copy of the network connection corresponding to
// the open file f.  It is the caller's responsibility to close f when
// finished.  Closing c does not affect f, and closing f does not
// affect c.
func FileConn(f *os.File) (c Conn, err os.Error) {
	fd, err := newFileFD(f)
	if err != nil {
		return nil, err
	}
	switch fd.laddr.(type) {
	case *TCPAddr:
		return newTCPConn(fd), nil
	case *UDPAddr:
		return newUDPConn(fd), nil
	case *UnixAddr:
		return newUnixConn(fd), nil
	case *IPAddr:
		return newIPConn(fd), nil
	}
	fd.Close()
	return nil, os.EINVAL
}

// FileListener returns a copy of the network listener corresponding
// to the open file f.  It is the caller's responsibility to close l
// when finished.  Closing c does not affect l, and closing l does not
// affect c.
func FileListener(f *os.File) (l Listener, err os.Error) {
	fd, err := newFileFD(f)
	if err != nil {
		return nil, err
	}
	switch laddr := fd.laddr.(type) {
	case *TCPAddr:
		return &TCPListener{fd}, nil
	case *UnixAddr:
		return &UnixListener{fd, laddr.Name}, nil
	}
	fd.Close()
	return nil, os.EINVAL
}

// FilePacketConn returns a copy of the packet network connection
// corresponding to the open file f.  It is the caller's
// responsibility to close f when finished.  Closing c does not affect
// f, and closing f does not affect c.
func FilePacketConn(f *os.File) (c PacketConn, err os.Error) {
	fd, err := newFileFD(f)
	if err != nil {
		return nil, err
	}
	switch fd.laddr.(type) {
	case *UDPAddr:
		return newUDPConn(fd), nil
	case *UnixAddr:
		return newUnixConn(fd), nil
	}
	fd.Close()
	return nil, os.EINVAL
}
