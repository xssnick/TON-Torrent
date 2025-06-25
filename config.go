package main

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	tunnelConfig "github.com/ton-blockchain/adnl-tunnel/config"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type Config struct {
	Version       uint
	DownloadsPath string
	SeedMode      bool
	ListenAddr    string
	Key           []byte

	IsDarkTheme  bool
	PortsChecked bool

	NetworkConfigPath string
	FetchIPOnStartup  bool

	TunnelConfig *tunnelConfig.ClientConfig

	mx sync.Mutex
}

func LoadConfig(dir string) (*Config, error) {
	var cfg *Config
	path := dir + "/config.json"
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		_, priv, err := ed25519.GenerateKey(nil)
		if err != nil {
			return nil, err
		}

		cfg = &Config{
			Version:       1,
			DownloadsPath: downloadsPath(),
			ListenAddr:    ":13333",
			Key:           priv.Seed(),
		}

		cfg.TunnelConfig, err = tunnelConfig.GenerateClientConfig()
		if err != nil {
			return nil, err
		}
		cfg.TunnelConfig.PaymentsEnabled = true
		cfg.TunnelConfig.Payments.DBPath = dir + "/payments-db"
		cfg.TunnelConfig.Payments.ChannelsConfig.SupportedCoins.Ton.BalanceControl.DepositUpToAmount = "3"
		cfg.TunnelConfig.Payments.ChannelsConfig.SupportedCoins.Ton.BalanceControl.DepositWhenAmountLessThan = "1"

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

		err = json.Unmarshal(data, &cfg)
		if err != nil {
			return nil, err
		}

		if cfg.Key == nil {
			_, priv, err := ed25519.GenerateKey(nil)
			if err != nil {
				return nil, err
			}
			cfg.Key = priv.Seed()
			_ = cfg.SaveConfig(dir)
		}
	}

	var updated bool
	if cfg.Version < 1 {
		cfg.Version = 1
		cfg.TunnelConfig, err = tunnelConfig.GenerateClientConfig()
		if err != nil {
			return nil, err
		}
		cfg.TunnelConfig.Payments.DBPath = dir + "/payments-db"
		cfg.TunnelConfig.Payments.ChannelsConfig.SupportedCoins.Ton.BalanceControl.DepositUpToAmount = "3"
		cfg.TunnelConfig.Payments.ChannelsConfig.SupportedCoins.Ton.BalanceControl.DepositWhenAmountLessThan = "1"
		updated = true
	}

	if cfg.Version < 2 {
		cfg.Version = 2
		cfg.TunnelConfig.PaymentsEnabled = true
		updated = true
	}

	if updated {
		err = cfg.SaveConfig(dir)
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func (cfg *Config) SaveConfig(dir string) error {
	cfg.mx.Lock()
	defer cfg.mx.Unlock()

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

func CheckCanSeed() (string, bool) {
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
