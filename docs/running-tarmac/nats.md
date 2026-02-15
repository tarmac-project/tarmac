---
description: Configuring Tarmac with NATS
---

# NATS Configuration

This page contains NATS specific configuration to utilize NATS key-value store with Tarmac.

Tarmac supports multiple configuration sources such as Environment Variables, JSON files, or using HashiCorp Consul. All of these configuration options can also exist together to provide both static and dynamic configurations.

When using Environment Variables, all configurations are prefixed with `APP_`. The list below will show both Environment and Consul/JSON format for configuration.

| Environment Variable | Consul/JSON | Type | Description |
| :--- | :--- | :--- | :--- |
| `APP_NATS_URL` | `nats_url` | `string` | NATS server URL (default: `nats://localhost:4222`) |
| `APP_NATS_BUCKET` | `nats_bucket` | `string` | NATS key-value bucket name (default: `tarmac`) |
| `APP_NATS_SERVERS` | `nats_servers` | `[]string` | NATS cluster server URLs for high availability |
| `APP_NATS_SKIP_TLS_VERIFY` | `nats_skip_tls_verify` | `bool` | Skip TLS hostname verification (default: `false`) |

## Consul Format

When using Consul the `consul_keys_prefix` should be the path to a key with a JSON `string` as the value. For example, a key of `tarmac/config` will have a value of `{"from_consul":true}`.

![](../.gitbook/assets/consul-example.png)

