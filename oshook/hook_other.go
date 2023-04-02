//go:build !darwin

package oshook

func HookFileStartup(callback func([]byte)) {
}
