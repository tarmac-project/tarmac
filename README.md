# Go Quick

[![PkgGoDev](https://pkg.go.dev/badge/github.com/madflojo/tarmac)](https://pkg.go.dev/github.com/madflojo/tarmac)
[![Go Report Card](https://goreportcard.com/badge/github.com/madflojo/tarmac)](https://goreportcard.com/report/github.com/madflojo/tarmac)
[![Build Status](https://travis-ci.com/madflojo/tarmac.svg?branch=main)](https://travis-ci.com/madflojo/tarmac)
[![Coverage Status](https://coveralls.io/repos/github/madflojo/tarmac/badge.svg?branch=main)](https://coveralls.io/github/madflojo/tarmac?branch=main)

This project is a boilerplate web application written in Go (Golang).

Starting a new Go Project and wish you had a basic application you could use as a starting point? Want to learn Go but don't know how to structure your project?

The goal of this project is to be all of those things. A clean, straightforward project that offers power users everything they need to start a new Go application. While also providing folks new to Go an example of a well-structured application.

## Features

* Environment Variable and/or HashiCorp Consul-based configuration ([spf13/viper](https://github.com/spf13/viper))
* Modular Key-Value Database integration ([madflojo/hord](https://github.com/madflojo/hord))
* Internal task scheduler for recurring tasks ([madflojo/tasks](https://github.com/madflojo/tasks))
* Live Enable/Disable of debug logging when using Consul for configuration
* Service Resiliency
  - Liveness probe support via `/health` end-point
  - Readiness probe support via `/ready` end-point
  - Graceful shutdown with a SIGTERM signal trap

## Getting Started

The easiest way to get started with this project is to run it locally via Docker Compose. Just follow the instructions below, and you will have an entire local instance with dependencies running.

```console
$ docker compose up tarmac
```

Once running, you can interact with the API via <http://localhost/hello>

### Configuring the Service

This application supports configuring the service from Environment Variables, a JSON file, or using HashiCorp Consul. All of these configuration options can also exist together to provide both static and dynamic configuration.

When using Environment Variables, all configurations are prefixed with `APP_`. The list below will show both Environment and Consul/JSON format for configuration.

| Environment Variable | Consul/JSON | Type | Description |
|----------------------|-------------|------|-------------|
| `APP_ENABLE_TLS` | `enable_tls` | bool | Enable the HTTPS Listener (default: `True`) |
| `APP_LISTEN_ADDR` | `listen_addr` | string | Define the HTTP/HTTPS Listener address (default: `0.0.0.0:8443`) |
| `APP_CONFIG_WATCH_INTERVAL` | `config_watch_interval` | integer | Frequency in seconds which Consul configuration will be refreshed (default: 15) |
| `APP_USE_CONSUL` | `use_consul` | bool | Enable Consul based configuration (default: `False`) |
| `APP_CONSUL_ADDR` | `consul_addr` | string | Consul address (i.e. `consul.example.com:8500`) |
| `APP_CONSUL_KEYS_PREFIX` | `consul_keys_prefix` | string | Key path for app specific consul configuration |
|| `from_consul` | bool | Indicator to reflect whether Consul config was loaded |
| `APP_DEBUG` | `debug` | bool | Enable debug logging |
| `APP_TRACE` | `trace` | bool | Enable trace logging | 
| `APP_DISABLE_LOGGING` | `disable_logging` | bool | Disable all logging |
| `APP_DB_SERVER` | `db_server` | string | Database server address |
| `APP_DB_PASSWORD` | `db_password` | string | Database password | 
| `APP_CERT_FILE` | `cert_file` | string | Certificate File Path (i.e. `/some/path/cert.crt`) |
| `APP_KEY_FILE` | `key_file` | string | Key File Path (i.e. `/some/path/cert.key`) 

#### Consul Format

When using Consul the `consul_keys_prefix` should be the path to a key with a JSON string as the value. For example, a key of `tarmac/config` will have a value of `{"from_consul":true}`.

![](static/img/consul-example.png)
