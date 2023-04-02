package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	// DaemonPath        string
	DaemonControlAddr string
	DownloadsPath     string
	ListenAddr        string
}

func LoadConfig(dir string) (*Config, error) {
	path := dir + "/config.json"
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		cfg := &Config{
			// DaemonPath:        exPath,
			DaemonControlAddr: "127.0.0.1:15555",
			DownloadsPath:     downloadsPath(),
			ListenAddr:        ":13333",
		}

		ip, seed := checkCanSeed()
		if seed {
			cfg.ListenAddr = ip + cfg.ListenAddr
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

func getPublicIP() (string, error) {
	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	var ip struct {
		Query string
	}
	err = json.Unmarshal(body, &ip)
	if err != nil {
		return "", err
	}

	return ip.Query, nil
}

func checkCanSeed() (string, bool) {
	ip, err := getPublicIP()
	if err != nil {
		return "", false
	}

	return ip, false
}

func PrepareRootPath() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		_, err = os.Stat(home + "/Library/Application Support/org.tonutils.tontorrent")
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(home+"/Library/Application Support/org.tonutils.tontorrent", 0766)
			}
			if err != nil {
				return "", err
			}
		}
		root = "" + home + "/Library/Application Support/org.tonutils.tontorrent"
	}
	return root, nil
}
