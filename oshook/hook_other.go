//go:build !darwin

package oshook

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/audrenbdb/goforeground"
)

var once sync.Once

func initCrossApp(cbFile func([]byte), cbHash func(string)) {
	mx := http.NewServeMux()
	mx.HandleFunc("/open", func(w http.ResponseWriter, r *http.Request) {
		_ = goforeground.Activate(os.Getpid()) // bring to front
	})
	mx.HandleFunc("/open/meta", func(w http.ResponseWriter, r *http.Request) {
		_ = goforeground.Activate(os.Getpid()) // bring to front
		data, err := io.ReadAll(r.Body)
		if err == nil {
			cbFile(data)
		}
	})
	mx.HandleFunc("/open/hash", func(w http.ResponseWriter, r *http.Request) {
		_ = goforeground.Activate(os.Getpid()) // bring to front
		hash, err := io.ReadAll(r.Body)
		if err == nil {
			cbHash(string(hash))
		}
	})
	_ = http.ListenAndServe("127.0.0.1:33038", mx)
}

func HookStartup(cbFile func([]byte), cbHash func(string)) {
	once.Do(func() {
		client := http.Client{
			Timeout: 500 * time.Millisecond,
		}
		if len(os.Args) > 1 {
			if strings.HasPrefix(os.Args[1], "tonbag://") || strings.HasPrefix(os.Args[1], "tonstorage://") {
				u, err := url.Parse(os.Args[1])
				if err == nil {
					_, err = client.Post("http://127.0.0.1:33038/open/hash",
						"application/octet-stream", bytes.NewBuffer([]byte(u.Host)))
					if err == nil {
						os.Exit(0)
						return
					}

					cbHash(u.Host)
				}
			} else {
				data, err := os.ReadFile(os.Args[1])
				if err == nil {
					_, err = client.Post("http://127.0.0.1:33038/open/meta",
						"application/octet-stream", bytes.NewBuffer(data))
					if err == nil {
						os.Exit(0)
						return
					}

					cbFile(data)
				}
			}
		} else {
			_, err := client.Get("http://127.0.0.1:33038/open")
			if err == nil {
				os.Exit(0)
				return
			}
		}

		// when no running instances, run cross server
		go initCrossApp(cbFile, cbHash)
	})
}
