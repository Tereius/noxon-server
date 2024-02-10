# noxon-server

Noxon<sup>TM</sup> has discontinued infrastructure services for legacy iRadio series devices. This is an unofficial server that keeps these old devices running. You finally get full control over your device and avoid e-waste.

Thanks goes to [cweiske](https://github.com/cweiske/noxon-api) for providing the API documentation.

Tested Devices (see [Server configuration](#server-configuration)):
* NOXON iRadio
* NOXON iRadio 300

## Quickstart (Docker)

Just run following command (set for `HOST_IP` the **ip (v4) of your machine**)

```bash
$ chmod 666 docker/presets.json && sudo HOST_IP=192.168.0.50 docker-compose -f docker/docker-compose.yaml up
```

### iRatio device configuration

Now you have to point the iRadio to your noxon-server. This is done by enabling the **static ip configuration** on the iRadio and providing **your machines ip (v4)** as the **primary** DNS like so (keep for the secondary DNS the default `0.0.0.0`):

<img src="https://user-images.githubusercontent.com/18425553/189549696-fa4c5c63-8860-4596-b7c8-a403240b97be.png"  width="300">

Now if you browse the radio stations you should see those stations configured in `docker/stations.json`. If you save a preset on the radio it gets written to `docker/presets.json`

## Quickstart (compile from source/pre-built binaries)

Compile the code for your desired [os and architecture](https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63) or download the **ready to use binaries** from [here](https://github.com/Tereius/noxon-server/releases):

```bash
$ GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o noxon-server cmd/main.go
```

Then copy the `stations.json` from the docker folder and start the application (set for `DNS_HOST_IP` the **ip (v4) of your machine**)

```bash
$ cp docker/stations.json ./
$ sudo DNS_ENABLED=true DNS_HOST_IP=192.168.0.50 GIN_MODE=release ./noxon-server
```

The configuration of the radio is the same as above

### Raspberry Pi

To compile for Raspberry Pi just use those go Env. vars.: `GOOS=linux GOARCH=arm GOARM=5`

## How it works
*Tested against NOXON iRadio 300, German edition*

The noxon-server launches a minimal DNS server on port 53/udp und a http server on port 80/tcp (make sure your firewall allows traffic to this port). This are privileged ports (the TCP/IP port numbers below 1024) that's why you need admin rights to run this server. There is no way around this because we can't set a specific port for the DNS lookup on the radio device - it always uses the default 53/udp port.

At first the iRadio device contacts the DNS server and asks for an record for the domain `legacy.noxonserver.eu`. The DNS server answers with its own ip (see dns.hostIp in the config file or the environment variable DNS_HOST_IP). If you select "Internetradio" on the display the device asks for the stations via the endpoint `/setupapp/fs/asp/BrowseXML/loginXML.asp` and the noxon-server serves them from the `stations.json` file. If a station is selected for playback on the device a search is first done via `/setupapp/fs/asp/BrowseXML/Search.asp` and the server returns the single station item but with a modified `stationUrl` pointing to the `/playback` endpoint. A reverse proxy then serves the mp3 stream to the device. You heared right, the device does not connect to the original server (as stated in `stations.json`) of the mp3 stream but to the endpoint `/playback` of the noxon-server which acts as a reverse proxy. This decision was made because I could not make it work otherwise - more research is needed. But this also has advantages because you could include e.g. m3u support and more advanced audio codecs which are not supported by iRadio devices.

If a preset button is pressed for 3 seconds on the device a DNS request is made for the domain `gate1.noxonserver.eu` or `gate2.noxonserver.eu`. The DNS server answers with its own ip agin. The radio calls the preset endpoint `/Favorites/AddPreset.aspx` of the noxon-server which creates a `presets.json` file (if not present) and a new entry in the file. If a preset button is pressed briefly the device requests a preset from `/Favorites/GetPreset.aspx` which is served from the `presets.json` file and the playback starts again.

*All the mentioned endpoints and domains are device specific and may need to be changed in [Server configuration](#server-configuration)*

## Server configuration

The configuration is read form a `config.toml` file. By default the noxon-server expects to find the config file in the cwd but you can overwrite the path by setting the Env. variable `CONFIG_FILE`. You can also do without the config file and configure the server only by setting the environment variables.

The different NOXON iRadio devices may expect different endpoints and domains (probably depending on country of marketing, revision and other criteria). You may have to change those endpoints via the configuration (see `endpoints` group and `dns.records`) - use Wireshark to find them.

| config.toml key     | Env. var.            | Default                                                                                    | Meaning                                                                                                                                                                                                                                  |
| ------------------- | -------------------- | ------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| dns.enabled         | DNS_ENABLED          | false                                                                                      | Enable a DNS server that redirects the radio to this server. If disabled you have to provide your own dns server that returns an A record for the [expected domains](#known-endpoints-and-domains) with the ip of the noxon-server       |
| dns.hostIp          | DNS_HOST_IP          |                                                                                            | The ip (v4) of the noxon-server. Only required if the DNS server is enabled                                                                                                                                                                   |
| dns.domains         | DNS_DOMAINS          | [ noxonserver.eu, vtuner.com, terratec.com, my-noxon.com ]                                 | Device [expected domains](#known-endpoints-and-domains) that will be resolved to the ip configured by `dns.hostIp`                                                                                                                                            |
| dns.ntpHost         | DNS_NTP_HOST         | de.pool.ntp.org                                                                            | The host of a ntp server where the radio should get the time from.                                                                                                                                                                       |
| endpoints.login     | ENDPOINTS_LOGIN      | [ /setupapp/fs/asp/BrowseXML/loginXML.asp, /setupapp/radio567/asp/BrowseXPA/LoginXML.asp ] | Device [expected login endpoints](#known-endpoints-and-domains) that get routed to this servers login endpoint                                                                                                                           |
| endpoints.search    | ENDPOINTS_SEARCH     | [ /setupapp/fs/asp/BrowseXML/Search.asp ]                                                  | Device [expected search endpoints](#known-endpoints-and-domains) that get routed to this servers search endpoint                                                                                                                         |
| endpoints.getPreset | ENDPOINTS_GET_PRESET | [ /Favorites/GetPreset.aspx ]                                                              | Device [expected getPreset endpoints](#known-endpoints-and-domains) that get routed to this servers getPreset endpoint                                                                                                                   |
| endpoints.addPreset | ENDPOINTS_ADD_PRESET | [ /Favorites/AddPreset.aspx ]                                                              | Device [expected addPreset endpoints](#known-endpoints-and-domains) that get routed to this servers addPreset endpoint                                                                                                                   |
| Whitelist           | WHITELIST            | \*                                                                                         | A list of hashed Mac adresses that are allowed to connect to the noxon-server or a wildcard `*`. For the Env. variable the entries are separated by `;` on windows and `:` on a unix-like os. The Whitelist overrules the Blacklist      |
| Blacklist           | BLACKLIST            |                                                                                            | A list of hashed Mac adresses that are blocked from connecting to the noxon-server or a wildcard `*`. For the Env. variable the entries are separated by `;` on windows and `:` on a unix-like os. The Whitelist overrules the Blacklist |

## Blacklist/Whitelist

You can blacklist, whitelist iRadio devices. You just have to get the hashed (and salted) Mac address of the iRadio device first. The easiest way to do is to observe the noxon-server logs and look for log messages while the device connects. You should find some log entries that contain the device info e.g.: `device="{b8f629d7e3480b61abdf48c7ba796dae 79 10143 Terratec ger}`. The first 32 character long hex string (here `b8f629d7e3480b61abdf48c7ba796dae`) is the hashed Mac that you are looking for.

## Stations list (stations.json)

You can create the station list according to your wishes but you must **restart noxon-server** if you made changes. Here are two examples:

A flat list of radio stations:

```json
[
  {
    "stationName": "HR info",
    "stationDescription": "HR info",
    "stationUrl": "https://dispatcher.rndfnk.com/hr/hrinfo/live/mp3/high"
  },
  {
    "stationName": "HR 2",
    "stationDescription": "HR 2",
    "stationUrl": "https://dispatcher.rndfnk.com/hr/hr2/live/mp3/high"
  }
]
```

A structured station list with nested folders:

```json
[
  {
    "dirName": "Hessischer Rundfunk",
    "children": [
      {
        "stationName": "HR info",
        "stationDescription": "HR info",
        "stationUrl": "https://dispatcher.rndfnk.com/hr/hrinfo/live/mp3/high"
      },
      {
        "stationName": "HR 2",
        "stationDescription": "HR 2",
        "stationUrl": "https://dispatcher.rndfnk.com/hr/hr2/live/mp3/high"
      }
    ]
  },
  {
    "dirName": "Root folder",
    "children": [
      {
        "dirName": "Nested folder",
        "children": [
          {
            "dirName": "Empty folder",
            "children": []
          }
        ]
      }
    ]
  }
]
```
## Known Endpoints and Domains

Different Noxon iRadio devices expect different endpoints and domains this server has to provide and resolve

### NOXON iRadio 300

| Domains        |
| -------------- |
| noxonserver.eu |

| Endpoint  | Path                                    |
| --------- | --------------------------------------- |
| login     | /setupapp/fs/asp/BrowseXML/loginXML.asp |
| search    | /setupapp/fs/asp/BrowseXML/Search.asp   |
| getPreset | /Favorites/GetPreset.aspx               |
| addPreset | /Favorites/AddPreset.aspx               |


### NOXON iRadio

| Domains      |
| ------------ |
| vtuner.com   |
| terratec.com |
| my-noxon.com |

| Endpoint  | Path                                          |
| --------- | --------------------------------------------- |
| login     | /setupapp/radio567/asp/BrowseXPA/LoginXML.asp |
| search    | /setupapp/fs/asp/BrowseXML/Search.asp         |
| getPreset | /Favorites/GetPreset.aspx                     |
| addPreset | /Favorites/AddPreset.aspx                     |

