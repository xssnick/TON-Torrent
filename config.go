package main

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"github.com/tonutils/torrent-client/core/gostorage"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type Config struct {
	// DaemonPath        string
	DaemonControlAddr string
	DownloadsPath     string
	ListenAddr        string
	GoStorage         gostorage.Config
}

func LoadConfig(dir string) (*Config, error) {
	path := dir + "/config.json"
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		_, priv, err := ed25519.GenerateKey(nil)
		if err != nil {
			return nil, err
		}

		cfg := &Config{
			DaemonControlAddr: "127.0.0.1:15555",
			DownloadsPath:     downloadsPath(),
			ListenAddr:        ":13333",
			GoStorage: gostorage.Config{
				Key:           priv,
				ListenAddr:    ":17777",
				DownloadsPath: downloadsPath(),
			},
		}

		ip, seed := checkCanSeed()
		if seed {
			cfg.ListenAddr = ip + cfg.ListenAddr
			cfg.GoStorage.ExternalIP = ip
		}

		err = cfg.SaveConfig(dir)
		if err != nil {
			return nil, err
		}

		return cfg, nil
	} else if err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var cfg Config
		err = json.Unmarshal(data, &cfg)
		if err != nil {
			return nil, err
		}
		return &cfg, nil
	}

	return nil, err
}

func (cfg *Config) SaveConfig(dir string) error {
	path := dir + "/config.json"

	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0766)
	if err != nil {
		return err
	}
	return nil
}

func downloadsPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		wd, err := os.Getwd()
		if err != nil {
			return "./"
		}
		return wd
	}
	return filepath.Join(homeDir, "Downloads")
}

func checkIPAddress(ip string) string {
	p := net.ParseIP(ip)
	if p == nil {
		println("bad ip", len(p))
		return ""
	}
	p = p.To4()
	if p == nil {
		println("bad ip, not v4", len(p))
		return ""
	}

	println("ip ok", p.String())
	return p.String()
}

func checkCanSeed() (string, bool) {
	ch := make(chan bool, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ip := ""
	go func() {
		defer func() {
			ch <- ip != ""
		}()

		listen, err := net.Listen("tcp", "0.0.0.0:18889")
		if err != nil {
			println("listen err", err.Error())
			return
		}
		defer listen.Close()

		conn, err := listen.Accept()
		if err != nil {
			println("accept err", err.Error())
			return
		}

		ipData := make([]byte, 256)
		n, err := conn.Read(ipData)
		if err != nil {
			println("read err", err.Error())
			return
		}

		ip = string(ipData[:n])
		println("got from server:", "'"+ip+"'")
		ip = checkIPAddress(ip)
		_ = conn.Close()
	}()

	ips, err := net.LookupIP("tonutils.com")
	if err != nil || len(ips) == 0 {
		return "", false
	}

	println("port checker at:", ips[0].String())
	conn, err := net.Dial("tcp", ips[0].String()+":9099")
	if err != nil {
		return "", false
	}

	_, err = conn.Write([]byte("ME"))
	if err != nil {
		return "", false
	}

	ok := false
	select {
	case k := <-ch:
		ok = k
		println("port result:", ok, "public ip:", ip)
	case <-ctx.Done():
		println("port check timeout, will use client mode")
	}

	return ip, ok
}

var CustomRoot = ""

func PrepareRootPath() (string, error) {
	if CustomRoot != "" {
		return CustomRoot, nil
	}

	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		path := home + "/Library/Application Support/org.tonutils.tontorrent"
		_, err = os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(path, 0766)
			}
			if err != nil {
				return "", err
			}
		}
		return path, nil
	case "windows":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		path := home + "\\AppData\\Roaming\\TON Torrent.exe"
		_, err = os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(path, 0766)
			}
			if err != nil {
				return "", err
			}
		}
		return path, nil
	}

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(ex), nil
}
