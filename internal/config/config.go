package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

type NewsServer struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	TLS         bool   `json:"tls"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Connections int    `json:"connections"`
	Priority    int    `json:"priority"`
	Enabled     bool   `json:"enabled"`
}

type Config struct {
	WebDir      string       `json:"-"`
	Listen      string       `json:"listen"`
	ConfigDir   string       `json:"config_dir"`
	DownloadDir string       `json:"download_dir"`
	TempDir     string       `json:"temp_dir"`
	APIKey      string       `json:"api_key"`
	MaxWorkers  int          `json:"max_workers"`
	CleanupTemp bool         `json:"cleanup_temp"`
	PostProcess bool         `json:"post_process"`
	Servers     []NewsServer `json:"servers"`
}

func Load() (*Config, error) {
	configDir := getenv("NZBHARBOR_CONFIG", "/config")
	downloadDir := getenv("NZBHARBOR_DOWNLOADS", "/downloads")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return nil, err
	}

	path := filepath.Join(configDir, "config.json")
	c := &Config{
		Listen: ":6789", ConfigDir: configDir, DownloadDir: downloadDir,
		TempDir: filepath.Join(downloadDir, "incomplete"), MaxWorkers: 8,
		CleanupTemp: true, PostProcess: true,
	}
	b, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(b, c); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	c.ConfigDir = configDir
	c.DownloadDir = downloadDir
	c.WebDir = getenv("NZBHARBOR_WEB", "/opt/nzbharbor/web")
	if c.TempDir == "" {
		c.TempDir = filepath.Join(downloadDir, "incomplete")
	}
	if c.APIKey == "" {
		x := make([]byte, 24)
		_, _ = rand.Read(x)
		c.APIKey = hex.EncodeToString(x)
	}
	if c.MaxWorkers < 1 {
		c.MaxWorkers = 4
	}
	for i := range c.Servers {
		if c.Servers[i].Connections < 1 {
			c.Servers[i].Connections = 4
		}
	}
	sort.SliceStable(c.Servers, func(i, j int) bool { return c.Servers[i].Priority < c.Servers[j].Priority })
	if err := os.MkdirAll(c.TempDir, 0755); err != nil {
		return nil, err
	}
	if err := Save(c); err != nil {
		return nil, err
	}
	return c, nil
}

func Save(c *Config) error {
	b, _ := json.MarshalIndent(c, "", "  ")
	return os.WriteFile(filepath.Join(c.ConfigDir, "config.json"), b, 0600)
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
