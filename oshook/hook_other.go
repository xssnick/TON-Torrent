//go:build !darwin

package oshook

import "os"

func HookFileStartup(callback func([]byte)) {
	if len(os.Args) > 1 {
		go func() { // to not block main thread
			data, err := os.ReadFile(os.Args[1])
			if err == nil {
				callback(data)
			}
		}()
	}
}
