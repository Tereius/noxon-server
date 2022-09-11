package noxon

import (
	"net"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

const noxonDomain = "noxonserver.eu"
const wifiradiofrontierDomain = "wifiradiofrontier.com"

func getHandleNoxonserver(resolvedIp string) func(w dns.ResponseWriter, r *dns.Msg) {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		if len(r.Question) > 0 {
			m.Answer = append(m.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
				A:   net.ParseIP(resolvedIp),
			})
		}
		w.WriteMsg(m)
	}
}

func getHandleWifiradiofrontier(ntpHostname string) func(w dns.ResponseWriter, r *dns.Msg) {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		if len(r.Question) > 0 {
			ips, err := net.LookupIP(ntpHostname)
			if err != nil {
				log.Errorf("Could not lookup host '%s'", err.Error())
			} else {
				for _, ip := range ips {
					if ipv4 := ip.To4(); ipv4 != nil {
						m.Answer = append(m.Answer, &dns.A{
							Hdr: dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
							A:   ip,
						})
					}
				}
			}
		}
		w.WriteMsg(m)
	}
}

func serve() {

	server := &dns.Server{Addr: "0.0.0.0:53", Net: "udp", TsigSecret: nil, ReusePort: false}
	if err := server.ListenAndServe(); err != nil {
		log.Errorf("Could not start dns server: %s", err.Error())
	}
}

func StartDnsServer(hostIp string, ntpHost string) {

	log.Infof("Starting dns server")
	ip := net.ParseIP(hostIp)
	if ip != nil {
		log.Infof("Registered ip '%s' for domain '%s'", ip.String(), noxonDomain)
		dns.DefaultServeMux.HandleFunc(noxonDomain, getHandleNoxonserver(ip.String()))
	} else {
		log.Errorf("Could not register dns entry for domain '%s': Invalid ip provided", noxonDomain)
	}
	if len(ntpHost) > 0 {
		log.Infof("Registered ntp host '%s' for domain '%s'", ntpHost, wifiradiofrontierDomain)
		dns.DefaultServeMux.HandleFunc(wifiradiofrontierDomain, getHandleWifiradiofrontier(ntpHost))
	}
	go serve()
}
