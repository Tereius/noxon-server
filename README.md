# noxon-server

Noxon<sup>TM</sup> discontinued the infrastructure-services for legacy iRadio series devices. This is an unofficial server that keeps those legacy devices running. You finally gain full control over your device and avoid e-waste.

Thanks goes to [cweiske](https://github.com/cweiske/noxon-api) for providing the API documentation.

## Quickstart

Just run following command (set for `HOST_IP` the **ipv4 of your machine**)

```bash
$ chmod 666 docker/presets.json && sudo HOST_IP=192.168.0.50 docker-compose -f docker/docker-compose.yaml up
```

Now you have to point the iRadio to your noxon-server. This is done by enabling the **static ip configuration** on the iRadio and providing **your machines ipv4** as the **primary** DNS like so (keep for the secondary DNS the default `0.0.0.0`):

<img src="https://user-images.githubusercontent.com/18425553/189549696-fa4c5c63-8860-4596-b7c8-a403240b97be.png"  width="300">

Now if you browse the radio stations you should see those stations configured in `docker/stations.json`. If you save a preset on the radio it gets written to `docker/presets.json`

## Configuration

The configuration is read form a `config.toml` file. By default the noxon-server expects to find the config file in the cwd but you can overwrite the path by setting the Env. variable `CONFIG_FILE`.

| config.toml key | Env var      | Default         | Meaning                                                                                                                                                                                                                       |
| --------------- | ------------ | --------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| dns.enabled     | DNS_ENABLED  | false           | Enable a DNS server that redirects the radio to this server. If disabled you have to provide your own dns server that returns an A record for the domain `noxonserver.eu` with the ip of the noxon-server                     |
| dns.hostIp      | DNS_HOST_IP  |                 | The ip of the noxon-server (check if http://\<ip\>/health shows "Ok"). Only required if the DNS server is enabled                                                                                                             |
| dns.ntpHost     | DNS_NTP_HOST | de.pool.ntp.org | The host of a ntp server where the radio should get the time from.                                                                                                                                                            |
| Whitelist       | WHITELIST    | \*              | A list of hashed Mac adresses that are allowed to connect to the noxon-server or `*`. For the Env. variable the entries are separated by `;` on windows and `:` on a unix-like os. The Whitelist overrules the Blacklist      |
| Blacklist       | BLACKLIST    |                 | A list of hashed Mac adresses that are blocked from connecting to the noxon-server or `*`. For the Env. variable the entries are separated by `;` on windows and `:` on a unix-like os. The Whitelist overrules the Blacklist |

## Stations list

A flat list of radio stations looks like this

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

A structured station list with nested folders may look like this

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
