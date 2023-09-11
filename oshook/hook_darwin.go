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
		ul, err := url.Parse(C.GoString(u))
		if err == nil {
			// to not block main thread
			go cbHash(ul.Host)
		}
	}
}

//export OnLoadFileFromPath
func OnLoadFileFromPath(path *C.char) {
	if cbFile != nil {
		pt := C.GoString(path)
		go func() { // to not block main thread
			data, err := os.ReadFile(pt)
			if err == nil {
				cbFile(data)
			}
		}()
	}
}

func HookStartup(callbackFile func([]byte), callbackHash func(string)) {
	cbFile = callbackFile
	cbHash = callbackHash
	C.HookDelegate()
}
