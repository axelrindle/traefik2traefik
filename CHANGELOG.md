# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Added

- Multiple rule filters
- The Traefik endpoint host name is resolved to the first address that answers an ICMP echo. Addresses that do not answer are skipped.
- Option `traefik.insecureSkipVerify` skips validation of the endpoint's TLS certificate (default `false`)

### Changed

- Renamed and restructured project
- Container runs as uid:gid `1000:1000` by default now

### Fixed

- A request to the Traefik endpoint keeps the original host name in the `Host` header and the TLS server name, even though the connection goes to the resolved address

## 0.1.1

### Fixed

- The HTTP endpoint returns an empty response when no routers are present

## 0.1.0

Initial release 🎉
