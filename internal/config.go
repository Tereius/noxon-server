package config

import (
	"os"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

type DnsConfig struct {
	Enabled bool     `json:"enabled" toml:"enabled"`
	HostIp  string   `json:"hostIp" toml:"hostIp"`
	Domains []string `json:"records" toml:"records"`
	NtpHost string   `json:"ntpHost" toml:"ntpHost"`
}

type EndpointsConfig struct {
	Login     []string `json:"login" toml:"login"`
	Search    []string `json:"search" toml:"search"`
	GetPreset []string `json:"getPreset" toml:"getPreset"`
	AddPreset []string `json:"addPreset" toml:"addPreset"`
}

type Config struct {
	DnsConfig      DnsConfig       `json:"dns" toml:"dns"`
	EndpointConfig EndpointsConfig `json:"endpoints" toml:"endpoints"`
	Whitelist      []string        `json:"whitelist" toml:"whitelist"`
	Blacklist      []string        `json:"blacklist" toml:"blacklist"`
}

func ParseConfig() Config {

	config := Config{
		DnsConfig: DnsConfig{
			Enabled: false,
			HostIp:  "",
			Domains: []string{"noxonserver.eu", "vtuner.com"},
			NtpHost: "de.pool.ntp.org",
		},
		EndpointConfig: EndpointsConfig{
			Login:     []string{"/setupapp/fs/asp/BrowseXML/loginXML.asp", "/setupapp/radio567/asp/BrowseXPA/LoginXML.asp"},
			Search:    []string{"/setupapp/fs/asp/BrowseXML/Search.asp"},
			GetPreset: []string{"/Favorites/GetPreset.aspx"},
			AddPreset: []string{"/Favorites/AddPreset.aspx"},
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

	if len(os.Getenv("ENDPOINTS_LOGIN")) > 0 {
		if runtime.GOOS == "windows" {
			config.EndpointConfig.Login = strings.Split(os.Getenv("ENDPOINTS_LOGIN"), ";")
		} else {
			config.EndpointConfig.Login = strings.Split(os.Getenv("ENDPOINTS_LOGIN"), ":")
		}
	}

	if len(os.Getenv("ENDPOINTS_SEARCH")) > 0 {
		if runtime.GOOS == "windows" {
			config.EndpointConfig.Search = strings.Split(os.Getenv("ENDPOINTS_SEARCH"), ";")
		} else {
			config.EndpointConfig.Search = strings.Split(os.Getenv("ENDPOINTS_SEARCH"), ":")
		}
	}

	if len(os.Getenv("ENDPOINTS_ADD_PRESET")) > 0 {
		if runtime.GOOS == "windows" {
			config.EndpointConfig.AddPreset = strings.Split(os.Getenv("ENDPOINTS_ADD_PRESET"), ";")
		} else {
			config.EndpointConfig.AddPreset = strings.Split(os.Getenv("ENDPOINTS_ADD_PRESET"), ":")
		}
	}

	if len(os.Getenv("ENDPOINTS_GET_PRESET")) > 0 {
		if runtime.GOOS == "windows" {
			config.EndpointConfig.GetPreset = strings.Split(os.Getenv("ENDPOINTS_GET_PRESET"), ";")
		} else {
			config.EndpointConfig.GetPreset = strings.Split(os.Getenv("ENDPOINTS_GET_PRESET"), ":")
		}
	}

	if len(os.Getenv("DNS_ENABLED")) > 0 && strings.ToLower(os.Getenv("DNS_ENABLED")) != "false" {
		config.DnsConfig.Enabled = true
	}

	if len(os.Getenv("DNS_HOST_IP")) > 0 {
		config.DnsConfig.HostIp = os.Getenv("DNS_HOST_IP")
	}

	if len(os.Getenv("DNS_DOMAINS")) > 0 {
		if runtime.GOOS == "windows" {
			config.DnsConfig.Domains = strings.Split(os.Getenv("DNS_DOMAINS"), ";")
		} else {
			config.DnsConfig.Domains = strings.Split(os.Getenv("DNS_DOMAINS"), ":")
		}
	}

	if len(os.Getenv("DNS_NTP_HOST")) > 0 {
		config.DnsConfig.NtpHost = os.Getenv("DNS_NTP_HOST")
	}

	return config
}
