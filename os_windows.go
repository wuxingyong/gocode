// +build windows
package main

import (
	"flag"
	"fmt"
	"syscall"
	"unicode/utf16"
	"unsafe"
	"net"
	"net/rpc"
	"runtime"
)

func CreateSockFlag(name, desc string) *string {
	return flag.String(name, "tcp", desc)
}

// Full path of the current executable
func GetExecutableFileName() string {
	kernel32, _ := syscall.LoadLibrary("kernel32.dll")
	defer syscall.FreeLibrary(kernel32)
	b := make([]uint16, syscall.MAX_PATH)
	getModuleFileName, _ := syscall.GetProcAddress(kernel32, "GetModuleFileNameW")
	ret, _, callErr := syscall.Syscall(uintptr(getModuleFileName),
		3, 0,
		uintptr(unsafe.Pointer(&b[0])),
		uintptr(len(b)))
	if callErr != 0 {
		panic(fmt.Sprintf("GetModuleFileNameW : err %d", int(callErr)))
	}
	return string(utf16.Decode(b[:ret]))
}

func (s *Server) Loop() {
	conn_in := make(chan net.Conn)
	go acceptConnections(conn_in, s.listener)
	for {
		// handle connections or server CMDs (currently one CMD)
		select {
		case c := <-conn_in:
			rpc.ServeConn(c)
			runtime.GC()
		case cmd := <-s.cmd_in:
			switch cmd {
			case SERVER_CLOSE:
				return
			}
		}
	}
}

