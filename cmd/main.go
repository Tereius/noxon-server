package main

import (
	conf "git.privatehive.de/bjoern/noxon-server/internal"
	"git.privatehive.de/bjoern/noxon-server/pkg/noxon"
	log "github.com/sirupsen/logrus"
)

func main() {

	//log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)

	config := conf.ParseConfig()

	if config.DnsConfig.Enabled {
		noxon.StartDnsServer(config.DnsConfig.HostIp, config.DnsConfig.NtpHost)
	}

	serverSettings := noxon.NewDefaultNoxonServerSettings()
	serverSettings = serverSettings.WithWhitelist(config.Whitelist)
	serverSettings = serverSettings.WithBlacklist(config.Blacklist)
	serverSettings = serverSettings.WithStationsModel(noxon.NewJsonStationsModel())
	serverSettings = serverSettings.WithPresetsModel(noxon.NewJsonPresetsModel())

	noxon.NewNoxonServer(serverSettings).StartAndServe()
}
