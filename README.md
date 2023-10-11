# Plex Monitor

[![gitleaks](https://github.com/bwhitehead0/plex_monitor/actions/workflows/gitleaks.yaml/badge.svg)](https://github.com/bwhitehead0/plex_monitor/actions/workflows/gitleaks.yaml) [![govulncheck](https://github.com/bwhitehead0/plex_monitor/actions/workflows/govuln.yaml/badge.svg)](https://github.com/bwhitehead0/plex_monitor/actions/workflows/govuln.yaml) [![Create Release and Assets](https://github.com/bwhitehead0/plex_monitor/actions/workflows/release.yaml/badge.svg)](https://github.com/bwhitehead0/plex_monitor/actions/workflows/release.yaml) [![Create Pre-Release and Assets](https://github.com/bwhitehead0/plex_monitor/actions/workflows/pre-release.yaml/badge.svg)](https://github.com/bwhitehead0/plex_monitor/actions/workflows/pre-release.yaml)
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
#### Endpoints

Plex Monitor responds on two endpoints: `/health` and `/status`.

`/health` returns either a HTTP 200 or 503 response, 200 if Plex is up, 503 if Plex is down.

`/status` returns a JSON payload with `"Status": "0"` for status up, and `"Status:: "1"` for down. For example:

```json
{"RequestDuration":43,"RequestTime":"2023-10-03T06:45:33.198734666Z","Status":"0","Version":"1.32.5.7516"}
```

Or, if the Plex API endpoint does not respond:

```json
{"RequestDuration":23,"RequestTime":"2023-10-03T18:34:15.126480975Z","Status":"1","Version":""}
```

## Installation and Usage

`plex_monitor -config.file=./plex_monitor.yaml`

An example systemd unit file can be found in the `/resources` folder.

Plex Monitor logs to `stderr`, for example:

```
2023/10/04 21:16:52.782045 Plex Monitor v0.1.1 starting up.
2023/10/04 21:16:52.782328 Using configuration file /testing/plex_monitor.yaml
2023/10/04 21:16:52.782661 Using default listen Address 0.0.0.0
2023/10/04 21:16:52.782678 Using default listen port 33131
2023/10/04 21:16:52.782686 Startup time elapsed: 1.196292ms
2023/10/04 21:16:52.782698 IgnoreSSL is set to true
2023/10/04 21:16:56.837337 Received request for endpoint '/status' from 192.168.1.119
2023/10/04 21:16:56.837372 Checking API endpoint https://plex01:32400/identity
2023/10/04 21:16:56.870147 API request completed in 32
2023/10/04 21:16:56.870335 JSON response: {"RequestDuration":32,"RequestTime":"2023-10-05T02:16:56.837384061Z","Status":0,"Version":"1.32.5.7516"}
2023/10/04 21:17:03.251391 Received request for endpoint '/health' from 192.168.1.97
2023/10/04 21:17:03.251424 Checking API endpoint https://plex01:32400/identity
2023/10/04 21:17:03.273992 API request completed in 22
2023/10/04 21:17:03.274022 Returning status 200.
2023/10/04 21:17:04.846245 Received request for endpoint '/status' from 192.168.1.97
2023/10/04 21:17:04.846273 Checking API endpoint https://plex01:32400/identity
2023/10/04 21:17:04.869583 API request completed in 23
2023/10/04 21:17:04.869701 JSON response: {"RequestDuration":23,"RequestTime":"2023-10-05T02:17:04.846284929Z","Status":0,"Version":"1.32.5.7516"}
2023/10/04 21:17:11.356157 Received request for endpoint '/health' from 192.168.1.119
2023/10/04 21:17:11.356187 Checking API endpoint https://plex01:32400/identity
2023/10/04 21:17:11.378328 API request completed in 22
2023/10/04 21:17:11.378357 Returning status 200.
2023/10/04 21:17:30.951365 Received terminated signal. Exiting.
```

### Configuration

The configuration file is in YAML format, and supports the following confuration items:

- **PlexAddress**: *Required.* Your Plex URL
- **PlexPort**: *Required.* Your Plex port
- **IgnoreSSL**: *Required.* Ignore SSL errors, accepts `true` or `false`
- **ListenAddress**: *Optional.* Bind to an IP, default `0.0.0.0`
- **ListenPort**: *Optional.* Bind to port, default `33131`


## Building

Built and tested against Go v1.21.1+

```
git clone https://github.com/bwhitehead0/plex_monitor.git
cd plex_monitor
go build main.go -o plex_monitor plex_monitor
```

# Known Issues

- `plex_monitor` v0.1.1 is not coded to function as a typical Windows service. (See [#16](https://github.com/bwhitehead0/plex_monitor/issues/16))

# Roadmap

- Accept parameters in lieu of config file.
- Find new flag package for easier extensibility and more flexibility.
- Provide more user friendly help output with `-h` flag (dependent on new flag package)
- Docker images