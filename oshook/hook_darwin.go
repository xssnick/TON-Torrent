//go:build darwin

package oshook

//#cgo CFLAGS: -x objective-c
//#cgo LDFLAGS: -framework Cocoa
//#include "app_darwin.h"
import "C"
import (
	"os"
	"unsafe"
)

var cb func([]byte)

//export OnLoadFile
func OnLoadFile(data *C.char, length C.uint) {
	if cb != nil && length < 1<<28 {
		dest := make([]byte, length)
		// copy from c to go byte slice
		copy(dest, (*(*[1<<31 - 1]byte)(unsafe.Pointer(data)))[:length:length])
		cb(dest)
	}
}

//export OnLoadFileFromPath
func OnLoadFileFromPath(path *C.char) {
	if cb != nil {
		go func() { // to not block main thread
			data, err := os.ReadFile(C.GoString(path))
			if err == nil {
				cb(data)
			}
		}()
	}
}

func HookFileStartup(callback func([]byte)) {
	cb = callback
	C.HookDelegate()
}
