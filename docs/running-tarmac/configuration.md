---
description: Configuring Tarmac
---

# Configuration

Tarmac supports multiple configuration sources such as Environment Variables, JSON files, or using HashiCorp Consul. All of these configuration options can also exist together to provide both static and dynamic configurations.

When using Environment Variables, all configurations are prefixed with `APP_`. The list below will show both Environment and Consul/JSON format for configuration.

| Environment Variable | Consul/JSON | Type | Description |
| :--- | :--- | :--- | :--- |
| `APP_ENABLE_TLS` | `enable_tls` | bool | Enable the HTTPS Listener \(default: `True`\) |
| `APP_LISTEN_ADDR` | `listen_addr` | string | Define the HTTP/HTTPS Listener address \(default: `0.0.0.0:8443`\) |
| `APP_CONFIG_WATCH_INTERVAL` | `config_watch_interval` | integer | Frequency in seconds which Consul configuration will be refreshed \(default: 15\) |
| `APP_USE_CONSUL` | `use_consul` | bool | Enable Consul based configuration \(default: `False`\) |
| `APP_CONSUL_ADDR` | `consul_addr` | string | Consul address \(i.e. `consul.example.com:8500`\) |
| `APP_CONSUL_KEYS_PREFIX` | `consul_keys_prefix` | string | Key path for app specific consul configuration |
|  | `from_consul` | bool | Indicator to reflect whether Consul config was loaded |
| `APP_DEBUG` | `debug` | bool | Enable debug logging |
| `APP_TRACE` | `trace` | bool | Enable trace logging |
| `APP_DISABLE_LOGGING` | `disable_logging` | bool | Disable all logging |
| `APP_ENABLE_KVSTORE` | `enable_kvstore` | bool | Enable the KV Store |
| `APP_KV_SERVER` | `kv_server` | string | KV Store server address |
| `APP_KV_PASSWORD` | `kv_password` | string | KV Store password |
| `APP_CERT_FILE` | `cert_file` | string | Certificate File Path \(i.e. `/some/path/cert.crt`\) |
| `APP_KEY_FILE` | `key_file` | string | Key File Path \(i.e. `/some/path/cert.key`\) |
| `APP_WASM_FUNCTION` | `wasm_function` | string | Path and Filename of the WASM Function to execute \(Default: `/functions/tarmac.wasm`\) |

## Consul Format

When using Consul the `consul_keys_prefix` should be the path to a key with a JSON string as the value. For example, a key of `tarmac/config` will have a value of `{"from_consul":true}`.

![](../.gitbook/assets/consul-example.png)

