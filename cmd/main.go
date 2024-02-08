package main

import (
	"os"

	conf "git.privatehive.de/bjoern/noxon-server/internal"
	"git.privatehive.de/bjoern/noxon-server/pkg/noxon"
	log "github.com/sirupsen/logrus"
)

func main() {

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	config := conf.ParseConfig()

	if config.DnsConfig.Enabled {
		noxon.StartDnsServer(config.DnsConfig.HostIp, config.DnsConfig.NtpHost, config.DnsConfig.Domains)
	}

	serverSettings := noxon.NewDefaultNoxonServerSettings()
	serverSettings = serverSettings.WithWhitelist(config.Whitelist)
	serverSettings = serverSettings.WithBlacklist(config.Blacklist)
	serverSettings = serverSettings.WithStationsModel(noxon.NewJsonStationsModel())
	serverSettings = serverSettings.WithPresetsModel(noxon.NewJsonPresetsModel())
	serverSettings = serverSettings.WithLoginEndpoints(config.EndpointConfig.Login)
	serverSettings = serverSettings.WithSearchEndpoints(config.EndpointConfig.Search)
	serverSettings = serverSettings.WithGetPresetsEndpoints(config.EndpointConfig.GetPreset)
	serverSettings = serverSettings.WithAddPresetsEndpoints(config.EndpointConfig.AddPreset)

	noxon.NewNoxonServer(serverSettings).StartAndServe()
}
