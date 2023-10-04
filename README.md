# Plex Monitor

[![gitleaks](https://github.com/bwhitehead0/plex_monitor/actions/workflows/gitleaks.yaml/badge.svg)](https://github.com/bwhitehead0/plex_monitor/actions/workflows/gitleaks.yaml) [![govulncheck](https://github.com/bwhitehead0/plex_monitor/actions/workflows/govuln.yaml/badge.svg)](https://github.com/bwhitehead0/plex_monitor/actions/workflows/govuln.yaml)
<hr>

plex_monitor is a simple status monitor for Plex written in Go.

It polls the Plex API to return the following data:

- **Request Duration** - Time elapsed in milliseconds for polling Plex API endpoint
- **Request Time** - Timestamp for the request from `plex_monitor` to Plex API
- **Status** - Up or Down
- **Version** - Plex version reported by API

This application polls API endpoint `/identity`, for example https://your.plexserver.com:32400/identity

By default, the application listens on `0.0.0.0:33131`.

The Plex server responds with an XML payload such as:

```xml
<MediaContainer size="0" claimed="1" machineIdentifier="ee2e37973bc957d96a81bad551adef994763b651" version="1.32.5.7516-8f4248874"> </MediaContainer>
```

Plex Monitor responds with a JSON payload, for example:

```json
{"RequestDuration":43,"RequestTime":"2023-10-03T06:45:33.198734666Z","Status":"Up","Version":"1.32.5.7516"}
```

Or, if the Plex API endpoint does not respond:

```json
{"RequestDuration":23,"RequestTime":"2023-10-03T18:34:15.126480975Z","Status":"Down","Version":""}
```

## Installation and Usage

`plex_monitor -config.file=./plex_monitor.yaml`

An example systemd unit file can be found in the `/resources` folder.

Plex Monitor logs to `stderr`, for example:

```
2023/10/03 21:15:18.520198 Plex Monitor v0.1.0 starting up.
2023/10/03 21:15:18.520933 Using configuration file /testing/plex_monitor.yaml
2023/10/03 21:15:18.595644 Using default listen Address 0.0.0.0
2023/10/03 21:15:18.595682 Using default listen port 33131
2023/10/03 21:15:18.595699 Startup time elapsed: 75.502074ms
2023/10/03 21:15:39.378319 Received request for endpoint '/status' from 127.0.0.1
2023/10/03 21:15:39.378366 Checking API endpoint https://plex01:32400/identity
2023/10/03 21:15:39.378374 IgnoreSSL is set to true
2023/10/03 21:15:39.409835 JSON response: {"RequestDuration":31,"RequestTime":"2023-10-04T02:15:39.378373798Z","Status":"Up","Version":"1.32.5.7516"}
2023/10/03 21:21:14.126986 Received request for endpoint '/status' from 192.168.1.97
2023/10/03 21:21:14.127028 Checking API endpoint https://plex01:32400/identity
2023/10/03 21:21:14.127036 IgnoreSSL is set to true
2023/10/03 21:21:14.143557 Error connecting to endpoint: Get "https://plex01:32400/identity": dial tcp 192.168.1.119:32400: connect: connection refused
2023/10/03 21:21:14.143627 JSON response: {"RequestDuration":16,"RequestTime":"2023-10-04T02:21:14.127036336Z","Status":"Down","Version":""}
2023/10/03 21:21:38.703434 Received terminated signal. Exiting.
```

### Configuration

The configuration file is in YAML format, and supports the following confuration items:

- **PlexAddress**: *Required.* Your Plex URL
- **PlexPort**: *Required.* Your Plex port
- **IgnoreSSL**: *Required.* Ignore SSL errors, accepts `true` or `false`
- **ListenAddress**: *Optional.* Bind to an IP, default `0.0.0.0`
- **ListenPort**: *Optional.* Bind to port, default `33131`


## Building

Built and tested against Go v1.21.1

```
git clone https://github.com/bwhitehead0/plex_monitor.git
cd plex_monitor
go build main.go -o plex_monitor plex_monitor
```
# Roadmap

- Accept parameters in lieu of config file.
- Find new flag package for easier extensibility and more flexibility.
- Provide more user friendly help output with `-h` flag (dependent on new flag package)