//go:build !darwin

package oshook

import (
	"net/url"
	"os"
	"strings"
	"time"
)

func HookStartup(cbFile func([]byte), cbHash func(string), dbg func(s string)) {
	go func() {
		time.Sleep(1 * time.Second)
		dbg(strings.Join(os.Args, " "))
	}()
	if len(os.Args) > 1 {
		go func() { // to not block main thread
			if strings.HasPrefix(os.Args[1], "tonbag://") || strings.HasPrefix(os.Args[1], "tonstorage://") {
				u, err := url.Parse(os.Args[1])
				if err == nil {
					cbHash(u.Host)
				}
			} else {
				data, err := os.ReadFile(os.Args[1])
				if err == nil {
					cbFile(data)
				}
			}
		}()
	}
}
