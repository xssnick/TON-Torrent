//go:build darwin

package oshook

//#cgo CFLAGS: -x objective-c
//#cgo LDFLAGS: -framework Cocoa
//#include "app_darwin.h"
import "C"
import (
	"net/url"
	"os"
	"unsafe"
)

var cbFile func([]byte)
var cbHash func(string)

//export OnLoadFile
func OnLoadFile(data *C.char, length C.uint) {
	if cbFile != nil && length < 1<<28 {
		dest := make([]byte, length)
		// copy from c to go byte slice
		copy(dest, (*(*[1<<31 - 1]byte)(unsafe.Pointer(data)))[:length:length])
		cbFile(dest)
	}
}

//export OnLoadURL
func OnLoadURL(u *C.char) {
	if cbHash != nil {
		go func() { // to not block main thread
			ul, err := url.Parse(C.GoString(u))
			if err == nil {
				cbHash(ul.Host)
			}
		}()
	}
}

//export OnLoadFileFromPath
func OnLoadFileFromPath(path *C.char) {
	if cbFile != nil {
		go func() { // to not block main thread
			data, err := os.ReadFile(C.GoString(path))
			if err == nil {
				cbFile(data)
			}
		}()
	}
}

func HookStartup(callbackFile func([]byte), callbackHash func(string), dbg func(s string)) {
	cbFile = callbackFile
	cbHash = callbackHash
	C.HookDelegate()
}
