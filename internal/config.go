package config

import (
	"os"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

type DnsConfig struct {
	Enabled bool   `json:"enabled" toml:"enabled"`
	HostIp  string `json:"hostIp" toml:"hostIp"`
	NtpHost string `json:"ntpHost" toml:"ntpHost"`
}

type Config struct {
	DnsConfig DnsConfig `json:"dns" toml:"dns"`
	Whitelist []string  `json:"whitelist" toml:"whitelist"`
	Blacklist []string  `json:"blacklist" toml:"blacklist"`
}

func ParseConfig() Config {

	config := Config{
		DnsConfig: DnsConfig{
			Enabled: false,
			HostIp:  "",
			NtpHost: "de.pool.ntp.org",
		},
		Whitelist: []string{"*"},
		Blacklist: []string{},
	}

	configFile := "config.toml"

	if len(os.Getenv("CONFIG_FILE")) > 0 {
		configFile = os.Getenv("CONFIG_FILE")
	}

	md, err := toml.DecodeFile(configFile, &config)
	if err != nil {
		log.Info("no config file found", err)
	}
	if len(md.Undecoded()) > 0 {
		log.Warn("Config file could not be parsed properly", "error", "some fields could not be decoded", "fields", md.Undecoded())
	}

	if len(os.Getenv("WHITELIST")) > 0 {
		if runtime.GOOS == "windows" {
			config.Whitelist = strings.Split(os.Getenv("WHITELIST"), ";")
		} else {
			config.Whitelist = strings.Split(os.Getenv("WHITELIST"), ":")
		}
	}

	if len(os.Getenv("BLACKLIST")) > 0 {
		if runtime.GOOS == "windows" {
			config.Blacklist = strings.Split(os.Getenv("BLACKLIST"), ";")
		} else {
			config.Blacklist = strings.Split(os.Getenv("BLACKLIST"), ":")
		}
	}

	if len(os.Getenv("DNS_ENABLED")) > 0 && strings.ToLower(os.Getenv("DNS_ENABLED")) != "false" {
		config.DnsConfig.Enabled = true
	}

	if len(os.Getenv("DNS_HOST_IP")) > 0 {
		config.DnsConfig.HostIp = os.Getenv("DNS_HOST_IP")
	}

	if len(os.Getenv("DNS_NTP_HOST")) > 0 {
		config.DnsConfig.NtpHost = os.Getenv("DNS_NTP_HOST")
	}

	return config
}
