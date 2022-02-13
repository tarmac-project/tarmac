---
description: Configuring Tarmac with Redis
---

# Redis Configuration

This page contains Redis specific configuration to utilize Redis with Tarmac.

Tarmac supports multiple configuration sources such as Environment Variables, JSON files, or using HashiCorp Consul. All of these configuration options can also exist together to provide both static and dynamic configurations.

When using Environment Variables, all configurations are prefixed with `APP_`. The list below will show both Environment and Consul/JSON format for configuration.

| Environment Variable | Consul/JSON | Type | Description |
| :--- | :--- | :--- | :--- |
| `APP_REDIS_SERVER` | `redis_server` | `string` | Redis server address |
| `APP_REDIS_DATABASE` | `redis_database` | `int` | Redis Database (default: `0`) |
| `APP_REDIS_PASSWORD` | `redis_password` | `string` | Redis password |
| `APP_REDIS_SENTINEL_SERVERS` | `redis_sentinel_servers` | `[]string` | Redis Sentinel Server Addresses |
| `APP_REDIS_SENTINEL_MASTER` | `redis_sentinel_master` | `string` | Redis Sentinel Master Instance Name |
| `APP_REDIS_CONNECT_TIMEOUT` | `redis_connect_timeout` | `int` | Redis Connection Timeout in seconds |
| `APP_REDIS_HOSTNAME_VERIFY` | `redis_hostname_verify` | `bool` | Skip hostname verification for TLS |
| `APP_REDIS_KEEPALIVE` | `redis_keepalive` | `int` | TCP Keepalive Interval in seconds (Default: `300`) |
| `APP_REDIS_MAX_ACTIVE` | `redis_max_active` | `int` | Max Active Connections |
| `APP_REDIS_READ_TIMEOUT` | `redis_read_timeout` | `int` | Read timeout in seconds |
| `APP_REDIS_WRITE_TIMEOUT` | `redis_write_timeout` | `int` | Write timeout in seconds |

## Consul Format

When using Consul the `consul_keys_prefix` should be the path to a key with a JSON `string` as the value. For example, a key of `tarmac/config` will have a value of `{"from_consul":true}`.

![](../.gitbook/assets/consul-example.png)

